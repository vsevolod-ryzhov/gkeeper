package storage

import (
	"context"
	"database/sql"
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

func (s *PostgresStorage) UpdateSecret(ctx context.Context, userID string, secretID string, title string, encryptedData string, metadata string, filePath string) (*model.Secret, error) {
	var secret model.Secret

	query := sq.Update("secrets").
		Set("title", title).
		Set("encrypted_data", encryptedData).
		Set("metadata", metadata).
		Set("file_path", filePath).
		Where(sq.Eq{"id": secretID, "user_id": userID}).
		Suffix("RETURNING id, user_id, title, type, encrypted_data, metadata, file_path, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db)

	err := query.QueryRowContext(ctx).
		Scan(&secret.ID, &secret.UserID, &secret.Title, &secret.Type, &secret.EncryptedData, &secret.Metadata, &secret.FilePath, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("update secret: %w", err)
	}

	return &secret, nil
}

func (s *PostgresStorage) GetSecretsByUserID(ctx context.Context, userID string) ([]model.Secret, error) {
	rows, err := sq.Select("id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at").
		From("secrets").
		Where(sq.Eq{"user_id": userID, "deleted_at": nil}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db).
		QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("get secrets: %w", err)
	}
	defer rows.Close()

	var secrets []model.Secret
	for rows.Next() {
		var secret model.Secret
		if err := rows.Scan(&secret.ID, &secret.UserID, &secret.Title, &secret.Type, &secret.EncryptedData, &secret.Metadata, &secret.FilePath, &secret.CreatedAt, &secret.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		secrets = append(secrets, secret)
	}

	return secrets, rows.Err()
}

func (s *PostgresStorage) GetSecretByID(ctx context.Context, userID string, secretID string) (*model.Secret, error) {
	var secret model.Secret

	err := sq.Select("id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at").
		From("secrets").
		Where(sq.Eq{"id": secretID, "user_id": userID, "deleted_at": nil}).
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db).
		QueryRowContext(ctx).
		Scan(&secret.ID, &secret.UserID, &secret.Title, &secret.Type, &secret.EncryptedData, &secret.Metadata, &secret.FilePath, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}

	return &secret, nil
}
