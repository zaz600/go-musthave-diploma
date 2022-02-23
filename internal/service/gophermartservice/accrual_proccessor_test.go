package gophermartservice

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/providers/accrual"
)

func TestGophermartService_calcNext(t *testing.T) {
	interval := 100 * time.Millisecond
	s := GophermartService{accrualRetryInterval: interval}

	t.Run("no error", func(t *testing.T) {
		resp := &accrual.GetAccrualResponse{
			Status: Accrual.ResponseStatusREGISTERED,
			Err:    nil,
		}
		next := s.calcNext(resp)
		assert.Equal(t, interval, next)
	})

	t.Run("too many requests error", func(t *testing.T) {
		resp := &accrual.GetAccrualResponse{
			Status: Accrual.ResponseStatusREGISTERED,
			Err:    accrual.TooManyRequestsError{RetryAfterSec: 60},
		}
		next := s.calcNext(resp)
		assert.Equal(t, 60*time.Second, next)
	})

	t.Run("other error", func(t *testing.T) {
		resp := &accrual.GetAccrualResponse{
			Status: Accrual.ResponseStatusREGISTERED,
			Err:    fmt.Errorf("boo"),
		}
		next := s.calcNext(resp)
		assert.Equal(t, interval, next)
	})
}
