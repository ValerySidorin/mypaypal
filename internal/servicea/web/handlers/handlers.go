package handlers

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ValerySidorin/mypaypal/internal/dto"
	"github.com/ValerySidorin/mypaypal/internal/messagebus/config"
	"github.com/ValerySidorin/mypaypal/internal/messagebus/kafka"
	"github.com/gofiber/fiber/v2"
)

type Publisher interface {
	Publish(ctx context.Context, r dto.BalanceRequest) error
}

func NewPublisher(cfg config.Config) (Publisher, error) {
	switch cfg.Type {
	case "kafka":
		return kafka.NewPublisher(cfg.Kafka), nil
	default:
		return nil, errors.New("invalid publisher")
	}
}

type Handlers struct {
	pub Publisher
}

func New(cfg config.Config) (*Handlers, error) {
	p, err := NewPublisher(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "init handlers:")
	}

	return &Handlers{
		pub: p,
	}, nil
}

func (h *Handlers) BalanceHandler(c *fiber.Ctx) error {
	var req dto.BalanceRequest
	if err := c.BodyParser(&req); err != nil {
		c.Write([]byte(err.Error()))
		return fiber.ErrBadRequest
	}

	c.Append("x-user-id", req.UserID)

	if err := h.pub.Publish(c.Context(), req); err != nil {
		c.Write([]byte(err.Error()))
		return fiber.ErrInternalServerError
	}

	return c.SendStatus(http.StatusAccepted)
}
