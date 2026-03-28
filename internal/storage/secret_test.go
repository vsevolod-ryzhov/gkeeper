package storage

import (
	"context"
	"database/sql"
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

func TestUpdateSecret_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	secretID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	meta := json.RawMessage(`{"url":"example.com"}`)

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"}).
		AddRow(secretID, userID, "Updated Title", "credentials", "new_encrypted", meta, nil, now, now)

	mock.ExpectQuery(`UPDATE secrets`).
		WithArgs("Updated Title", "new_encrypted", `{"url":"example.com"}`, "", secretID.String(), userID.String()).
		WillReturnRows(rows)

	secret, err := s.UpdateSecret(context.Background(), userID.String(), secretID.String(), "Updated Title", "new_encrypted", `{"url":"example.com"}`, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if secret.ID != secretID {
		t.Errorf("expected ID %s, got %s", secretID, secret.ID)
	}
	if secret.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", secret.Title)
	}
	if secret.EncryptedData != "new_encrypted" {
		t.Errorf("expected encrypted_data 'new_encrypted', got %s", secret.EncryptedData)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateSecret_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	secretID := uuid.New()
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`UPDATE secrets`).
		WithArgs("title", "data", "{}", "", secretID.String(), userID.String()).
		WillReturnError(dbErr)

	_, err = s.UpdateSecret(context.Background(), userID.String(), secretID.String(), "title", "data", "{}", "")
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

func TestUpdateSecret_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	secretID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"})

	mock.ExpectQuery(`UPDATE secrets`).
		WithArgs("title", "data", "{}", "", secretID.String(), userID.String()).
		WillReturnRows(rows)

	_, err = s.UpdateSecret(context.Background(), userID.String(), secretID.String(), "title", "data", "{}", "")
	if err == nil {
		t.Fatal("expected error for non-existent secret, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretsByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	secretID1 := uuid.New()
	secretID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"}).
		AddRow(secretID1, userID, "Secret One", "credentials", "enc1", json.RawMessage(`{}`), nil, now, now).
		AddRow(secretID2, userID, "Secret Two", "text", "enc2", json.RawMessage(`{}`), nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	secrets, err := s.GetSecretsByUserID(context.Background(), userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if secrets[0].ID != secretID1 {
		t.Errorf("expected first secret ID %s, got %s", secretID1, secrets[0].ID)
	}
	if secrets[0].Title != "Secret One" {
		t.Errorf("expected title 'Secret One', got %s", secrets[0].Title)
	}
	if secrets[1].ID != secretID2 {
		t.Errorf("expected second secret ID %s, got %s", secretID2, secrets[1].ID)
	}
	if secrets[1].Type != "text" {
		t.Errorf("expected type 'text', got %s", secrets[1].Type)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretsByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	secrets, err := s.GetSecretsByUserID(context.Background(), userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets, got %d", len(secrets))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretsByUserID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()

	// Return a row with wrong column types to trigger a scan error
	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"}).
		AddRow("not-a-uuid", userID, "Title", "text", "enc", "{}", nil, "not-a-time", "not-a-time")

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	_, err = s.GetSecretsByUserID(context.Background(), userID.String())
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretsByUserID_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(userID.String()).
		WillReturnError(dbErr)

	_, err = s.GetSecretsByUserID(context.Background(), userID.String())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	secretID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "type", "encrypted_data", "metadata", "file_path", "created_at", "updated_at"}).
		AddRow(secretID, userID, "My Secret", "credentials", "encrypted", json.RawMessage(`{}`), nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(secretID.String(), userID.String()).
		WillReturnRows(rows)

	secret, err := s.GetSecretByID(context.Background(), userID.String(), secretID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if secret.ID != secretID {
		t.Errorf("expected ID %s, got %s", secretID, secret.ID)
	}
	if secret.Title != "My Secret" {
		t.Errorf("expected title 'My Secret', got %s", secret.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	secretID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(secretID.String(), userID.String()).
		WillReturnError(sql.ErrNoRows)

	_, err = s.GetSecretByID(context.Background(), userID.String(), secretID.String())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetSecretByID_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	secretID := uuid.New()
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`SELECT (.+) FROM secrets`).
		WithArgs(secretID.String(), userID.String()).
		WillReturnError(dbErr)

	_, err = s.GetSecretByID(context.Background(), userID.String(), secretID.String())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
