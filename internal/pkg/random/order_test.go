package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderID(t *testing.T) {
	assert.Len(t, OrderID(), 16)
	assert.NotEqual(t, OrderID(), OrderID())
}
