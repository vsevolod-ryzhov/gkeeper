package grpcserver

import (
	pb "gkeeper/api/proto"
	"gkeeper/internal/jwt"
	"gkeeper/internal/storage"

	"go.uber.org/zap"
)

type GKeeperServer struct {
	pb.UnimplementedGKeeperServer
	logger     *zap.Logger
	storage    *storage.PostgresStorage
	jwtManager *jwt.JWTManager
}

func NewGKeeperServer(logger *zap.Logger, storage *storage.PostgresStorage, jwtManager *jwt.JWTManager) *GKeeperServer {
	return &GKeeperServer{
		logger:     logger,
		storage:    storage,
		jwtManager: jwtManager,
	}
}
