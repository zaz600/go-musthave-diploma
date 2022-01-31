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
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
)

const sessionCookieName = "GM_LS_SESSION"

type key int

const (
	sessionKey key = iota
)

var _ Gophermart.ServerInterface = &GophermartController{}

type GophermartController struct {
	userService    userservice.UserService
	sessionService sessionservice.SessionService
	orderService   orderservice.OrderService
}

func (c *GophermartController) CreateAuthCookie(ctx context.Context, userEntity *entity.UserEntity) (*http.Cookie, error) {
	// TODO в сервис унести и криптовать/подписывать куку
	session, err := c.sessionService.NewSession(ctx, userEntity.UID)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:  sessionCookieName,
		Value: session.SessionID, // TODO подписать
	}, nil
}

func (c GophermartController) PostApiUserRegister(w http.ResponseWriter, r *http.Request) { //nolint:revive
	var request Gophermart.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Login == "" || request.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := c.userService.Register(context.TODO(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrUserExists) {
			http.Error(w, "login already in use", http.StatusConflict)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	cookie, err := c.CreateAuthCookie(context.TODO(), user)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}

func (c GophermartController) PostApiUserLogin(w http.ResponseWriter, r *http.Request) { //nolint:revive
	var request Gophermart.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Login == "" || request.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := c.userService.Login(context.TODO(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrAuth) {
			http.Error(w, "invalid login/password", http.StatusUnauthorized)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cookie, err := c.CreateAuthCookie(context.TODO(), user)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}

func (c GophermartController) PostApiUserOrders(w http.ResponseWriter, r *http.Request) { //nolint:revive
	session, ok := r.Context().Value(sessionKey).(*entity.Session)
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

	log.Info().Str("uid", session.UID).Str("orderID", orderID).Msg("")
	err = c.orderService.UploadOrder(context.TODO(), session.UID, orderID)
	if err != nil {
		if errors.Is(err, orderservice.ErrOrderExists) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, orderservice.ErrOrderOwnedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, orderservice.ErrInvalidOrderFormat) {
			http.Error(w, "invalid orderID format, check luhn mismatch", http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func NewRouter(
	userService userservice.UserService,
	sessionService sessionservice.SessionService,
	orderService orderservice.OrderService,
) *chi.Mux {
	c := &GophermartController{
		userService:    userService,
		sessionService: sessionService,
		orderService:   orderService,
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
		for _, cookie := range r.Cookies() {
			if cookie.Name == sessionCookieName {
				sessionID := cookie.Value
				session, err := c.sessionService.Get(context.TODO(), sessionID)
				if err == nil {
					ctx := context.WithValue(r.Context(), sessionKey, session)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				break
			}
		}

		next.ServeHTTP(w, r)
	})
}
