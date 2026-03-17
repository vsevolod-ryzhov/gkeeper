package grpcserver

import (
	"gkeeper/api/proto"
	"gkeeper/internal/storage"

	"go.uber.org/zap"
)

type GKeeperServer struct {
	proto.UnimplementedGKeeperServer
	logger  *zap.Logger
	storage *storage.PostgresStorage
}

func NewGKeeperServer(logger *zap.Logger, storage *storage.PostgresStorage) *GKeeperServer {
	return &GKeeperServer{
		logger:  logger,
		storage: storage,
	}
}
