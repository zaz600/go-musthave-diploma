package accrualclient

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type GetAccrualResponse struct {
	Status  entity.OrderStatus
	Accrual float32
	Err     error
}

type Client struct {
	accrualAPIClient Accrual.ClientWithResponsesInterface
	limitCh          chan bool
}

func (c Client) GetAccrual(ctx context.Context, orderID string) chan *GetAccrualResponse {
	resultCh := make(chan *GetAccrualResponse)
	go func() {
		c.limitCh <- true
		resp := c.getAccrual(ctx, orderID)
		resultCh <- resp
		close(resultCh)
		<-c.limitCh
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
		return &GetAccrualResponse{Accrual: 0.0, Status: entity.OrderStatusINVALID}
	case Accrual.ResponseStatusPROCESSED:
		return &GetAccrualResponse{Accrual: *resp.JSON200.Accrual, Status: entity.OrderStatusPROCESSED}
	case Accrual.ResponseStatusREGISTERED:
		return &GetAccrualResponse{Status: entity.OrderStatusPROCESSING}
	case Accrual.ResponseStatusPROCESSING:
		return &GetAccrualResponse{Status: entity.OrderStatusPROCESSING}
	}
	return &GetAccrualResponse{Err: ErrUnknownAccrualStatus}
}

func New(accrualAPIClient Accrual.ClientWithResponsesInterface) *Client {
	return &Client{accrualAPIClient: accrualAPIClient, limitCh: make(chan bool, 10)}
}
