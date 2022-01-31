package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/zaz600/go-musthave-diploma/internal/app/config"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
	"github.com/zaz600/go-musthave-diploma/internal/utils/httpserver"
	"github.com/zaz600/go-musthave-diploma/internal/utils/logger"
)

func Run(args []string) error {
	ctxBg := context.Background()
	ctx, cancel := signal.NotifyContext(ctxBg, os.Interrupt, syscall.SIGINT)
	defer cancel()

	l := logger.New()
	cfg := config.GetConfig(args)
	l.Info().
		Str("addr", cfg.ServerAddress).
		Str("db", cfg.DatabaseDSN).
		Str("accrual", cfg.AccrualAddress).
		Msg("config")

	// TODO выбрать нужный тип репозитория
	userRepo := userrepository.NewInmemoryUserRepository()
	userService := userservice.NewService(userRepo)
	sessionService := sessionservice.NewService(sessionrepository.NewInmemorySessionRepository())
	orderService := orderservice.NewService(orderrepository.NewInmemoryOrderRepository())
	server := httpserver.New(httpcontroller.NewRouter(userService, sessionService, orderService), httpserver.WithAddr(cfg.ServerAddress))

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutdown...")
		ctx, cancel := context.WithTimeout(ctxBg, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Err(err).Msg("error during shutdown server")
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
