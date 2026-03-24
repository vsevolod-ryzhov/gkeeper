package grpcclient

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/model"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// encryptSecretData handles encryption for all secret types.
// For binary secrets, it reads the file from disk and encrypts the raw bytes.
// For other types, it JSON-marshals the structured data and encrypts that.
// Returns the encrypted data string and file path (empty for non-binary).
func (c *Client) encryptSecretData(secretType string, data map[string]interface{}) (string, string, error) {
	if secretType == model.SecretTypeBinary {
		return c.encryptBinaryFile(data)
	}

	encryptedDataMap, err := c.prepareEncryptedDataMap(secretType, data)
	if err != nil {
		return "", "", err
	}

	plaintextJSON, err := json.Marshal(encryptedDataMap)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal data: %w", err)
	}

	encryptedData, err := c.crypto.Encrypt(plaintextJSON)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encryptedData, "", nil
}

// encryptBinaryFile reads a file from disk, encrypts its contents, and returns the encrypted data along with the original filename stored in metadata.
func (c *Client) encryptBinaryFile(data map[string]interface{}) (string, string, error) {
	filePath, ok := data["file_path"].(string)
	if !ok || filePath == "" {
		return "", "", fmt.Errorf("file path is required for binary secrets")
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	encryptedData, err := c.crypto.Encrypt(fileContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt file: %w", err)
	}

	fileName := filepath.Base(filePath)

	return encryptedData, fileName, nil
}

// DecryptBinarySecret decrypts a binary secret and writes it to the specified path.
func (c *Client) DecryptBinarySecret(encryptedData []byte, savePath string) error {
	decrypted, err := c.crypto.Decrypt(string(encryptedData))
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	if err := os.WriteFile(savePath, decrypted, 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", savePath, err)
	}

	return nil
}

func (c *Client) UpdateSecret(ctx context.Context, token string, id string, title string, secretType string, data map[string]interface{}) error {
	encryptedData, binaryFileName, err := c.encryptSecretData(secretType, data)
	if err != nil {
		return err
	}

	metadataJSON, err := c.prepareMetadata(data)
	if err != nil {
		return err
	}

	filePath := binaryFileName
	if secretType == model.SecretTypeBinary && filePath == "" {
		if path, ok := data["file_path"].(string); ok {
			filePath = filepath.Base(path)
		}
	}

	ctxWithToken := c.createContextWithToken(ctx)

	response, reqErr := c.client.UpdateSecret(ctxWithToken, pb.UpdateSecretRequest_builder{
		Token:         proto.String(token),
		Id:            proto.String(id),
		Title:         proto.String(title),
		Type:          proto.String(secretType),
		EncryptedData: []byte(encryptedData),
		Metadata:      proto.String(metadataJSON),
		FilePath:      proto.String(filePath),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	c.logger.Debug("UpdateSecret response", zap.String("response", response.String()))

	return nil
}

func (c *Client) CreateSecret(ctx context.Context, token string, title string, secretType string, data map[string]interface{}) error {
	encryptedData, binaryFileName, err := c.encryptSecretData(secretType, data)
	if err != nil {
		return err
	}

	metadataJSON, err := c.prepareMetadata(data)
	if err != nil {
		return err
	}

	filePath := binaryFileName

	ctxWithToken := c.createContextWithToken(ctx)

	response, reqErr := c.client.CreateSecret(ctxWithToken, pb.CreateSecretRequest_builder{
		Token:         proto.String(token),
		Title:         proto.String(title),
		Type:          proto.String(secretType),
		EncryptedData: []byte(encryptedData),
		Metadata:      proto.String(metadataJSON),
		FilePath:      proto.String(filePath),
	}.Build())

	if reqErr != nil {
		return reqErr
	}

	c.logger.Info("Secret created successfully", zap.String("id", response.GetId()))
	return nil
}

func (c *Client) GetSecrets(ctx context.Context, token string) ([]*pb.Secret, error) {
	ctxWithToken := c.createContextWithToken(ctx)

	response, err := c.client.GetSecrets(ctxWithToken, pb.GetSecretsRequest_builder{
		Token: proto.String(token),
	}.Build())

	if err != nil {
		return nil, err
	}

	return response.GetSecrets(), nil
}

func (c *Client) prepareEncryptedDataMap(secretType string, data map[string]interface{}) (map[string]interface{}, error) {
	switch secretType {
	case model.SecretTypeCredentials:
		return map[string]interface{}{
			"username": data["username"],
			"password": data["password"],
			"url":      data["url"],
			"notes":    data["notes"],
		}, nil

	case model.SecretTypeText:
		return map[string]interface{}{
			"content": data["content"],
			"notes":   data["notes"],
		}, nil

	case model.SecretTypeCard:
		return map[string]interface{}{
			"card_number":      data["card_number"],
			"card_holder_name": data["card_holder_name"],
			"expiry_date":      data["expiry_date"],
			"cvv":              data["cvv"],
			"notes":            data["notes"],
		}, nil

	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
}

func (c *Client) prepareMetadata(data map[string]interface{}) (string, error) {
	metadata := make(map[string]interface{})

	if name, ok := data["name"].(string); ok {
		metadata["name"] = name
	}

	if tags, ok := data["tags"].([]string); ok {
		metadata["tags"] = tags
	}

	if customMetadata, ok := data["metadata"].(map[string]string); ok {
		for k, v := range customMetadata {
			metadata[k] = v
		}
	}

	if notes, ok := data["notes"].(string); ok && notes != "" {
		metadata["notes"] = notes
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return string(metadataJSON), nil
}
