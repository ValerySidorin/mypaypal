package listener

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sync/atomic"

	"github.com/ValerySidorin/mypaypal/internal/dto"
	"github.com/ValerySidorin/mypaypal/internal/messagebus/kafka"
	"github.com/ValerySidorin/mypaypal/internal/serviceb/config"
	"github.com/ValerySidorin/mypaypal/internal/serviceb/storage/pg"
)

type Subscriber interface {
	Subscribe(ctx context.Context, f func(r dto.BalanceRequest) error)
}

type Storage interface {
	ApplyTransaction(ctx context.Context, r dto.BalanceRequest) error
}

func NewSubscriber(cfg config.Config) (Subscriber, error) {
	switch cfg.Queue.Type {
	case "kafka":
		return kafka.NewSubscriber(cfg.Queue.Kafka), nil
	default:
		return nil, errors.New("invalid subscriber")
	}
}

func NewStorage(ctx context.Context, cfg config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case "pg":
		return pg.NewStorage(ctx, cfg.Storage.Pg)
	default:
		return nil, errors.New("invalid storage")
	}
}

type Listener struct {
	ctx    context.Context
	cancel func()

	sub       Subscriber
	storage   Storage
	stopCh    chan struct{}
	IsStarted atomic.Bool
}

func NewListener(cfg config.Config) *Listener {
	shutDownCtx, cancel := context.WithCancel(context.Background())

	sub, err := NewSubscriber(cfg)
	if err != nil {
		log.Fatalf("failed to create sub: %s", err.Error())
	}

	storage, err := NewStorage(shutDownCtx, cfg)

	l := &Listener{
		ctx:     shutDownCtx,
		cancel:  cancel,
		sub:     sub,
		storage: storage,
		stopCh:  make(chan struct{}),
	}
	l.IsStarted.Store(false)

	return l
}

func (l *Listener) Stop() {
	if l.IsStarted.Load() {
		l.stopCh <- struct{}{}
	}
}

func (l *Listener) Listen() {
	go func() {
		l.sub.Subscribe(l.ctx, func(r dto.BalanceRequest) error {
			log.Println(r)
			return l.storage.ApplyTransaction(l.ctx, r)
		})
	}()

	go func() {
		<-l.stopCh
		l.cancel()
	}()
}

func (l *Listener) Run() {
	l.Listen()
	log.Println("Listener started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	stopCh := make(chan struct{})

	go func(ch <-chan os.Signal, st chan<- struct{}) {
		<-ch
		log.Println("Stop signal received")
		l.Stop()
		log.Println("Stop signal sent to listener")
		for l.IsStarted.Load() {
			time.Sleep(time.Microsecond)
		}
		log.Println("Listener stopped")
		st <- struct{}{}
	}(sigCh, stopCh)

	<-stopCh
}
