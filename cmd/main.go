package main

import (
	"avito/internal/config"
	"avito/internal/handlers"
	"avito/internal/manager"
	"avito/internal/metrics"
	"avito/internal/storage"
	"avito/internal/storage/migration"
	"avito/internal/storage/transactionManager"
	"avito/pkg/auth"
	"avito/pkg/logger"
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	err = logger.InitLogger()
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}

	transactionManager := transactionManager.NewTransactionManager(*cfg)

	tm := auth.NewTokenManager(cfg.JWTSecretKey, cfg.TokenExp)

	db := storage.NewStorage(transactionManager)
	defer db.Tm.DB.Close()

	err = migration.GooseUp(db.Tm.DB)
	if err != nil {
		logger.Log.Fatal("migration error: %v", zap.Error(err))
	}

	pm := metrics.Init()

	mgr := manager.NewManager(db, tm, pm)

	h := handlers.Routes(mgr)

	muxMetrics := http.NewServeMux()
	muxMetrics.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:    cfg.PrometheusAddress,
		Handler: muxMetrics,
	}
	go func() {
		if err = metricsSrv.ListenAndServe(); err != nil {
			logger.Log.Error("metrics ListenAndServe error", zap.Error(err))
		}
	}()

	srv := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: h,
	}
	done := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)
		<-sigs
		if err = srv.Shutdown(context.Background()); err != nil {
			logger.Log.Info("HTTP server Shutdown: %v", zap.Error(err))
		}
		close(done)
	}()

	err = srv.ListenAndServe()
	if err != nil {
		logger.Log.Info("HTTP server ListenAndServe error", zap.Error(err))
	}
}
