package grpcclient

import (
	"context"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/crypto"
	mockproto "gkeeper/internal/mocks/proto"
	"gkeeper/internal/model"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func newTestClient(mockClient *mockproto.MockGKeeperClient) *Client {
	cryptoObj, _ := crypto.NewCryptoFromPassword("testpassword", make([]byte, 32))
	return &Client{
		client: mockClient,
		logger: zap.NewNop(),
		crypto: cryptoObj,
		token:  "test-token",
	}
}

func TestCreateSecret_Success(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		CreateSecret(mock.Anything, mock.MatchedBy(func(req *pb.CreateSecretRequest) bool {
			return req.GetTitle() == "my secret" && req.GetType() == pb.SecretType_SECRET_TYPE_CREDENTIALS
		}), mock.Anything).
		Return(pb.CreateSecretResponse_builder{
			Id:        proto.String("secret-id-123"),
			CreatedAt: proto.String("2026-01-01T00:00:00Z"),
		}.Build(), nil)

	c := newTestClient(m)
	data := map[string]interface{}{
		"username": "user1",
		"password": "pass1",
		"url":      "https://example.com",
		"notes":    "",
	}

	err := c.CreateSecret(context.Background(), "my secret", model.SecretTypeCredentials, data)
	require.NoError(t, err)
}

func TestCreateSecret_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		CreateSecret(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("server unavailable"))

	c := newTestClient(m)
	data := map[string]interface{}{
		"content": "some text",
		"notes":   "",
	}

	err := c.CreateSecret(context.Background(), "text secret", model.SecretTypeText, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server unavailable")
}

func TestUpdateSecret_Success(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		UpdateSecret(mock.Anything, mock.MatchedBy(func(req *pb.UpdateSecretRequest) bool {
			return req.GetId() == "secret-id-123" &&
				req.GetTitle() == "updated title" &&
				req.GetType() == pb.SecretType_SECRET_TYPE_CARD
		}), mock.Anything).
		Return(pb.UpdateSecretResponse_builder{
			UpdatedAt: proto.String("2026-01-02T00:00:00Z"),
		}.Build(), nil)

	c := newTestClient(m)
	data := map[string]interface{}{
		"card_number":      "4111111111111111",
		"card_holder_name": "Test User",
		"expiry_date":      "12/28",
		"cvv":              "123",
		"notes":            "",
	}

	err := c.UpdateSecret(context.Background(), "secret-id-123", "updated title", model.SecretTypeCard, data)
	require.NoError(t, err)
}

func TestUpdateSecret_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		UpdateSecret(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("permission denied"))

	c := newTestClient(m)
	data := map[string]interface{}{
		"content": "text",
		"notes":   "",
	}

	err := c.UpdateSecret(context.Background(), "id", "title", model.SecretTypeText, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestDeleteSecret_Success(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		DeleteSecret(mock.Anything, mock.MatchedBy(func(req *pb.DeleteSecretRequest) bool {
			return req.GetId() == "secret-id-456"
		}), mock.Anything).
		Return(pb.DeleteSecretResponse_builder{
			Success: proto.Bool(true),
		}.Build(), nil)

	c := newTestClient(m)

	err := c.DeleteSecret(context.Background(), "secret-id-456")
	require.NoError(t, err)
}

func TestDeleteSecret_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		DeleteSecret(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("not found"))

	c := newTestClient(m)

	err := c.DeleteSecret(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetSecrets_Success(t *testing.T) {
	secretType := pb.SecretType_SECRET_TYPE_TEXT
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		GetSecrets(mock.Anything, mock.Anything, mock.Anything).
		Return(pb.GetSecretsResponse_builder{
			Secrets: []*pb.Secret{
				pb.Secret_builder{
					Id:    proto.String("s1"),
					Title: proto.String("Secret One"),
					Type:  &secretType,
				}.Build(),
				pb.Secret_builder{
					Id:    proto.String("s2"),
					Title: proto.String("Secret Two"),
					Type:  &secretType,
				}.Build(),
			},
		}.Build(), nil)

	c := newTestClient(m)

	secrets, err := c.GetSecrets(context.Background())
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "s1", secrets[0].GetId())
	assert.Equal(t, "Secret Two", secrets[1].GetTitle())
}

func TestGetSecrets_ServerError(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)
	m.EXPECT().
		GetSecrets(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("connection refused"))

	c := newTestClient(m)

	secrets, err := c.GetSecrets(context.Background())
	require.Error(t, err)
	assert.Nil(t, secrets)
}

func TestPrepareEncryptedDataMap_Credentials(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))
	data := map[string]interface{}{
		"username": "user",
		"password": "pass",
		"url":      "https://example.com",
		"notes":    "a note",
	}

	result, err := c.prepareEncryptedDataMap(model.SecretTypeCredentials, data)
	require.NoError(t, err)
	assert.Equal(t, "user", result["username"])
	assert.Equal(t, "pass", result["password"])
	assert.Equal(t, "https://example.com", result["url"])
	assert.Equal(t, "a note", result["notes"])
}

func TestPrepareEncryptedDataMap_Text(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))
	data := map[string]interface{}{
		"content": "hello world",
		"notes":   "",
	}

	result, err := c.prepareEncryptedDataMap(model.SecretTypeText, data)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result["content"])
}

func TestPrepareEncryptedDataMap_Card(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))
	data := map[string]interface{}{
		"card_number":      "4111111111111111",
		"card_holder_name": "Test",
		"expiry_date":      "12/28",
		"cvv":              "123",
		"notes":            "",
	}

	result, err := c.prepareEncryptedDataMap(model.SecretTypeCard, data)
	require.NoError(t, err)
	assert.Equal(t, "4111111111111111", result["card_number"])
	assert.Equal(t, "Test", result["card_holder_name"])
}

func TestPrepareEncryptedDataMap_UnknownType(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	_, err := c.prepareEncryptedDataMap("unknown", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown secret type")
}

func TestPrepareMetadata(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))
	data := map[string]interface{}{
		"name":  "my-secret",
		"notes": "important",
		"metadata": map[string]string{
			"env": "production",
		},
	}

	result, err := c.prepareMetadata(data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "my-secret")
	assert.Contains(t, result, "important")
	assert.Contains(t, result, "production")
}

func TestCreateContextWithToken(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))
	c.token = "my-jwt-token"

	ctx := c.createContextWithToken(context.Background())
	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)

	values := md.Get("authorization")
	require.NotEmpty(t, values)
	assert.Equal(t, "Bearer my-jwt-token", values[0])
}

func TestEncryptBinaryFile_Success(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	tmpFile := filepath.Join(t.TempDir(), "testfile.txt")
	err := os.WriteFile(tmpFile, []byte("binary content"), 0600)
	require.NoError(t, err)

	data := map[string]interface{}{
		"file_path": tmpFile,
	}

	encrypted, fileName, err := c.encryptBinaryFile(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.Equal(t, "testfile.txt", fileName)
}

func TestEncryptBinaryFile_MissingPath(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	data := map[string]interface{}{}

	_, _, err := c.encryptBinaryFile(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file path is required")
}

func TestEncryptBinaryFile_FileNotFound(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	data := map[string]interface{}{
		"file_path": "/nonexistent/file.bin",
	}

	_, _, err := c.encryptBinaryFile(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestEncryptSecretData_Binary(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	tmpFile := filepath.Join(t.TempDir(), "doc.pdf")
	err := os.WriteFile(tmpFile, []byte("pdf content"), 0600)
	require.NoError(t, err)

	data := map[string]interface{}{
		"file_path": tmpFile,
	}

	encrypted, filePath, err := c.encryptSecretData(model.SecretTypeBinary, data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.Equal(t, "doc.pdf", filePath)
}

func TestDecryptBinarySecret_Success(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	plaintext := []byte("secret binary data")
	encrypted, err := c.crypto.Encrypt(plaintext)
	require.NoError(t, err)

	savePath := filepath.Join(t.TempDir(), "decrypted.bin")
	err = c.DecryptBinarySecret([]byte(encrypted), savePath)
	require.NoError(t, err)

	content, err := os.ReadFile(savePath)
	require.NoError(t, err)
	assert.Equal(t, plaintext, content)
}

func TestDecryptBinarySecret_InvalidData(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	err := c.DecryptBinarySecret([]byte("not-encrypted"), filepath.Join(t.TempDir(), "out.bin"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
}

func TestDecryptBinarySecret_InvalidPath(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	plaintext := []byte("data")
	encrypted, err := c.crypto.Encrypt(plaintext)
	require.NoError(t, err)

	err = c.DecryptBinarySecret([]byte(encrypted), "/nonexistent/dir/file.bin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write file")
}

func TestPrepareMetadata_Empty(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	data := map[string]interface{}{}

	result, err := c.prepareMetadata(data)
	require.NoError(t, err)
	assert.Equal(t, "{}", result)
}

func TestPrepareMetadata_WithTags(t *testing.T) {
	c := newTestClient(mockproto.NewMockGKeeperClient(t))

	data := map[string]interface{}{
		"tags": []string{"personal", "finance"},
	}

	result, err := c.prepareMetadata(data)
	require.NoError(t, err)
	assert.Contains(t, result, "personal")
	assert.Contains(t, result, "finance")
}

func TestCreateSecret_BinaryType(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	tmpFile := filepath.Join(t.TempDir(), "upload.bin")
	err := os.WriteFile(tmpFile, []byte("file data"), 0600)
	require.NoError(t, err)

	m.EXPECT().
		CreateSecret(mock.Anything, mock.MatchedBy(func(req *pb.CreateSecretRequest) bool {
			return req.GetType() == pb.SecretType_SECRET_TYPE_BINARY &&
				req.GetFilePath() == "upload.bin"
		}), mock.Anything).
		Return(pb.CreateSecretResponse_builder{
			Id:        proto.String("binary-id"),
			CreatedAt: proto.String("2026-01-01T00:00:00Z"),
		}.Build(), nil)

	c := newTestClient(m)
	data := map[string]interface{}{
		"file_path": tmpFile,
	}

	err = c.CreateSecret(context.Background(), "binary secret", model.SecretTypeBinary, data)
	require.NoError(t, err)
}

func TestUpdateSecret_BinaryWithExistingFilePath(t *testing.T) {
	m := mockproto.NewMockGKeeperClient(t)

	m.EXPECT().
		UpdateSecret(mock.Anything, mock.MatchedBy(func(req *pb.UpdateSecretRequest) bool {
			return req.GetType() == pb.SecretType_SECRET_TYPE_BINARY &&
				req.GetFilePath() == "existing.bin"
		}), mock.Anything).
		Return(pb.UpdateSecretResponse_builder{
			UpdatedAt: proto.String("2026-01-02T00:00:00Z"),
		}.Build(), nil)

	c := newTestClient(m)

	tmpFile := filepath.Join(t.TempDir(), "existing.bin")
	err := os.WriteFile(tmpFile, []byte("content"), 0600)
	require.NoError(t, err)

	data := map[string]interface{}{
		"file_path": tmpFile,
	}

	err = c.UpdateSecret(context.Background(), "id-1", "binary update", model.SecretTypeBinary, data)
	require.NoError(t, err)
}
