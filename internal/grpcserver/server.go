package grpcserver

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "gkeeper/api/proto"
)

type ServerConfig struct {
	AppPort string
}

type Server struct {
	grpcServer *grpc.Server
	config     *ServerConfig
	logger     *zap.Logger
}

func NewServer(config *ServerConfig, logger *zap.Logger) *Server {
	return &Server{
		config: config,
		logger: logger,
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", s.config.AppPort)
	if err != nil {
		s.logger.Error("failed to listen", zap.Error(err))

		return fmt.Errorf("gRPC listener init error: %w", err)
	}

	s.logger.Info("gRPC server listening", zap.String("port", s.config.AppPort))
	s.grpcServer = grpc.NewServer()
	s.logger.Info("gRPC server started", zap.String("port", s.config.AppPort))

	gkeeperServer := NewGKeeperServer(s.logger)
	pb.RegisterGKeeperServer(s.grpcServer, gkeeperServer)

	if serveErr := s.grpcServer.Serve(listen); serveErr != nil {
		return fmt.Errorf("gRPC server failed: %w", serveErr)
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.logger.Info("gRPC server stopped")
	}
}
