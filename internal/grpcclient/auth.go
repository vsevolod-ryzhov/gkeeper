package grpcclient

import (
	"context"
	"encoding/base64"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/crypto"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Login authenticates a user with the given email and password, stores the returned
// token, user ID, email, and initializes the encryption handler from the user's password and salt.
func (c *Client) Login(ctx context.Context, email string, password string) error {
	var header metadata.MD
	response, reqErr := c.client.Login(ctx, pb.LoginRequest_builder{
		Email:    proto.String(email),
		Password: proto.String(password),
	}.Build(), grpc.Header(&header))

	if reqErr != nil {
		return reqErr
	}

	token, err := extractTokenFromHeader(header)
	if err != nil {
		return err
	}

	salt, err := base64.StdEncoding.DecodeString(response.GetSalt())
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	cryptoObj, err := crypto.NewCryptoFromPassword(password, salt)
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	c.SetToken(token)
	c.SetEmail(response.GetEmail())
	c.SetUserID(response.GetUserId())
	c.SetCrypto(cryptoObj)

	c.logger.Info("Logged in successfully", zap.String("user_id", c.userID))
	return nil
}

func extractTokenFromHeader(header metadata.MD) (string, error) {
	values := header.Get("authorization")
	if len(values) == 0 {
		return "", fmt.Errorf("authorization token not found in response metadata")
	}
	token := values[0]
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	return token, nil
}

// Register creates a new user account with the given email and password on the server.
func (c *Client) Register(ctx context.Context, email string, password string) error {
	_, reqErr := c.client.Register(ctx, pb.RegisterRequest_builder{
		Email:    proto.String(email),
		Password: proto.String(password),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	c.logger.Info("Registered successfully", zap.String("user_id", c.userID))
	return nil
}
