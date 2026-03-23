package storage

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCreateSecret_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	secretID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	meta := json.RawMessage(`{"name":"test"}`)

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"}).
		AddRow(secretID, userID, "My Secret", "credentials", "encrypted", meta, nil, now, now)

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(userID.String(), "My Secret", "credentials", "encrypted", `{"name":"test"}`, "").
		WillReturnRows(rows)

	secret, err := s.CreateSecret(context.Background(), userID.String(), "My Secret", "credentials", "encrypted", `{"name":"test"}`, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if secret.ID != secretID {
		t.Errorf("expected ID %s, got %s", secretID, secret.ID)
	}
	if secret.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, secret.UserID)
	}
	if secret.Title != "My Secret" {
		t.Errorf("expected title 'My Secret', got %s", secret.Title)
	}
	if secret.Type != "credentials" {
		t.Errorf("expected type 'credentials', got %s", secret.Type)
	}
	if secret.EncryptedData != "encrypted" {
		t.Errorf("expected encrypted_data 'encrypted', got %s", secret.EncryptedData)
	}
	if !secret.CreatedAt.Equal(now) {
		t.Errorf("expected created_at %v, got %v", now, secret.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateSecret_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(userID.String(), "title", "text", "data", "{}", "").
		WillReturnError(dbErr)

	_, err = s.CreateSecret(context.Background(), userID.String(), "title", "text", "data", "{}", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("expected wrapped dbErr, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
