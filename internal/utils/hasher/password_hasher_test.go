package hasher

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

func TestHashPassword(t *testing.T) {
	password := random.String(10)
	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEqual(t, password, hash)
	assert.NotEmpty(t, hash)

	hash2, err := HashPassword(strings.ToUpper(password))
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestCheckPasswordHash(t *testing.T) {
	password := random.String(10)
	hash, err := HashPassword(password)
	require.NoError(t, err)

	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash(strings.ToUpper(password), hash))
}
