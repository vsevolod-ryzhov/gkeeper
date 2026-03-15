package main

import (
	"context"
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/grpcserver"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	log, err := zap.NewDevelopment()
	logger = log

	if err != nil {
		panic(err)
	}
	defer log.Sync()
}

func Run(ctx context.Context) error {
	config.ParseFlags()
	errCh := make(chan error, 1)

	grpcServer := grpcserver.NewServer(
		&grpcserver.ServerConfig{
			AppPort: config.Options.AppPort,
		},
		logger,
	)

	go func() {
		if serverErr := grpcServer.Start(); serverErr != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", serverErr)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down server...")
		grpcServer.Stop()
	case err := <-errCh:
		return err
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := Run(ctx); err != nil {
		logger.Fatal(err.Error())
	}
}
