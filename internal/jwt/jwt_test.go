package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAndVerifyToken_Success(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	userID := uuid.New()
	email := "user@example.com"

	token, err := manager.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	manager := NewJWTManager("test-secret", -time.Hour)
	userID := uuid.New()

	token, err := manager.GenerateToken(userID, "user@example.com")
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	_, err = manager.VerifyToken(token)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)

	_, err := manager.VerifyToken("not.a.valid.token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	generator := NewJWTManager("secret-one", time.Hour)
	verifier := NewJWTManager("secret-two", time.Hour)

	token, err := generator.GenerateToken(uuid.New(), "user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = verifier.VerifyToken(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
