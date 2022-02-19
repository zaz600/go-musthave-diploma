package accrual

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"go.uber.org/ratelimit"
)

type GetAccrualResponse struct {
	Status  entity.OrderStatus
	Accrual float32
	Err     error
}

type Client struct {
	accrualAPIClient Accrual.ClientWithResponsesInterface
	rateLimiter      ratelimit.Limiter
}

func (c Client) GetAccrual(ctx context.Context, orderID string) chan *GetAccrualResponse {
	resultCh := make(chan *GetAccrualResponse)
	go func() {
		c.rateLimiter.Take()
		resp := c.getAccrual(ctx, orderID)
		resultCh <- resp
		close(resultCh)
	}()
	return resultCh
}

func (c Client) getAccrual(ctx context.Context, orderID string) *GetAccrualResponse {
	resp, err := c.accrualAPIClient.GetOrderAccrualWithResponse(ctx, Accrual.Order(orderID))
	if err != nil {
		return &GetAccrualResponse{Err: err}
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfter := resp.HTTPResponse.Header.Get("Retry-After")
		retryAfterSec, err := strconv.Atoi(retryAfter)
		if err != nil {
			retryAfterSec = 5
		}
		return &GetAccrualResponse{Err: NewTooManyRequestsError(retryAfterSec)}
	}

	if resp.StatusCode() != 200 {
		return &GetAccrualResponse{Err: ErrWrongStatusCode}
	}

	log.Info().
		Str("orderID", orderID).
		Str("accrualStatus", string(resp.JSON200.Status)).
		Float32("accrual", *resp.JSON200.Accrual).
		Msg("get accrual result")

	switch resp.JSON200.Status {
	case Accrual.ResponseStatusINVALID:
		return &GetAccrualResponse{Accrual: 0.0, Status: entity.OrderStatusInvalid}
	case Accrual.ResponseStatusPROCESSED:
		return &GetAccrualResponse{Accrual: *resp.JSON200.Accrual, Status: entity.OrderStatusProcessed}
	case Accrual.ResponseStatusREGISTERED:
		return &GetAccrualResponse{Status: entity.OrderStatusProcessing}
	case Accrual.ResponseStatusPROCESSING:
		return &GetAccrualResponse{Status: entity.OrderStatusProcessing}
	}
	return &GetAccrualResponse{Err: ErrUnknownAccrualStatus}
}

func New(accrualAPIClient Accrual.ClientWithResponsesInterface) *Client {
	rl := ratelimit.New(1000, ratelimit.Per(1*time.Minute)) // per second
	return &Client{accrualAPIClient: accrualAPIClient, rateLimiter: rl}
}
