package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ValerySidorin/mypaypal/internal/dto"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Topic    string `mapstructure:"topic"`
}

type Publisher struct {
	writer *kafka.Writer
}

type Subscriber struct {
	reader *kafka.Reader
}

func NewPublisher(cfg Config) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Endpoint),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func NewSubscriber(cfg Config) *Subscriber {
	return &Subscriber{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.Endpoint},
			Topic:   cfg.Topic,
		}),
	}
}

func (k *Publisher) Publish(ctx context.Context, r dto.BalanceRequest) error {
	msg := kafka.Message{}

	reqStr, err := json.Marshal(r)
	if err != nil {
		return err
	}

	msg.Value = reqStr
	return k.writer.WriteMessages(ctx, msg)
}

func (k *Subscriber) Subscribe(ctx context.Context, f func(r dto.BalanceRequest) error) {
	for {
		select {
		case <-time.After(1 * time.Millisecond):
			msg, err := k.reader.ReadMessage(ctx)
			if err != nil {
				log.Fatalf("can't read message from kafka: %s", err.Error())
			}

			var req dto.BalanceRequest
			if err := json.Unmarshal(msg.Value, &req); err != nil {
				log.Printf("failed to unmarshal kafka message: %s\n", err.Error())
				continue
			}

			if err := f(req); err != nil {
				log.Printf("failed to process kafka message: %s\n", err.Error())
			}
		case <-ctx.Done():
			if err := k.reader.Close(); err != nil {
				log.Printf("failed to close kafka reader: %s\n", err.Error())
			}
			return
		}

	}
}
