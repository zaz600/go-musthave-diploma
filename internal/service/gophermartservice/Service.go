package gophermartservice

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/withdrawalrepository"
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/withdrawalservice"
)

const sessionCookieName = "GM_LS_SESSION"

type GophermartService struct {
	userService       userservice.UserService
	sessionService    sessionservice.SessionService
	OrderService      orderservice.OrderService
	WithdrawalService withdrawalservice.WithdrawalService

	accrualClient Accrual.ClientWithResponsesInterface
}

func (s GophermartService) SetAuthCookie(w http.ResponseWriter, session *entity.Session) error {
	cookie := &http.Cookie{
		Name:  sessionCookieName,
		Value: session.SessionID, // TODO подписать
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s GophermartService) GetSession(r *http.Request) (*entity.Session, error) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == sessionCookieName {
			sessionID := cookie.Value // TODO validate
			return s.sessionService.Get(context.TODO(), sessionID)
		}
	}
	return nil, ErrSessionNotFounf
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
	orders, err := s.OrderService.GetUserOrders(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	var balance float32
	for _, order := range orders {
		if order.Status == entity.OrderStatusPROCESSED {
			balance += order.Accrual
		}
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

func (s GophermartService) GetAccruals(ctx context.Context, orderID string) {
	order, err := s.OrderService.GetOrder(ctx, orderID)
	if err != nil {
		log.Err(err).Str("orderID", orderID).Msg("order not found")
		return
	}

	taskContext := order.Context
	logError := func(err error) {
		if err != nil {
			log.Err(err).Str("orderID", orderID).Int("retryCount", taskContext.RetryCount).Msg("error during getAccrual occurred")
		}
	}

	if taskContext.RetryCount > 5 {
		log.Info().Str("orderID", orderID).Int("retryCount", taskContext.RetryCount).Msg("GetAccruals retry limit")
		err = s.OrderService.SetOrderStatus(ctx, orderID, entity.OrderStatusTooManyRetries, 0)
		logError(err)
		return
	}

	resp := s.getAccrual(ctx, orderID, &taskContext)
	switch resp.status {
	case entity.OrderStatusPROCESSED, entity.OrderStatusINVALID:
		log.Info().Str("orderID", orderID).Int("retryCount", taskContext.RetryCount).Float32("accrual", resp.accrual).Msg("GetAccruals completed")
		err = s.OrderService.SetOrderStatus(ctx, orderID, resp.status, resp.accrual)
		logError(err)
		return
	default:
		err = s.OrderService.SetOrderStatus(ctx, orderID, resp.status, resp.accrual)
		logError(err)
	}

	next := 50 * time.Millisecond
	if resp.retryAfterSec > 0 {
		next = time.Duration(resp.retryAfterSec) * time.Second
	}
	err = s.OrderService.ReScheduleOrderProcessingTask(ctx, orderID, time.Now().Add(next))
	logError(err)

	select {
	case <-ctx.Done():
		log.Info().Str("orderID", orderID).Int("retryCount", taskContext.RetryCount).Msg("GetAccruals context done")
		return
	case <-time.After(next):
		go s.GetAccruals(ctx, orderID) // решедуллим
		return
	}
}

type getAccrualStatus struct {
	status        entity.OrderStatus
	retryAfterSec int
	accrual       float32
}

//nolint:funlen
func (s GophermartService) getAccrual(ctx context.Context, orderID string, taskContext *entity.TaskContext) getAccrualStatus {
	logError := func(err error) {
		if err != nil {
			log.Err(err).Str("orderID", orderID).Int("retryCount", taskContext.RetryCount).Msg("error during getAccrual occurred")
		}
	}

	resp, err := s.accrualClient.GetOrderAccrualWithResponse(ctx, Accrual.Order(orderID))
	if err != nil {
		logError(err)
		return getAccrualStatus{status: entity.OrderStatusPROCESSING}
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfter := resp.HTTPResponse.Header.Get("Retry-After")
		retryAfterSec, err := strconv.Atoi(retryAfter)
		if err != nil {
			retryAfterSec = 5
		}
		return getAccrualStatus{retryAfterSec: retryAfterSec, status: entity.OrderStatusPROCESSING}
	}

	if resp.StatusCode() != 200 {
		logError(fmt.Errorf("unknown http status %d", resp.StatusCode()))
		return getAccrualStatus{status: entity.OrderStatusPROCESSING}
	}

	log.Info().
		Str("orderID", orderID).
		Int("retryCount", taskContext.RetryCount).
		Str("accrualStatus", string(resp.JSON200.Status)).
		Float32("accrual", *resp.JSON200.Accrual).
		Msg("get accrual result")

	switch resp.JSON200.Status {
	case Accrual.ResponseStatusINVALID:
		return getAccrualStatus{accrual: 0.0, status: entity.OrderStatusINVALID}
	case Accrual.ResponseStatusPROCESSED:
		return getAccrualStatus{accrual: *resp.JSON200.Accrual, status: entity.OrderStatusPROCESSED}
	case Accrual.ResponseStatusREGISTERED:
		return getAccrualStatus{status: entity.OrderStatusPROCESSING}
	case Accrual.ResponseStatusPROCESSING:
		return getAccrualStatus{status: entity.OrderStatusPROCESSING}
	}
	logError(fmt.Errorf("unknown accrual status %s", resp.JSON200.Status))
	return getAccrualStatus{status: entity.OrderStatusPROCESSING}
}

func NewWithMemStorage(accrualClient Accrual.ClientWithResponsesInterface) *GophermartService {
	return &GophermartService{
		accrualClient:     accrualClient,
		userService:       userservice.NewService(userrepository.NewInmemoryUserRepository()),
		sessionService:    sessionservice.NewService(sessionrepository.NewInmemorySessionRepository()),
		OrderService:      orderservice.NewService(orderrepository.NewInmemoryOrderRepository()),
		WithdrawalService: withdrawalservice.NewService(withdrawalrepository.NewInmemoryWithdrawalRepository()),
	}
}
