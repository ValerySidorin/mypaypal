package main

import (
	"log"

	"github.com/ValerySidorin/mypaypal/internal/servicea/config"
	"github.com/ValerySidorin/mypaypal/internal/servicea/web"
	"github.com/satmaelstorm/envviper"
)

func main() {
	cfg := config.Config{}

	v := envviper.NewEnvViper()
	v.SetEnvParamsSimple("SERVICEA")
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	s := web.NewServer(cfg)
	s.Run()
}
