package model

import (
	"encoding/json"
	pb "gkeeper/api/proto"
	"time"

	"github.com/google/uuid"
)

const (
	// SecretTypeCredentials represents a login/password credential secret.
	SecretTypeCredentials = "credentials"
	// SecretTypeText represents an arbitrary text secret.
	SecretTypeText = "text"
	// SecretTypeCard represents a bank card secret.
	SecretTypeCard = "card"
	// SecretTypeBinary represents an arbitrary binary data secret.
	SecretTypeBinary = "binary"
)

// SecretTypeToProto converts a string secret type to the protobuf enum value.
func SecretTypeToProto(s string) pb.SecretType {
	switch s {
	case SecretTypeCredentials:
		return pb.SecretType_SECRET_TYPE_CREDENTIALS
	case SecretTypeText:
		return pb.SecretType_SECRET_TYPE_TEXT
	case SecretTypeCard:
		return pb.SecretType_SECRET_TYPE_CARD
	case SecretTypeBinary:
		return pb.SecretType_SECRET_TYPE_BINARY
	default:
		return pb.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

// ProtoToSecretType converts a protobuf enum value to a string secret type.
func ProtoToSecretType(t pb.SecretType) string {
	switch t {
	case pb.SecretType_SECRET_TYPE_CREDENTIALS:
		return SecretTypeCredentials
	case pb.SecretType_SECRET_TYPE_TEXT:
		return SecretTypeText
	case pb.SecretType_SECRET_TYPE_CARD:
		return SecretTypeCard
	case pb.SecretType_SECRET_TYPE_BINARY:
		return SecretTypeBinary
	default:
		return ""
	}
}

// EncryptedPayload holds the decrypted fields of a secret, varying by secret type.
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

// Metadata holds arbitrary descriptive information associated with a secret.
type Metadata struct {
	Name         string            `json:"name"`
	Tags         []string          `json:"tags,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

// Secret represents an encrypted secret record stored in the database.
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
