package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &PostgresStorage{db: db}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at"}).
		AddRow(int64(1), "test@example.com", "hashed_pw", now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test@example.com", "hashed_pw").
		WillReturnRows(rows)

	user, err := s.CreateUser(context.Background(), "test@example.com", "hashed_pw")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
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
		WithArgs("test@example.com", "hashed_pw").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err = s.CreateUser(context.Background(), "test@example.com", "hashed_pw")
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
