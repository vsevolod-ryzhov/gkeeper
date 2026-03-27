package grpcclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestExtractTokenFromHeader_WithBearer(t *testing.T) {
	header := metadata.Pairs("authorization", "Bearer my-token-123")

	token, err := extractTokenFromHeader(header)
	require.NoError(t, err)
	assert.Equal(t, "my-token-123", token)
}

func TestExtractTokenFromHeader_WithoutBearer(t *testing.T) {
	header := metadata.Pairs("authorization", "raw-token")

	token, err := extractTokenFromHeader(header)
	require.NoError(t, err)
	assert.Equal(t, "raw-token", token)
}

func TestExtractTokenFromHeader_Missing(t *testing.T) {
	header := metadata.Pairs()

	_, err := extractTokenFromHeader(header)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization token not found")
}
