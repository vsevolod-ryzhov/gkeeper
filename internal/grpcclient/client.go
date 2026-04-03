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

// Client is a gRPC client for communicating with the GKeeper server.
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

// NewClient creates a new gRPC client connected to the server address specified in config.
func NewClient(logger *zap.Logger) *Client {
	tlsCreds, err := generateTLSCreds()
	if err != nil {
		log.Fatal(err)
	}

	const maxMessageSize = 51 * 1024 * 1024
	conn, err := grpc.NewClient(config.Options.AppPort,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	)
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

// Close closes the underlying gRPC connection.
func (c *Client) Close() {
	c.conn.Close()
}

func generateTLSCreds() (credentials.TransportCredentials, error) {
	certFile := "crt/ca.crt"

	return credentials.NewClientTLSFromFile(certFile, "")
}

// SetToken stores the authentication token in a thread-safe manner.
func (c *Client) SetToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = token
}

// GetToken returns the current authentication token in a thread-safe manner.
func (c *Client) GetToken() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.token
}

// SetUserID stores the authenticated user's ID in a thread-safe manner.
func (c *Client) SetUserID(userID string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.userID = userID
}

// GetUserID returns the authenticated user's ID in a thread-safe manner.
func (c *Client) GetUserID() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.userID
}

// SetEmail stores the authenticated user's email in a thread-safe manner.
func (c *Client) SetEmail(email string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.email = email
}

// GetEmail returns the authenticated user's email in a thread-safe manner.
func (c *Client) GetEmail() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.email
}

// SetCrypto stores the encryption handler used for client-side data encryption.
func (c *Client) SetCrypto(crypto *crypto.Crypto) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.crypto = crypto
}

// GetCrypto returns the encryption handler used for client-side data encryption.
func (c *Client) GetCrypto() *crypto.Crypto {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.crypto
}

func (c *Client) createContextWithToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}
