package grpcserver

import (
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/filestorage"
	"gkeeper/internal/jwt"
	"gkeeper/internal/storage"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "gkeeper/api/proto"
)

// ServerConfig holds the network and TLS configuration for the gRPC server.
type ServerConfig struct {
	AppPort  string
	CertFile string
	KeyFile  string
}

// Server wraps a gRPC server with configuration and logging.
type Server struct {
	grpcServer *grpc.Server
	config     *ServerConfig
	logger     *zap.Logger
}

// NewServer creates a new Server with the given configuration and logger.
func NewServer(config *ServerConfig, logger *zap.Logger) *Server {
	return &Server{
		config: config,
		logger: logger,
	}
}

const maxMessageSize = 51 * 1024 * 1024 // 51MB to accommodate 50MB payloads + overhead

// Start initializes TLS, registers services, and begins serving gRPC requests.
func (s *Server) Start(storage storage.Storage, fileStorage filestorage.FileStorage) error {
	listen, err := net.Listen("tcp", s.config.AppPort)
	if err != nil {
		s.logger.Error("failed to listen", zap.Error(err))

		return fmt.Errorf("gRPC listener init error: %w", err)
	}

	tlsCreds, err := generateTLSCreds(s.config.CertFile, s.config.KeyFile)
	if err != nil {
		s.logger.Error("failed to generate tls creds", zap.Error(err))
		return err
	}

	jwtManager := jwt.NewJWTManager(config.Options.JWTSecretKey, 24*time.Hour)
	authInterceptor := NewAuthInterceptor(jwtManager)

	s.logger.Info("gRPC server listening", zap.String("port", s.config.AppPort))
	s.grpcServer = grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.MaxRecvMsgSize(maxMessageSize),
		grpc.MaxSendMsgSize(maxMessageSize),
	)
	s.logger.Info("gRPC server started", zap.String("port", s.config.AppPort))

	gkeeperServer := NewGKeeperServer(s.logger, storage, fileStorage, jwtManager)
	pb.RegisterGKeeperServer(s.grpcServer, gkeeperServer)

	if serveErr := s.grpcServer.Serve(listen); serveErr != nil {
		return fmt.Errorf("gRPC server failed: %w", serveErr)
	}

	return nil
}

// Stop gracefully shuts down the gRPC server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.logger.Info("gRPC server stopped")
	}
}

func generateTLSCreds(certFile, keyFile string) (credentials.TransportCredentials, error) {
	return credentials.NewServerTLSFromFile(certFile, keyFile)
}
