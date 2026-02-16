package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sales-tracker/internal/config"
	analytics_handler "sales-tracker/internal/http-server/handler/analytics"
	items_handler "sales-tracker/internal/http-server/handler/items"
	"sales-tracker/internal/http-server/router"
	analytics_postgres "sales-tracker/internal/repository/analytics/postgres"
	items_postgres "sales-tracker/internal/repository/items/postgres"
	analytics_usecase "sales-tracker/internal/usecase/analytics"
	items_usecase "sales-tracker/internal/usecase/items"
	"syscall"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	cfg    *config.Config
	logger *zlog.Zerolog
	server *http.Server
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()
	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}
	db, err := dbpg.New(cfg.DBDSN(), []string{}, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	itemsRepo := items_postgres.NewPostgresRepository(db, retries)
	analyticsRepo := analytics_postgres.NewAnalyticsPostgresRepository(db, retries)

	itemsUsecase := items_usecase.NewService(itemsRepo, logger)
	analyticsUsecase := analytics_usecase.NewService(analyticsRepo, logger)

	itemsHandler := items_handler.NewHandler(itemsUsecase, analyticsUsecase, logger)
	analyticsHandler := analytics_handler.NewHandler(analyticsUsecase, logger)

	mux := router.NewRouter(itemsHandler, analyticsHandler, logger)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		a.logger.Info().Msg("Shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		a.logger.Info().Msg("Server shutdown complete")
		return nil
	case err := <-errCh:
		return err
	}
}
