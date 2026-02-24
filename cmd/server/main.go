package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/app"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srvCh := make(chan os.Signal, 1)
	isSrvDone := make(chan struct{}, 1)
	signal.Notify(srvCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	appServer := app.New(conf, isSrvDone)
	errG := new(errgroup.Group)
	errG.Go(func() error {
		return appServer.Start()
	})

	errG.Go(func() error {
		<-srvCh
		appServer.Shutdown()
		<-isSrvDone

		return nil
	})

	if err := errG.Wait(); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
