package main

import (
	"context"
	"gkeeper/internal/config"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	log, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}
	defer log.Sync()
}

func Run(ctx context.Context) error {
	config.ParseFlags()

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := Run(ctx); err != nil {
		logger.Fatal(err.Error())
	}
}
