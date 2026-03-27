package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "salt", "created_at"}).
		AddRow(userID, "test@example.com", "hashed_pw", "salt123", now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test@example.com", "hashed_pw", "salt123").
		WillReturnRows(rows)

	user, err := s.CreateUser(context.Background(), "test@example.com", "hashed_pw", "salt123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != userID {
		t.Errorf("expected ID %s, got %s", userID, user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.PasswordHash != "hashed_pw" {
		t.Errorf("expected password_hash hashed_pw, got %s", user.PasswordHash)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected created_at %v, got %v", now, user.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test@example.com", "hashed_pw", "salt123").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err = s.CreateUser(context.Background(), "test@example.com", "hashed_pw", "salt123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test@example.com", "hashed_pw", "salt123").
		WillReturnError(dbErr)

	_, err = s.CreateUser(context.Background(), "test@example.com", "hashed_pw", "salt123")
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

func TestGetUserByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	userID := uuid.New()
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "salt", "created_at"}).
		AddRow(userID, "test@example.com", "hashed_pw", "salt123", now)

	mock.ExpectQuery(`SELECT (.+) FROM users`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := s.GetUserByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != userID {
		t.Errorf("expected ID %s, got %s", userID, user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.PasswordHash != "hashed_pw" {
		t.Errorf("expected password_hash hashed_pw, got %s", user.PasswordHash)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected created_at %v, got %v", now, user.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at"})

	mock.ExpectQuery(`SELECT (.+) FROM users`).
		WithArgs("missing@example.com").
		WillReturnRows(rows)

	_, err = s.GetUserByEmail(context.Background(), "missing@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
