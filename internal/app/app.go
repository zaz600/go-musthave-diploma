package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	Accrual "github.com/zaz600/go-musthave-diploma/api/accrual"
	"github.com/zaz600/go-musthave-diploma/internal/app/config"
	"github.com/zaz600/go-musthave-diploma/internal/controller/httpcontroller"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/migration"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/httpserver"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/logger"
	"github.com/zaz600/go-musthave-diploma/internal/service/gophermartservice"
)

func Run(args []string) error {
	ctxBg := context.Background()
	ctx, cancel := signal.NotifyContext(ctxBg, os.Interrupt, syscall.SIGINT)
	defer cancel()

	l := logger.New()
	cfg := config.Config(args)
	l.Info().
		Str("addr", cfg.ServerAddress).
		Str("db", cfg.DatabaseDSN).
		Str("accrual", cfg.AccrualAddress).
		Msg("config")

	accrualClient, err := Accrual.NewClientWithResponses(cfg.AccrualAddress)
	if err != nil {
		return err
	}

	var db *sql.DB
	var service *gophermartservice.GophermartService
	switch cfg.RepositoryType() {
	case config.MemoryRepo:
		service, err = gophermartservice.New(accrualClient, gophermartservice.WithMemoryStorage())
		if err != nil {
			return err
		}
	case config.DatabaseRepo:
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return err
		}

		err = migration.Migrate(db)
		if err != nil {
			return err
		}
		service, err = gophermartservice.New(accrualClient, gophermartservice.WithPgStorage(db))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown repo type")
	}

	server := httpserver.New(httpcontroller.NewRouter(service), httpserver.WithAddr(cfg.ServerAddress))

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutdown...")
		service.Shutdown()
		if db != nil {
			_ = db.Close()
		}
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
