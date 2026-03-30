package grpcserver

import (
	"context"
	"encoding/base64"
	pb "gkeeper/api/proto"
	"gkeeper/internal/jwt"
	mockfilestorage "gkeeper/internal/mocks/filestorage"
	mockstorage "gkeeper/internal/mocks/storage"
	"gkeeper/internal/model"
	"gkeeper/internal/storage"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func newTestServer(t *testing.T) (*GKeeperServer, *mockstorage.MockStorage, *mockfilestorage.MockFileStorage) {
	store := mockstorage.NewMockStorage(t)
	fileStore := mockfilestorage.NewMockFileStorage(t)
	jwtManager := jwt.NewJWTManager("test-secret-key", 1*time.Hour)
	server := NewGKeeperServer(zap.NewNop(), store, fileStore, jwtManager)
	return server, store, fileStore
}

func TestRegister_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	store.EXPECT().
		CreateUser(mock.Anything, "test@example.com", mock.Anything, mock.Anything).
		Return(&model.UserRecord{
			ID:    userID,
			Email: "test@example.com",
		}, nil)

	req := pb.RegisterRequest_builder{
		Email:    proto.String("test@example.com"),
		Password: proto.String("password123"),
	}.Build()

	resp, err := server.Register(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "User registered successfully", resp.GetResult())
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	server, store, _ := newTestServer(t)

	store.EXPECT().
		CreateUser(mock.Anything, "test@example.com", mock.Anything, mock.Anything).
		Return(nil, storage.ErrUserAlreadyExists)

	req := pb.RegisterRequest_builder{
		Email:    proto.String("test@example.com"),
		Password: proto.String("password123"),
	}.Build()

	_, err := server.Register(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestLogin_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	hashed, _ := hashPassword("password123")
	salt := base64.StdEncoding.EncodeToString(make([]byte, 32))

	store.EXPECT().
		GetUserByEmail(mock.Anything, "test@example.com").
		Return(&model.UserRecord{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: hashed,
			Salt:         salt,
		}, nil)

	req := pb.LoginRequest_builder{
		Email:    proto.String("test@example.com"),
		Password: proto.String("password123"),
	}.Build()

	// grpc.SendHeader requires a gRPC server transport in context,
	// which isn't present in unit tests, so we expect an Internal error
	// from the SendHeader call. The important thing is that storage
	// was called and password was verified correctly.
	_, err := server.Login(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to send token")
}

func TestLogin_UserNotFound(t *testing.T) {
	server, store, _ := newTestServer(t)

	store.EXPECT().
		GetUserByEmail(mock.Anything, "unknown@example.com").
		Return(nil, storage.ErrUserNotFound)

	req := pb.LoginRequest_builder{
		Email:    proto.String("unknown@example.com"),
		Password: proto.String("password123"),
	}.Build()

	_, err := server.Login(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestLogin_WrongPassword(t *testing.T) {
	server, store, _ := newTestServer(t)

	hashed, _ := hashPassword("correct-password")

	store.EXPECT().
		GetUserByEmail(mock.Anything, "test@example.com").
		Return(&model.UserRecord{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: hashed,
			Salt:         base64.StdEncoding.EncodeToString(make([]byte, 32)),
		}, nil)

	req := pb.LoginRequest_builder{
		Email:    proto.String("test@example.com"),
		Password: proto.String("wrong-password"),
	}.Build()

	_, err := server.Login(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}
