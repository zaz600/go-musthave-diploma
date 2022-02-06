package gophermartservice

import (
	"context"
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

//nolint:funlen
func (s GophermartService) GetAccruals(ctx context.Context, orderID string, retryNum int) {
	// TODO переписать без рекурсии))
	if retryNum > 5 {
		log.Info().Int("retryNum", retryNum).Msg("GetAccruals retry limit")
		return
	}

	select {
	case <-ctx.Done():
		log.Info().Int("retryNum", retryNum).Msg("GetAccruals context done")
		return
	default:
		resp, err := s.accrualClient.GetOrderAccrualWithResponse(ctx, Accrual.Order(orderID))
		if err != nil {
			log.Err(err).Str("orderID", orderID).Int("retryNum", retryNum).Msg("GetOrderAccrualWithResponse error")
			time.Sleep(1 * time.Second)
			go s.GetAccruals(ctx, orderID, retryNum+1)
			return
		}
		if resp.StatusCode() == http.StatusTooManyRequests {
			retryAfter := resp.HTTPResponse.Header.Get("Retry-After")
			retryAfterSec, err := strconv.Atoi(retryAfter)
			if err != nil {
				retryAfterSec = 5
			}
			log.Info().Str("orderID", orderID).Int("retryNum", retryNum).Int("retryAfterSec", retryAfterSec).Msg("StatusTooManyRequests")
			time.Sleep(time.Duration(retryAfterSec) * time.Second)
			go s.GetAccruals(ctx, orderID, retryNum+1)
			return
		}

		if resp.StatusCode() == 200 {
			switch resp.JSON200.Status {
			case Accrual.ResponseStatusINVALID:
				// TODO err
				err = s.OrderService.SetOrderStatus(ctx, orderID, entity.OrderStatusINVALID, 0)
				if err != nil {
					log.Err(err).Str("orderID", orderID).Int("retryNum", retryNum).Msg("SetOrderStatus error")
				}
				return
			case Accrual.ResponseStatusPROCESSED:
				// TODO err
				err = s.OrderService.SetOrderStatus(ctx, orderID, entity.OrderStatusPROCESSED, *resp.JSON200.Accrual)
				if err != nil {
					log.Err(err).Str("orderID", orderID).Int("retryNum", retryNum).Float32("accrual", *resp.JSON200.Accrual).Msg("SetOrderStatus error")
					return
				}
				log.Info().Str("orderID", orderID).Int("retryNum", retryNum).Float32("accrual", *resp.JSON200.Accrual).Msg("GetAccruals processed")
				return
			case Accrual.ResponseStatusREGISTERED:
				err = s.OrderService.SetOrderStatus(ctx, orderID, entity.OrderStatusPROCESSING, 0.0)
				if err != nil {
					log.Err(err).Str("orderID", orderID).Int("retryNum", retryNum).Msg("SetOrderStatus error")
					return
				}
				log.Info().Str("orderID", orderID).Int("retryNum", retryNum).Float32("accrual", *resp.JSON200.Accrual).Msg("GetAccruals request registered")
			}
			// будем повторять на других статусах
			log.Info().Str("orderID", orderID).Int("retryNum", retryNum).Str("accrual status", string(resp.JSON200.Status)).Msg("GetAccruals in progress")
			time.Sleep(50 * time.Millisecond) // TODO конфиг
			go s.GetAccruals(ctx, orderID, retryNum+1)
			return
		}
	}
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
