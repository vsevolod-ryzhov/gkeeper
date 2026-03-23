package grpcclient

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/model"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (c *Client) CreateSecret(ctx context.Context, token string, title string, secretType string, data map[string]interface{}) error {
	encryptedDataMap, err := c.prepareEncryptedDataMap(secretType, data)
	if err != nil {
		return err
	}

	plaintextJSON, err := json.Marshal(encryptedDataMap)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	encryptedData, err := c.crypto.Encrypt(plaintextJSON)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	metadataJSON, err := c.prepareMetadata(data)
	if err != nil {
		return err
	}

	filePath := ""
	if secretType == model.SecretTypeBinary {
		if path, ok := data["file_path"].(string); ok {
			filePath = path
		}
	}
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

	case model.SecretTypeBinary:
		return map[string]interface{}{
			"description": data["description"],
			"file_name":   data["file_name"],
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

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return string(metadataJSON), nil
}
