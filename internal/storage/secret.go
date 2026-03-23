package storage

import (
	"context"
	"fmt"

	"gkeeper/internal/model"

	sq "github.com/Masterminds/squirrel"
)

func (s *PostgresStorage) CreateSecret(ctx context.Context, userID string, title string, secretType string, encryptedData string, metadata string, filePath string) (*model.Secret, error) {
	var secret model.Secret

	query := sq.Insert("secrets").
		Columns("user_id", "title", "type", "encrypted_data", "metadata", "file_path").
		Values(userID, title, secretType, encryptedData, metadata, filePath).
		Suffix("RETURNING id, user_id, title, type, encrypted_data, metadata, file_path, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db)

	err := query.QueryRowContext(ctx).
		Scan(&secret.ID, &secret.UserID, &secret.Title, &secret.Type, &secret.EncryptedData, &secret.Metadata, &secret.FilePath, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}

	return &secret, nil
}
