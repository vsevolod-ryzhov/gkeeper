package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func ctxWithUserID(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), ctxKeyUserID, userID)
}

func TestCreateSecret_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	store.EXPECT().
		CreateSecret(mock.Anything, userID.String(), "my secret", "credentials", "encrypted", "{}", "").
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			Title:     "my secret",
			Type:      "credentials",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	secretType := pb.SecretType_SECRET_TYPE_CREDENTIALS
	req := pb.CreateSecretRequest_builder{
		Title:         proto.String("my secret"),
		Type:          &secretType,
		EncryptedData: []byte("encrypted"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String(""),
	}.Build()

	resp, err := server.CreateSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.Equal(t, secretID.String(), resp.GetId())
}

func TestCreateSecret_Unauthenticated(t *testing.T) {
	server, _, _ := newTestServer(t)

	secretType := pb.SecretType_SECRET_TYPE_TEXT
	req := pb.CreateSecretRequest_builder{
		Title: proto.String("test"),
		Type:  &secretType,
	}.Build()

	_, err := server.CreateSecret(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestCreateSecret_StorageError(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		CreateSecret(mock.Anything, userID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("db connection lost"))

	secretType := pb.SecretType_SECRET_TYPE_TEXT
	req := pb.CreateSecretRequest_builder{
		Title:         proto.String("test"),
		Type:          &secretType,
		EncryptedData: []byte("data"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String(""),
	}.Build()

	_, err := server.CreateSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestUpdateSecret_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	store.EXPECT().
		UpdateSecret(mock.Anything, userID.String(), secretID.String(), "updated", "encrypted", "{}", "").
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			Title:     "updated",
			Type:      "text",
			UpdatedAt: now,
		}, nil)

	req := pb.UpdateSecretRequest_builder{
		Id:            proto.String(secretID.String()),
		Title:         proto.String("updated"),
		EncryptedData: []byte("encrypted"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String(""),
	}.Build()

	resp, err := server.UpdateSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetUpdatedAt())
}

func TestUpdateSecret_Unauthenticated(t *testing.T) {
	server, _, _ := newTestServer(t)

	req := pb.UpdateSecretRequest_builder{
		Id:    proto.String("some-id"),
		Title: proto.String("test"),
	}.Build()

	_, err := server.UpdateSecret(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestUpdateSecret_StorageError(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		UpdateSecret(mock.Anything, userID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("db error"))

	req := pb.UpdateSecretRequest_builder{
		Id:            proto.String("some-id"),
		Title:         proto.String("test"),
		EncryptedData: []byte("data"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String(""),
	}.Build()

	_, err := server.UpdateSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestDeleteSecret_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()

	store.EXPECT().
		DeleteSecret(mock.Anything, userID.String(), secretID.String()).
		Return(nil)

	req := pb.DeleteSecretRequest_builder{
		Id: proto.String(secretID.String()),
	}.Build()

	resp, err := server.DeleteSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
}

func TestDeleteSecret_Unauthenticated(t *testing.T) {
	server, _, _ := newTestServer(t)

	req := pb.DeleteSecretRequest_builder{
		Id: proto.String("some-id"),
	}.Build()

	_, err := server.DeleteSecret(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestDeleteSecret_StorageError(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		DeleteSecret(mock.Anything, userID.String(), "nonexistent").
		Return(fmt.Errorf("not found"))

	req := pb.DeleteSecretRequest_builder{
		Id: proto.String("nonexistent"),
	}.Build()

	_, err := server.DeleteSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetSecrets_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	now := time.Now()

	store.EXPECT().
		GetSecretsByUserID(mock.Anything, userID.String()).
		Return([]model.Secret{
			{
				ID:            uuid.New(),
				UserID:        userID,
				Title:         "Secret One",
				Type:          "credentials",
				EncryptedData: "enc1",
				Metadata:      json.RawMessage(`{}`),
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			{
				ID:            uuid.New(),
				UserID:        userID,
				Title:         "Secret Two",
				Type:          "text",
				EncryptedData: "enc2",
				Metadata:      json.RawMessage(`{}`),
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		}, nil)

	req := pb.GetSecretsRequest_builder{}.Build()

	resp, err := server.GetSecrets(ctxWithUserID(userID), req)
	require.NoError(t, err)
	require.Len(t, resp.GetSecrets(), 2)
	assert.Equal(t, "Secret One", resp.GetSecrets()[0].GetTitle())
	assert.Equal(t, pb.SecretType_SECRET_TYPE_CREDENTIALS, resp.GetSecrets()[0].GetType())
	assert.Equal(t, "Secret Two", resp.GetSecrets()[1].GetTitle())
	assert.Equal(t, pb.SecretType_SECRET_TYPE_TEXT, resp.GetSecrets()[1].GetType())
}

func TestGetSecrets_Empty(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		GetSecretsByUserID(mock.Anything, userID.String()).
		Return([]model.Secret{}, nil)

	req := pb.GetSecretsRequest_builder{}.Build()

	resp, err := server.GetSecrets(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.Empty(t, resp.GetSecrets())
}

func TestGetSecrets_Unauthenticated(t *testing.T) {
	server, _, _ := newTestServer(t)

	req := pb.GetSecretsRequest_builder{}.Build()

	_, err := server.GetSecrets(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGetSecrets_StorageError(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		GetSecretsByUserID(mock.Anything, userID.String()).
		Return(nil, fmt.Errorf("db error"))

	req := pb.GetSecretsRequest_builder{}.Build()

	_, err := server.GetSecrets(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetSecret_Success(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	store.EXPECT().
		GetSecretByID(mock.Anything, userID.String(), secretID.String()).
		Return(&model.Secret{
			ID:            secretID,
			UserID:        userID,
			Title:         "My Secret",
			Type:          "card",
			EncryptedData: "encrypted",
			Metadata:      json.RawMessage(`{"bank":"test"}`),
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil)

	req := pb.GetSecretRequest_builder{
		Id: proto.String(secretID.String()),
	}.Build()

	resp, err := server.GetSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.Equal(t, secretID.String(), resp.GetSecret().GetId())
	assert.Equal(t, "My Secret", resp.GetSecret().GetTitle())
	assert.Equal(t, pb.SecretType_SECRET_TYPE_CARD, resp.GetSecret().GetType())
}

func TestGetSecret_Unauthenticated(t *testing.T) {
	server, _, _ := newTestServer(t)

	req := pb.GetSecretRequest_builder{
		Id: proto.String("some-id"),
	}.Build()

	_, err := server.GetSecret(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestCreateSecret_Binary_Success(t *testing.T) {
	server, store, fileStore := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	store.EXPECT().
		CreateSecret(mock.Anything, userID.String(), "binary secret", "binary", "", mock.Anything, "file.bin").
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			Title:     "binary secret",
			Type:      "binary",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	objectKey := fmt.Sprintf("%s/%s", userID.String(), secretID.String())
	fileStore.EXPECT().
		Upload(mock.Anything, objectKey, []byte("binary-data")).
		Return(nil)

	store.EXPECT().
		UpdateSecret(mock.Anything, userID.String(), secretID.String(), "binary secret", objectKey, mock.Anything, "file.bin").
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			Title:     "binary secret",
			Type:      "binary",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	secretType := pb.SecretType_SECRET_TYPE_BINARY
	req := pb.CreateSecretRequest_builder{
		Title:         proto.String("binary secret"),
		Type:          &secretType,
		EncryptedData: []byte("binary-data"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String("file.bin"),
	}.Build()

	resp, err := server.CreateSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.Equal(t, secretID.String(), resp.GetId())
}

func TestCreateSecret_Binary_UploadError(t *testing.T) {
	server, store, fileStore := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	store.EXPECT().
		CreateSecret(mock.Anything, userID.String(), mock.Anything, "binary", "", mock.Anything, mock.Anything).
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	fileStore.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("upload failed"))

	secretType := pb.SecretType_SECRET_TYPE_BINARY
	req := pb.CreateSecretRequest_builder{
		Title:         proto.String("test"),
		Type:          &secretType,
		EncryptedData: []byte("data"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String("file.bin"),
	}.Build()

	_, err := server.CreateSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestCreateSecret_Binary_ExceedsMaxSize(t *testing.T) {
	server, _, _ := newTestServer(t)

	userID := uuid.New()
	largeData := make([]byte, maxBinarySecretSize+1)

	secretType := pb.SecretType_SECRET_TYPE_BINARY
	req := pb.CreateSecretRequest_builder{
		Title:         proto.String("too large"),
		Type:          &secretType,
		EncryptedData: largeData,
		Metadata:      proto.String("{}"),
		FilePath:      proto.String("big.bin"),
	}.Build()

	_, err := server.CreateSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUpdateSecret_Binary_Success(t *testing.T) {
	server, store, fileStore := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()

	objectKey := fmt.Sprintf("%s/%s", userID.String(), secretID.String())
	fileStore.EXPECT().
		Upload(mock.Anything, objectKey, []byte("new-binary")).
		Return(nil)

	store.EXPECT().
		UpdateSecret(mock.Anything, userID.String(), secretID.String(), "updated binary", objectKey, mock.Anything, "file.bin").
		Return(&model.Secret{
			ID:        secretID,
			UserID:    userID,
			Title:     "updated binary",
			Type:      "binary",
			UpdatedAt: now,
		}, nil)

	secretType := pb.SecretType_SECRET_TYPE_BINARY
	req := pb.UpdateSecretRequest_builder{
		Id:            proto.String(secretID.String()),
		Title:         proto.String("updated binary"),
		Type:          &secretType,
		EncryptedData: []byte("new-binary"),
		Metadata:      proto.String("{}"),
		FilePath:      proto.String("file.bin"),
	}.Build()

	resp, err := server.UpdateSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetUpdatedAt())
}

func TestGetSecret_Binary_Success(t *testing.T) {
	server, store, fileStore := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()
	objectKey := fmt.Sprintf("%s/%s", userID.String(), secretID.String())

	store.EXPECT().
		GetSecretByID(mock.Anything, userID.String(), secretID.String()).
		Return(&model.Secret{
			ID:            secretID,
			UserID:        userID,
			Title:         "binary secret",
			Type:          "binary",
			EncryptedData: objectKey,
			Metadata:      json.RawMessage(`{}`),
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil)

	fileStore.EXPECT().
		Download(mock.Anything, objectKey).
		Return([]byte("binary-content"), nil)

	req := pb.GetSecretRequest_builder{
		Id: proto.String(secretID.String()),
	}.Build()

	resp, err := server.GetSecret(ctxWithUserID(userID), req)
	require.NoError(t, err)
	assert.Equal(t, []byte("binary-content"), resp.GetSecret().GetEncryptedData())
}

func TestGetSecret_Binary_DownloadError(t *testing.T) {
	server, store, fileStore := newTestServer(t)

	userID := uuid.New()
	secretID := uuid.New()
	now := time.Now()
	objectKey := fmt.Sprintf("%s/%s", userID.String(), secretID.String())

	store.EXPECT().
		GetSecretByID(mock.Anything, userID.String(), secretID.String()).
		Return(&model.Secret{
			ID:            secretID,
			UserID:        userID,
			Title:         "binary secret",
			Type:          "binary",
			EncryptedData: objectKey,
			Metadata:      json.RawMessage(`{}`),
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil)

	fileStore.EXPECT().
		Download(mock.Anything, objectKey).
		Return(nil, fmt.Errorf("storage unavailable"))

	req := pb.GetSecretRequest_builder{
		Id: proto.String(secretID.String()),
	}.Build()

	_, err := server.GetSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetSecrets_OmitsBinaryData(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()
	now := time.Now()

	store.EXPECT().
		GetSecretsByUserID(mock.Anything, userID.String()).
		Return([]model.Secret{
			{
				ID:            uuid.New(),
				UserID:        userID,
				Title:         "Binary File",
				Type:          "binary",
				EncryptedData: "user-id/secret-id",
				Metadata:      json.RawMessage(`{}`),
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			{
				ID:            uuid.New(),
				UserID:        userID,
				Title:         "Text Secret",
				Type:          "text",
				EncryptedData: "some-text",
				Metadata:      json.RawMessage(`{}`),
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		}, nil)

	req := pb.GetSecretsRequest_builder{}.Build()

	resp, err := server.GetSecrets(ctxWithUserID(userID), req)
	require.NoError(t, err)
	require.Len(t, resp.GetSecrets(), 2)
	// Binary secret should have empty encrypted data
	assert.Empty(t, resp.GetSecrets()[0].GetEncryptedData())
	// Text secret should have data
	assert.Equal(t, []byte("some-text"), resp.GetSecrets()[1].GetEncryptedData())
}

func TestGetSecret_NotFound(t *testing.T) {
	server, store, _ := newTestServer(t)

	userID := uuid.New()

	store.EXPECT().
		GetSecretByID(mock.Anything, userID.String(), "nonexistent").
		Return(nil, fmt.Errorf("not found"))

	req := pb.GetSecretRequest_builder{
		Id: proto.String("nonexistent"),
	}.Build()

	_, err := server.GetSecret(ctxWithUserID(userID), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
