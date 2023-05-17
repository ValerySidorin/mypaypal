package config

import "github.com/ValerySidorin/mypaypal/internal/messagebus/config"

type Config struct {
	Http  HTTP          `mapstructure:"http"`
	Queue config.Config `mapstructure:"queue"`
}

type HTTP struct {
	Port string `mapstructure:"port"`
}
