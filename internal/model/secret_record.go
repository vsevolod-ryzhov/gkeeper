package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	SecretTypeCredentials = "credentials"
	SecretTypeText        = "text"
	SecretTypeCard        = "card"
	SecretTypeBinary      = "binary"
)

type EncryptedPayload struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	URL      string `json:"url,omitempty"`

	Content string `json:"content,omitempty"`

	CardNumber     string `json:"card_number,omitempty"`
	CardHolderName string `json:"card_holder_name,omitempty"`
	ExpiryDate     string `json:"expiry_date,omitempty"`
	CVV            string `json:"cvv,omitempty"`

	Notes string `json:"notes,omitempty"`
}

type Metadata struct {
	Name         string            `json:"name"`
	Tags         []string          `json:"tags,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

type Secret struct {
	ID            uuid.UUID       `db:"id"`
	UserID        uuid.UUID       `db:"user_id"`
	Title         string          `db:"title"`
	Type          string          `db:"type"`
	EncryptedData string          `db:"encrypted_data"`
	Metadata      json.RawMessage `db:"metadata"`
	FilePath      *string         `db:"file_path"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	DeletedAt     *time.Time      `db:"deleted_at"`
}
