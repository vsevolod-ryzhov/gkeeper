package grpcserver

import (
	"fmt"
	"gkeeper/internal/storage"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

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

func (s *Server) Start(storage *storage.PostgresStorage) error {
	listen, err := net.Listen("tcp", s.config.AppPort)
	if err != nil {
		s.logger.Error("failed to listen", zap.Error(err))

		return fmt.Errorf("gRPC listener init error: %w", err)
	}

	tlsCreds, err := generateTLSCreds()
	if err != nil {
		s.logger.Fatal("failed to generate tls creds: %v", zap.Error(err))
	}

	s.logger.Info("gRPC server listening", zap.String("port", s.config.AppPort))
	s.grpcServer = grpc.NewServer(grpc.Creds(tlsCreds))
	s.logger.Info("gRPC server started", zap.String("port", s.config.AppPort))

	gkeeperServer := NewGKeeperServer(s.logger, storage)
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

func generateTLSCreds() (credentials.TransportCredentials, error) {
	certFile := "crt/server.crt"
	keyFile := "crt/server.key"

	return credentials.NewServerTLSFromFile(certFile, keyFile)
}
