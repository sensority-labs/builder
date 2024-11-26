package main

import (
	"log"

	"github.com/sensority-labs/builder/internal/service"
)

func main() {
	log.Fatal(service.Run())
}
