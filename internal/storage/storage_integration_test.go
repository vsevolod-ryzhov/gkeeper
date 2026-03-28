//go:build integration

package storage

import (
	"context"
	"os"
	"testing"
)

const testConnString = "postgres://postgres_user:postgres_password@localhost:5432/postgres_db?sslmode=disable"

func TestMain(m *testing.M) {
	// applyMigrations uses "file://migrations" relative to working directory,
	// so we need to run from the project root.
	if err := os.Chdir("../.."); err != nil {
		panic("failed to change to project root: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestNewPostgresStorage_Success(t *testing.T) {
	s, err := NewPostgresStorage(testConnString)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := s.Ping(context.Background()); err != nil {
		t.Errorf("expected ping to succeed, got %v", err)
	}
}

func TestNewPostgresStorage_InvalidConnection(t *testing.T) {
	_, err := NewPostgresStorage("postgres://bad:bad@localhost:9999/nope?sslmode=disable")
	if err == nil {
		t.Fatal("expected error for invalid connection, got nil")
	}
}

func TestCreateUser_Integration(t *testing.T) {
	s, err := NewPostgresStorage(testConnString)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Clean up before and after test
	s.db.Exec("DELETE FROM users WHERE email = $1", "integration@test.com")
	t.Cleanup(func() {
		s.db.Exec("DELETE FROM users WHERE email = $1", "integration@test.com")
	})

	user, err := s.CreateUser(context.Background(), "integration@test.com", "hashed_pw")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if user.Email != "integration@test.com" {
		t.Errorf("expected email integration@test.com, got %s", user.Email)
	}

	// Attempt duplicate — should return ErrUserAlreadyExists
	_, err = s.CreateUser(context.Background(), "integration@test.com", "hashed_pw")
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}
