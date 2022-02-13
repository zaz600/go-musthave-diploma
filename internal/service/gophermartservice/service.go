package gophermartservice

import (
	"time"

	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/webclient/accrualclient"
)

const accrualDefaultRetryInterval = 50 * time.Millisecond

type GophermartService struct {
	repo repository.RepoRegistry

	accrualClient        *accrualclient.Client
	accrualRetryInterval time.Duration
}

func New(accrualAPIClient Accrual.ClientWithResponsesInterface, opts ...Option) *GophermartService {
	s := &GophermartService{
		accrualClient:        accrualclient.New(accrualAPIClient),
		accrualRetryInterval: accrualDefaultRetryInterval,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
