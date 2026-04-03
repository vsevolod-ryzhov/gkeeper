package main

import (
	"context"
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/filestorage"
	"gkeeper/internal/grpcserver"
	"gkeeper/internal/storage"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var logger *zap.Logger

// Run initializes the server dependencies and starts the gRPC server, blocking until the context is cancelled.
func Run(ctx context.Context) error {
	config.ParseFlags()

	dbStorage, storageErr := storage.NewPostgresStorage(config.Options.DatabaseDSN)
	if storageErr != nil {
		return storageErr
	}

	fileStore, err := filestorage.NewMinioStorage(
		config.Options.MinioEndpoint,
		config.Options.MinioAccessKey,
		config.Options.MinioSecretKey,
		config.Options.MinioBucket,
		config.Options.MinioUseSSL,
	)
	if err != nil {
		return fmt.Errorf("failed to init minio: %w", err)
	}

	errCh := make(chan error, 1)

	grpcServer := grpcserver.NewServer(
		&grpcserver.ServerConfig{
			AppPort:  config.Options.AppPort,
			CertFile: config.Options.CertFile,
			KeyFile:  config.Options.KeyFile,
		},
		logger,
	)

	go func() {
		if serverErr := grpcServer.Start(dbStorage, fileStore); serverErr != nil {
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
