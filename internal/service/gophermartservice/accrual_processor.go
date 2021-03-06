package gophermartservice

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/providers/accrual"
)

//nolint:funlen
func (s GophermartService) GetAccruals(orderID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	order, err := s.repo.OrderRepo.GetOrder(ctx, orderID)
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
		err = s.repo.OrderRepo.SetOrderStatusAndAccrual(ctx, orderID, entity.OrderStatusTooManyRetries, 0)
		logError(err)
		return
	}

	if order.Status == entity.OrderStatusProcessed || order.Status == entity.OrderStatusInvalid {
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Msg("GetAccruals finnish order status")
		return
	}

	var resp *accrual.GetAccrualResponse
	resultCh := s.accrualProvider.GetAccrual(ctx, orderID)
	select {
	case <-ctx.Done():
		return
	case resp = <-resultCh:
	}

	next := s.calcNext(resp)
	if err := resp.Err; err != nil {
		log.Err(err).Str("orderID", orderID).Msg("error during GetAccrual")
		if errors.Is(err, accrual.ErrFatalError) {
			return
		}
	}

	switch resp.Status {
	case Accrual.ResponseStatusPROCESSED:
		accrualAmount := *resp.Accrual
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Str("status", string(resp.Status)).Float32("accrual", accrualAmount).Msg("GetAccruals completed")
		err = s.repo.OrderRepo.SetOrderStatusAndAccrual(ctx, orderID, entity.OrderStatusProcessed, accrualAmount)
		logError(err)
		if *resp.Accrual > 0 {
			// TODO баланс изменять надо в одной транзакции с регистрацией зачисления
			err = s.repo.AccountRepo.RefillAmount(ctx, order.UID, accrualAmount)
			logError(err)
		}
		return

	case Accrual.ResponseStatusINVALID:
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Str("status", string(resp.Status)).Msg("GetAccruals completed")
		err = s.repo.OrderRepo.SetOrderStatusAndAccrual(ctx, orderID, entity.OrderStatusInvalid, 0)
		logError(err)
		return
	case Accrual.ResponseStatusPROCESSING, Accrual.ResponseStatusREGISTERED:
		if resp.Err == nil {
			err = s.repo.OrderRepo.SetOrderStatusAndAccrual(ctx, orderID, entity.OrderStatusProcessing, 0)
			logError(err)
		}

	default:
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Str("status", string(resp.Status)).Msg("unknown accrual status")
	}

	err = s.repo.OrderRepo.SetOrderNextRetryAt(ctx, orderID, time.Now().Add(next))
	logError(err)

	select {
	case <-ctx.Done():
		log.Info().Str("orderID", orderID).Int("retryCount", order.RetryCount).Msg("GetAccruals context done")
		return
	case <-time.After(next):
		go s.GetAccruals(orderID) // решедуллим
		return
	}
}

// calcNext вычисляет через сколько надо повторить запрос
func (s GophermartService) calcNext(resp *accrual.GetAccrualResponse) time.Duration {
	next := s.accrualRetryInterval
	if err := resp.Err; err != nil {
		var errTooManyRequests accrual.TooManyRequestsError
		if errors.As(err, &errTooManyRequests) {
			next = time.Duration(errTooManyRequests.RetryAfterSec) * time.Second
		}
	}
	return next
}
