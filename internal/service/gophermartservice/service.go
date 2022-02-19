package gophermartservice

import (
	"time"

	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/providers/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository"
)

const accrualDefaultRetryInterval = 50 * time.Millisecond

type GophermartService struct {
	repo repository.RepoRegistry

	accrualProvider      accrual.Provider
	accrualRetryInterval time.Duration
}

func New(accrualAPIClient Accrual.ClientWithResponsesInterface, opts ...Option) *GophermartService {
	s := &GophermartService{
		accrualProvider:      accrual.NewProvider(accrualAPIClient),
		accrualRetryInterval: accrualDefaultRetryInterval,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
