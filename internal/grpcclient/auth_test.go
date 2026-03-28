package grpcclient

import (
	"context"
	"encoding/base64"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/crypto"
	mockproto "gkeeper/internal/mocks/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestExtractTokenFromHeader_WithBearer(t *testing.T) {
	header := metadata.Pairs("authorization", "Bearer my-token-123")

	token, err := extractTokenFromHeader(header)
	require.NoError(t, err)
	assert.Equal(t, "my-token-123", token)
}

func TestExtractTokenFromHeader_WithoutBearer(t *testing.T) {
	header := metadata.Pairs("authorization", "raw-token")

	token, err := extractTokenFromHeader(header)
	require.NoError(t, err)
	assert.Equal(t, "raw-token", token)
}

func TestExtractTokenFromHeader_Missing(t *testing.T) {
	header := metadata.Pairs()

	_, err := extractTokenFromHeader(header)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization token not found")
}

func TestLogin_Success(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	salt := make([]byte, 32)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	m.EXPECT().
		Login(mock.Anything, mock.MatchedBy(func(req *pb.LoginRequest) bool {
			return req.GetEmail() == "test@example.com" && req.GetPassword() == "password123"
		}), mock.Anything).
		Run(func(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) {
			// Simulate server sending token via response header
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = metadata.Pairs("authorization", "Bearer test-jwt-token")
				}
			}
		}).
		Return(pb.LoginResponse_builder{
			Salt:   proto.String(encodedSalt),
			Email:  proto.String("test@example.com"),
			UserId: proto.String("user-id-123"),
		}.Build(), nil)

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, "test-jwt-token", c.GetToken())
	assert.Equal(t, "test@example.com", c.GetEmail())
	assert.Equal(t, "user-id-123", c.GetUserID())
	assert.NotNil(t, c.GetCrypto())
}

func TestLogin_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		Login(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("connection refused"))

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Login(context.Background(), "test@example.com", "password123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestLogin_MissingTokenInHeader(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	salt := make([]byte, 32)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	m.EXPECT().
		Login(mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) {
			// Don't set any header — simulates missing token
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = metadata.Pairs()
				}
			}
		}).
		Return(pb.LoginResponse_builder{
			Salt:   proto.String(encodedSalt),
			Email:  proto.String("test@example.com"),
			UserId: proto.String("user-id-123"),
		}.Build(), nil)

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Login(context.Background(), "test@example.com", "password123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization token not found")
}

func TestLogin_InvalidSalt(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	m.EXPECT().
		Login(mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) {
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = metadata.Pairs("authorization", "Bearer token")
				}
			}
		}).
		Return(pb.LoginResponse_builder{
			Salt:   proto.String("not-valid-base64!!!"),
			Email:  proto.String("test@example.com"),
			UserId: proto.String("user-id-123"),
		}.Build(), nil)

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Login(context.Background(), "test@example.com", "password123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode salt")
}

func TestRegister_Success(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		Register(mock.Anything, mock.MatchedBy(func(req *pb.RegisterRequest) bool {
			return req.GetEmail() == "new@example.com" && req.GetPassword() == "securepass"
		}), mock.Anything).
		Return(pb.RegisterResponse_builder{
			Result: proto.String("User registered successfully"),
		}.Build(), nil)

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Register(context.Background(), "new@example.com", "securepass")
	require.NoError(t, err)
}

func TestRegister_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		Register(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("user already exists"))

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Register(context.Background(), "existing@example.com", "password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}

func TestLogin_SetsCryptoCorrectly(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	salt := make([]byte, 32)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	m.EXPECT().
		Login(mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) {
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = metadata.Pairs("authorization", "Bearer token")
				}
			}
		}).
		Return(pb.LoginResponse_builder{
			Salt:   proto.String(encodedSalt),
			Email:  proto.String("test@example.com"),
			UserId: proto.String("user-id"),
		}.Build(), nil)

	c := &Client{
		client: m,
		logger: zap.NewNop(),
	}

	err := c.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)

	// Verify crypto works by encrypting and decrypting
	expectedCrypto, _ := crypto.NewCryptoFromPassword("password123", salt)
	plaintext := []byte("test data")

	encrypted, err := c.GetCrypto().Encrypt(plaintext)
	require.NoError(t, err)

	decrypted, err := expectedCrypto.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
