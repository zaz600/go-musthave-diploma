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
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/app/config"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
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
	accrualClient, err := Accrual.NewClientWithResponses(cfg.AccrualAddress)
	if err != nil {
		return err
	}
	service := gophermartservice.NewWithMemStorage(accrualClient)
	server := httpserver.New(httpcontroller.NewRouter(service), httpserver.WithAddr(cfg.ServerAddress))

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
