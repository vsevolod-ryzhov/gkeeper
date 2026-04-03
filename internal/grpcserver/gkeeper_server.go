package grpcserver

import (
	pb "gkeeper/api/proto"
	"gkeeper/internal/filestorage"
	"gkeeper/internal/jwt"
	"gkeeper/internal/storage"

	"go.uber.org/zap"
)

// GKeeperServer implements the gRPC GKeeper service with storage and file storage backends.
type GKeeperServer struct {
	pb.UnimplementedGKeeperServer
	logger      *zap.Logger
	storage     storage.Storage
	fileStorage filestorage.FileStorage
	jwtManager  *jwt.JWTManager
}

// NewGKeeperServer creates a new GKeeperServer with the given dependencies.
func NewGKeeperServer(logger *zap.Logger, storage storage.Storage, fileStorage filestorage.FileStorage, jwtManager *jwt.JWTManager) *GKeeperServer {
	return &GKeeperServer{
		logger:      logger,
		storage:     storage,
		fileStorage: fileStorage,
		jwtManager:  jwtManager,
	}
}
