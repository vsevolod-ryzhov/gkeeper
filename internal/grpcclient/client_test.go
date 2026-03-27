package grpcclient

import (
	"gkeeper/internal/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetGetToken(t *testing.T) {
	c := &Client{logger: zap.NewNop()}

	c.SetToken("token-123")
	assert.Equal(t, "token-123", c.GetToken())

	c.SetToken("token-456")
	assert.Equal(t, "token-456", c.GetToken())
}

func TestSetGetUserID(t *testing.T) {
	c := &Client{logger: zap.NewNop()}

	c.SetUserID("user-1")
	assert.Equal(t, "user-1", c.GetUserID())
}

func TestSetGetEmail(t *testing.T) {
	c := &Client{logger: zap.NewNop()}

	c.SetEmail("test@example.com")
	assert.Equal(t, "test@example.com", c.GetEmail())
}

func TestSetGetCrypto(t *testing.T) {
	c := &Client{logger: zap.NewNop()}

	assert.Nil(t, c.GetCrypto())

	cryptoObj, err := crypto.NewCryptoFromPassword("password", make([]byte, 32))
	assert.NoError(t, err)

	c.SetCrypto(cryptoObj)
	assert.Equal(t, cryptoObj, c.GetCrypto())
}
