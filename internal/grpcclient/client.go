package grpcclient

import (
	pb "gkeeper/api/proto"
	"gkeeper/internal/config"
	"log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	client pb.GKeeperClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

func NewClient(logger *zap.Logger) *Client {
	tlsCreds, err := generateTLSCreds()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.NewClient(config.Options.AppPort, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		panic(err)
	}

	client := pb.NewGKeeperClient(conn)

	return &Client{
		client: client,
		conn:   conn,
		logger: logger,
	}
}

func (c *Client) Close() {
	c.conn.Close()
}

func generateTLSCreds() (credentials.TransportCredentials, error) {
	certFile := "crt/ca.crt"

	return credentials.NewClientTLSFromFile(certFile, "")
}
