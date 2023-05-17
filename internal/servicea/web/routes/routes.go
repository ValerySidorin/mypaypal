package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/ValerySidorin/mypaypal/internal/servicea/config"
	"github.com/ValerySidorin/mypaypal/internal/servicea/web/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func AddRoutes(app *fiber.App, cfg config.Config) {
	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("x-user-id")
		},
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	hs, err := handlers.New(cfg.Queue)
	if err != nil {
		log.Fatalf("error adding routes: %s", err.Error())
	}

	app.Post("/balance", hs.BalanceHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
}
