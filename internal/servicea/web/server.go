package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ValerySidorin/mypaypal/internal/servicea/config"
	"github.com/ValerySidorin/mypaypal/internal/servicea/web/routes"
	"github.com/gofiber/fiber/v2"
)

type WebServer struct {
	App       *fiber.App
	Config    config.Config
	StopCh    chan struct{}
	IsStarted atomic.Bool
}

func NewServer(cfg config.Config) *WebServer {
	s := &WebServer{}
	s.Config = cfg
	s.StopCh = make(chan struct{})
	s.IsStarted.Store(false)

	return s
}

func (s *WebServer) Stop() {
	if s.IsStarted.Load() {
		s.StopCh <- struct{}{}
	}
}

func (s *WebServer) IsServerStarted() bool {
	return s.IsStarted.Load()
}

func (s *WebServer) Serve() {
	s.App = fiber.New()
	routes.AddRoutes(s.App, s.Config)

	go func() {
		s.IsStarted.Store(true)
		if err := s.App.Listen(":" + s.Config.Http.Port); err != nil {
			if err != http.ErrServerClosed {
				fmt.Println("Can't start webserver: " + err.Error())
			} else {
				fmt.Println(err)
			}
		}
	}()

	go func() {
		<-s.StopCh
		log.Println("WebServer stop signal received")
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		done := make(chan struct{})

		go func() {
			err := s.App.Shutdown()
			if err != nil {
				log.Fatal("WebServer shutdown error: " + err.Error())
			}
			done <- struct{}{}
		}()

		select {
		case <-shutDownCtx.Done():
			log.Println("WebServer shutdown forced")
		case <-done:
			log.Println("WebServer shutdown completed")
		}

		cancel()
		close(s.StopCh)
		close(done)
		s.IsStarted.Store(false)
	}()
}

func (s *WebServer) Run() {
	s.Serve()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	stopCh := make(chan struct{})

	go func(ch <-chan os.Signal, st chan<- struct{}) {
		<-ch
		log.Println("Stop signal received")
		s.Stop()
		log.Println("Stop signal sent to webserver")
		for s.IsServerStarted() {
			time.Sleep(time.Microsecond)
		}
		log.Println("Webserver stopped")
		st <- struct{}{}
	}(sigCh, stopCh)

	<-stopCh
}
