package grpcclient

import (
	"context"
	pb "gkeeper/api/proto"
	"gkeeper/internal/config"
	"log"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client     pb.GKeeperClient
	conn       *grpc.ClientConn
	logger     *zap.Logger
	token      string
	tokenMutex sync.RWMutex
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

func (c *Client) SetToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = token
}

func (c *Client) GetToken() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.token
}

func (c *Client) createContextWithToken(ctx context.Context) context.Context {
	token := c.GetToken()
	if token != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return ctx
}
