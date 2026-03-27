package main

import (
	"context"
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/grpcserver"
	"gkeeper/internal/storage"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var logger *zap.Logger

func Run(ctx context.Context) error {
	config.ParseFlags()

	dbStorage, storageErr := storage.NewPostgresStorage(config.Options.DatabaseDSN)
	if storageErr != nil {
		return storageErr
	}

	errCh := make(chan error, 1)

	grpcServer := grpcserver.NewServer(
		&grpcserver.ServerConfig{
			AppPort: config.Options.AppPort,
		},
		logger,
	)

	go func() {
		if serverErr := grpcServer.Start(dbStorage); serverErr != nil {
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
	log, err := zap.NewDevelopment()
	logger = log

	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if errRun := Run(ctx); errRun != nil {
		logger.Fatal(errRun.Error())
	}
}
