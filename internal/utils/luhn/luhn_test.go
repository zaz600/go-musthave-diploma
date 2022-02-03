package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLuhn(t *testing.T) {
	assert.True(t, CheckLuhn("12345678903"))
	assert.True(t, CheckLuhn("683458"))
	assert.False(t, CheckLuhn("92345678903"))
	assert.False(t, CheckLuhn("sdfsdfd"))
}
