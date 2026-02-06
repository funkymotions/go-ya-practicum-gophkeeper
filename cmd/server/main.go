package main

import (
	"log"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/app"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// TODO: now grpc server blocks main goroutine, in case of graceful shutdown
	// need to run it in a separate goroutine
	appServer := app.New(conf)
	appServer.Start()
}
