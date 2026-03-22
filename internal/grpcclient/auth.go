package grpcclient

import (
	"context"
	"encoding/base64"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/crypto"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (c *Client) Login(ctx context.Context, email string, password string) error {
	response, reqErr := c.client.Login(ctx, pb.LoginRequest_builder{
		Email:    proto.String(email),
		Password: proto.String(password),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	salt, err := base64.StdEncoding.DecodeString(response.GetSalt())
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	cryptoObj, err := crypto.NewCryptoFromPassword(password, salt)
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	c.SetToken(response.GetToken())
	c.SetEmail(response.GetEmail())
	c.SetUserID(response.GetUserId())
	c.SetCrypto(cryptoObj)

	c.logger.Info("Logged in successfully", zap.String("user_id", c.userID))
	return nil
}

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
