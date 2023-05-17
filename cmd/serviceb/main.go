package main

import (
	"log"

	"github.com/ValerySidorin/mypaypal/internal/serviceb/config"
	"github.com/ValerySidorin/mypaypal/internal/serviceb/listener"
	"github.com/satmaelstorm/envviper"
)

func main() {
	cfg := config.Config{}

	v := envviper.NewEnvViper()
	v.SetEnvParamsSimple("SERVICEB")
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	l := listener.NewListener(cfg)
	l.Run()
}
