package config

import (
	bus_config "github.com/ValerySidorin/mypaypal/internal/messagebus/config"
	store_config "github.com/ValerySidorin/mypaypal/internal/serviceb/storage/config"
)

type Config struct {
	Storage store_config.Config `mapstructure:"storage"`
	Queue   bus_config.Config   `mapstructure:"queue"`
}
