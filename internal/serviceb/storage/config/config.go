package config

import "github.com/ValerySidorin/mypaypal/internal/serviceb/storage/pg"

type Config struct {
	Type string    `mapstructure:"type"`
	Pg   pg.Config `mapstructure:"pg"`
}
