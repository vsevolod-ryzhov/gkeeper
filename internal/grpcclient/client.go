package grpcclient

//go:generate mockery

import (
	"context"
	pb "gkeeper/api/proto"
	"gkeeper/internal/config"
	"gkeeper/internal/crypto"
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
	crypto     *crypto.Crypto
	token      string
	userID     string
	email      string
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

func (c *Client) SetUserID(userID string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.userID = userID
}

func (c *Client) GetUserID() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.userID
}

func (c *Client) SetEmail(email string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.email = email
}

func (c *Client) GetEmail() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.email
}

func (c *Client) SetCrypto(crypto *crypto.Crypto) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.crypto = crypto
}

func (c *Client) GetCrypto() *crypto.Crypto {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.crypto
}

func (c *Client) createContextWithToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}
