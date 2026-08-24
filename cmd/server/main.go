package main

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/audit"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/config"
	"github.com/11DingKing/robot-athlete-village/internal/httpapi"
	"github.com/11DingKing/robot-athlete-village/internal/middleware"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/service"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"github.com/11DingKing/robot-athlete-village/internal/worker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}
	cfg := config.Load()
	ctx := context.Background()
	db, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	store := repository.NewSQLite(db)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	aud := audit.New(store)
	village := service.NewVillage(store, aud)
	report := service.NewReport(db)
	as := auth.New(store, cfg.SessionTTL)
	api := httpapi.New(as, village, report, store)
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go worker.NewMaintenance(store, cfg.WorkerInterval, logger).Run(workerCtx)
	handler := middleware.RequestID(middleware.Recovery(logger, api.Routes()))
	srv := &http.Server{Addr: cfg.Addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("server_started", "addr", cfg.Addr)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			logger.Error("server_failed", "error", e)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown_failed", "error", err)
	}
}
func runHealthcheck() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("healthy")
}
