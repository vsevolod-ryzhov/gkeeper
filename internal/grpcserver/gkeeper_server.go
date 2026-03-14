package grpcserver

import (
	"gkeeper/api/proto"

	"go.uber.org/zap"
)

type GKeeperServer struct {
	proto.UnimplementedGKeeperServer
	logger *zap.Logger
}

func NewGKeeperServer(logger *zap.Logger) *GKeeperServer {
	return &GKeeperServer{
		logger: logger,
	}
}
