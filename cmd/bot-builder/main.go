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

	log.Fatal(service.Run(cfg))
}
