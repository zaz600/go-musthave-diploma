package accrual

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
)

type mockAccrualClient struct {
	mock.Mock
}

func (m *mockAccrualClient) GetOrderAccrualWithResponse(ctx context.Context, number Accrual.Order, _ ...Accrual.RequestEditorFn) (*Accrual.GetOrderAccrualResponse, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Accrual.GetOrderAccrualResponse), args.Error(1)
}

func TestClient_GetAccrual(t *testing.T) {
	accrualAPIClient := new(mockAccrualClient)
	accrualAPIClient.On("GetOrderAccrualWithResponse", context.TODO(), Accrual.Order("1")).Return(nil, fmt.Errorf("foo"))
	accrualProvider := NewProvider(accrualAPIClient)
	result := accrualProvider.getAccrual(context.TODO(), "1")
	assert.Error(t, result.Err)
	assert.Nil(t, result.Accrual)
	assert.Equal(t, Accrual.ResponseStatus(""), result.Status)
	accrualAPIClient.AssertExpectations(t)
}

func TestClient_GetAccrual2(t *testing.T) {
	accrualAPIClient := new(mockAccrualClient)
	amount := float32(555)
	resp := &Accrual.GetOrderAccrualResponse{
		Body: nil,
		HTTPResponse: &http.Response{
			Status:     "200",
			StatusCode: 200,
		},
		JSON200: &Accrual.Response{
			Accrual: &amount,
			Order:   "1",
			Status:  Accrual.ResponseStatusPROCESSED,
		},
	}

	accrualAPIClient.On("GetOrderAccrualWithResponse", context.TODO(), Accrual.Order("1")).Return(resp, nil)

	accrualProvider := NewProvider(accrualAPIClient)
	result := accrualProvider.getAccrual(context.TODO(), "1")
	assert.NoError(t, result.Err)
	assert.Equal(t, float32(555), *result.Accrual)
	assert.Equal(t, Accrual.ResponseStatusPROCESSED, result.Status)
	accrualAPIClient.AssertExpectations(t)
}

func TestClient_GetAccrual3(t *testing.T) {
	accrualAPIClient := new(mockAccrualClient)
	header := http.Header{}
	header.Add("Retry-After", "60")
	resp := &Accrual.GetOrderAccrualResponse{
		Body: nil,
		HTTPResponse: &http.Response{
			Status:     "429",
			StatusCode: http.StatusTooManyRequests,
			Header:     header,
		},
		JSON200: nil,
	}

	accrualAPIClient.On("GetOrderAccrualWithResponse", context.TODO(), Accrual.Order("1")).Return(resp, nil)

	accrualProvider := NewProvider(accrualAPIClient)
	result := accrualProvider.getAccrual(context.TODO(), "1")
	var actualError *TooManyRequestsError
	assert.ErrorAs(t, result.Err, &actualError)
	assert.Equal(t, 60, actualError.RetryAfterSec)
	assert.ErrorIs(t, result.Err, ErrTooManyRedirects)
	assert.Nil(t, result.Accrual)
	accrualAPIClient.AssertExpectations(t)
}
