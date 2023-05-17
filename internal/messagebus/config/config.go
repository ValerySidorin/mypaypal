package config

import "github.com/ValerySidorin/mypaypal/internal/messagebus/kafka"

type Config struct {
	Type  string       `mapstructure:"type"`
	Kafka kafka.Config `mapstructure:"kafka"`
}
