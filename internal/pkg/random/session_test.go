package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionID(t *testing.T) {
	assert.Len(t, SessionID(), 32)
	assert.NotEqual(t, SessionID(), SessionID())
}
