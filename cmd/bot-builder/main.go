package main

import (
	"log"

	"github.com/sensority-labs/builder/internal/config"
	"github.com/sensority-labs/builder/internal/service"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Debug {
		log.Default().Printf("Run bot-builder with config: %+v", cfg)
	}

	log.Fatal(service.Run(cfg))
}
