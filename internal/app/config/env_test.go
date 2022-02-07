package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

func TestGetEnvOrDefault_Env_Exists(t *testing.T) {
	value := "abcdef"
	key := random.String(10)
	defValue := "foobarbaz"
	_ = os.Setenv(key, value)
	actual := getEnvOrDefault(key, defValue)
	assert.Equal(t, value, actual)
}

func TestGetEnvOrDefault_Env_not_Exists(t *testing.T) {
	key := random.String(10)
	defValue := "foobarbaz"

	actual := getEnvOrDefault(key, defValue)
	assert.Equal(t, defValue, actual)
}
