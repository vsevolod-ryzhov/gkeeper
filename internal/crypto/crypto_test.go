package crypto

import (
	"bytes"
	"testing"
)

func TestNewCryptoFromPassword_WithSalt(t *testing.T) {
	salt := make([]byte, saltSize)
	c, err := NewCryptoFromPassword("password", salt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Crypto")
	}
	if len(c.masterKey) != keySize {
		t.Errorf("expected key size %d, got %d", keySize, len(c.masterKey))
	}
}

func TestNewCryptoFromPassword_GeneratesSalt(t *testing.T) {
	c, err := NewCryptoFromPassword("password", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Crypto")
	}
}

func TestNewCryptoFromPassword_Deterministic(t *testing.T) {
	salt := []byte("fixed-salt-for-testing-32-bytes!")
	c1, _ := NewCryptoFromPassword("password", salt)
	c2, _ := NewCryptoFromPassword("password", salt)

	if !bytes.Equal(c1.masterKey, c2.masterKey) {
		t.Error("same password and salt should produce the same key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	c, _ := NewCryptoFromPassword("password", nil)
	plaintext := []byte("hello secret world")

	encrypted, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if encrypted == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	decrypted, err := c.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	salt := []byte("fixed-salt-for-testing-32-bytes!")
	c1, _ := NewCryptoFromPassword("password1", salt)
	c2, _ := NewCryptoFromPassword("password2", salt)

	encrypted, _ := c1.Encrypt([]byte("secret"))

	_, err := c2.Decrypt(encrypted)
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	c, _ := NewCryptoFromPassword("password", nil)

	_, err := c.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	c, _ := NewCryptoFromPassword("password", nil)

	_, err := c.Decrypt("YQ==") // base64 of "a" — too short for nonce
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
