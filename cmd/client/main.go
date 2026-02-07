package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: add build info flag to retrieve metainfo
var clientVersion = ""
var buildDate = ""

func main() {
	appChan := make(chan os.Signal, 1)
	listenerChan := make(chan struct{}, 1)
	tickerChan := make(chan struct{}, 1)
	isListernerDone := make(chan struct{}, 1)
	isTickerDone := make(chan struct{}, 1)
	streamChan := make(chan struct{}, 1)
	isStreamDone := make(chan struct{}, 1)
	signal.Notify(appChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	g, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,
				Multiplier: 1,
				MaxDelay:   time.Second,
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	awaitAsyncTasks := func() {
		close(listenerChan)
		close(tickerChan)
		close(streamChan)
		<-isListernerDone
		<-isTickerDone
		<-isStreamDone
	}

	defer g.Close()
	app := client.NewClientApp(
		config.ClientConf{},
		g,
		appChan,
		listenerChan,
		tickerChan,
		streamChan,
		isListernerDone,
		isTickerDone,
		isStreamDone,
		buildDate,
		clientVersion,
	)

	app.Init()
	errGroup, ctx := errgroup.WithContext(context.Background())
	errGroup.Go(func() error {
		return app.Start()
	})

	errGroup.Go(func() error {
		select {
		case <-appChan:
			app.Close()
			awaitAsyncTasks()
			return nil
		case <-ctx.Done():
			awaitAsyncTasks()
			return ctx.Err()
		}
	})

	if err := errGroup.Wait(); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
