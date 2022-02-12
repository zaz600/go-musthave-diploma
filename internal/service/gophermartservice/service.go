package gophermartservice

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/webclient/accrualclient"
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/withdrawalservice"
	"github.com/zaz600/go-musthave-diploma/internal/utils/auth"
)

const accrualDefaultRetryInterval = 50 * time.Millisecond

type GophermartService struct {
	userService       userservice.UserService
	sessionService    sessionservice.SessionService
	OrderService      orderservice.OrderService
	WithdrawalService withdrawalservice.WithdrawalService

	accrualClient        *accrualclient.Client
	accrualRetryInterval time.Duration
}

func (s GophermartService) SetJWT(w http.ResponseWriter, session *entity.Session) error {
	jwtToken, err := auth.CreateToken(session.UID, session.SessionID)
	if err != nil {
		return err
	}
	w.Header().Set("Authorization", "Bearer "+jwtToken)
	return nil
}

func (s GophermartService) GetUserID(r *http.Request) (string, error) {
	jwtToken := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
	if jwtToken == "" {
		return "", ErrUserNotFound
	}
	claims, err := auth.GetUserClaims(jwtToken)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (s GophermartService) RegisterUser(ctx context.Context, login string, password string) (*entity.Session, error) {
	user, err := s.userService.Register(ctx, login, password)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionService.NewSession(ctx, user.UID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s GophermartService) LoginUser(ctx context.Context, login string, password string) (*entity.Session, error) {
	user, err := s.userService.Login(ctx, login, password)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionService.NewSession(ctx, user.UID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s GophermartService) GetUserBalance(ctx context.Context, userID string) (float32, float32, error) {
	balance, err := s.OrderService.GetUserAccrual(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	var withdrawalsSum float32
	withdrawals, err := s.WithdrawalService.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	for _, withdrawal := range withdrawals {
		withdrawalsSum += withdrawal.Sum
	}
	return balance - withdrawalsSum, withdrawalsSum, nil
}

//nolint:funlen
func (s GophermartService) GetAccruals(ctx context.Context, orderID string) {
	order, err := s.OrderService.GetOrder(ctx, orderID)
	if err != nil {
		log.Err(err).Str("orderID", orderID).Msg("order not found")
		return
	}

	logError := func(err error) {
		if err != nil {
			log.Err(err).Str("orderID", orderID).Int("retryCount", order.RetryCount).Msg("error during getAccrual occurred")
		}
	}

	if order.RetryCount > 5 {
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Msg("GetAccruals retry limit")
		err = s.OrderService.SetOrderStatus(ctx, orderID, entity.OrderStatusTooManyRetries, 0)
		logError(err)
		return
	}

	var resp *accrualclient.GetAccrualResponse
	resultCh := s.accrualClient.GetAccrual(ctx, orderID)
	select {
	case <-ctx.Done():
		return
	case resp = <-resultCh:
	}

	next := s.calcNext(resp)
	if err := resp.Err; err != nil {
		log.Err(err).Str("orderID", orderID).Msg("error during GetAccrual")
		if errors.Is(err, accrualclient.ErrFatalError) {
			return
		}
	}

	switch resp.Status {
	case entity.OrderStatusPROCESSED, entity.OrderStatusINVALID:
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Float32("accrual", resp.Accrual).Msg("GetAccruals completed")
		err = s.OrderService.SetOrderStatus(ctx, orderID, resp.Status, resp.Accrual)
		logError(err)
		return
	default:
		if resp.Err == nil {
			err = s.OrderService.SetOrderStatus(ctx, orderID, resp.Status, resp.Accrual)
			logError(err)
		}
	}

	err = s.OrderService.ReScheduleOrderProcessingTask(ctx, orderID, time.Now().Add(next))
	logError(err)

	select {
	case <-ctx.Done():
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Msg("GetAccruals context done")
		return
	case <-time.After(next):
		go s.GetAccruals(ctx, orderID) // решедуллим
		return
	}
}

// calcNext вычисляет через сколько надо повторить запрос
func (s GophermartService) calcNext(resp *accrualclient.GetAccrualResponse) time.Duration {
	next := s.accrualRetryInterval
	if err := resp.Err; err != nil {
		var errTooManyRequests accrualclient.TooManyRequestsError
		if errors.As(err, &errTooManyRequests) {
			next = time.Duration(errTooManyRequests.RetryAfterSec) * time.Second
		}
	}
	return next
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
