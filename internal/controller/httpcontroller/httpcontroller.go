package httpcontroller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	Gophermart "github.com/zaz600/go-musthave-diploma/api"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/auth"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/luhn"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
)

type key int

const (
	userIDKey key = iota
)

var _ Gophermart.ServerInterface = &GophermartController{}

type GophermartController struct {
	gophermartService *gophermartservice.GophermartService
}

func (c GophermartController) UserRegister(w http.ResponseWriter, r *http.Request) { //nolint:revive
	var request Gophermart.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Login == "" || request.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	session, err := c.gophermartService.RegisterUser(r.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, gophermartservice.ErrUserExists) {
			http.Error(w, "login already in use", http.StatusConflict)
			return
		}
		log.Err(err).Msg("user register error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = auth.SetJWT(w, session)
	if err != nil {
		log.Err(err).Msg("user register error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}

func (c GophermartController) UserLogin(w http.ResponseWriter, r *http.Request) { //nolint:revive
	var request Gophermart.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Login == "" || request.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	session, err := c.gophermartService.LoginUser(r.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, gophermartservice.ErrAuth) {
			http.Error(w, "invalid login/password", http.StatusUnauthorized)
			return
		}
		log.Err(err).Msg("user login error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = auth.SetJWT(w, session)
	if err != nil {
		log.Err(err).Msg("user login error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}

func (c GophermartController) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil || len(bytes) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	orderID := string(bytes)

	log.Info().Str("uid", userID).Str("orderID", orderID).Msg("UploadOrder")
	err = c.gophermartService.UploadOrder(r.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, gophermartservice.ErrOrderExists) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, gophermartservice.ErrOrderOwnedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, gophermartservice.ErrInvalidOrderFormat) {
			http.Error(w, "invalid orderID format, check luhn mismatch", http.StatusUnprocessableEntity)
			return
		}
		log.Err(err).Msg("upload order error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	go c.gophermartService.GetAccruals(orderID)

	w.Header().Set("Content-Type", "application/ json")
	w.WriteHeader(http.StatusAccepted)
}

func (c *GophermartController) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	orders, err := c.gophermartService.GetUserOrders(r.Context(), userID)
	if err != nil {
		log.Err(err).Msg("get user orders error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}

	var resp = Gophermart.OrdersResponse{}
	for _, order := range orders {
		respOrder := Gophermart.Order{
			Number:     order.OrderID,
			Status:     Gophermart.OrderStatus(order.Status),
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		if order.Status == entity.OrderStatusProcessed {
			respOrder.Accrual = &order.Accrual
		}
		resp = append(resp, respOrder)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Err(err).Msg("get user orders error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bytes)
}

func (c *GophermartController) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	currentBalance, withdrawalsSum, err := c.gophermartService.GetUserBalance(r.Context(), userID)
	if err != nil {
		log.Err(err).Msg("get user balance error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	balance := Gophermart.UserBalanceResponse{
		Current:   Gophermart.Amount(currentBalance),
		Withdrawn: Gophermart.Amount(withdrawalsSum),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(balance)
}

func (c *GophermartController) UserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var request Gophermart.UserBalanceWithdrawRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Order == "" || request.Sum <= 0.0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if ok := luhn.CheckLuhn(request.Order); !ok {
		http.Error(w, "invalid orderID format, check luhn mismatch", http.StatusUnprocessableEntity)
		return
	}

	currentBalance, _, err := c.gophermartService.GetUserBalance(r.Context(), userID)
	if err != nil {
		log.Err(err).Msg("user balance withdraw error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if currentBalance < float32(request.Sum) {
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
		return
	}

	err = c.gophermartService.UploadWithdrawal(r.Context(), userID, request.Order, float32(request.Sum))
	if err != nil {
		log.Err(err).Msg("user balance withdraw error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (c *GophermartController) UserBalanceWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := c.gophermartService.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		log.Err(err).Msg("user balance withdrawals error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}

	var resp Gophermart.UserBalanceWithdrawalsResponse
	for _, withdrawal := range withdrawals {
		respWithdrawal := Gophermart.UserBalanceWithdrawal{
			Order:       withdrawal.OrderID,
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
			Sum:         Gophermart.Amount(withdrawal.Sum),
		}
		resp = append(resp, respWithdrawal)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Err(err).Msg("user balance withdrawals error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bytes)
}

func NewRouter(gophermartService *gophermartservice.GophermartService) *chi.Mux {
	c := &GophermartController{
		gophermartService: gophermartService,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.Compress(5))
	// TODO gzip

	r.Route("/", func(r chi.Router) {
		r.Use(c.AuthCtx)
		r.Mount("/", Gophermart.Handler(c))
	})

	return r
}

func (c GophermartController) AuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.GetClaims(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
