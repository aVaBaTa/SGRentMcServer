package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/config"
	"github.com/aVaBaTa/SGRentMcServer/internal/db"
	"github.com/aVaBaTa/SGRentMcServer/internal/orchestrator"
	"github.com/aVaBaTa/SGRentMcServer/internal/server"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	pg, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pg.Close()

	rdb := db.NewRedis(cfg.RedisURL)
	defer rdb.Close()

	orch := orchestrator.New()
	for _, node := range cfg.DockerNodes {
		if err := orch.AddNode(node.ID, node.Host); err != nil {
			slog.Warn("failed to connect to node", "node", node.ID, "err", err)
		} else {
			slog.Info("node connected", "node", node.ID, "host", node.Host)
		}
	}

	srv := server.New(cfg, pg, rdb, orch)

	// Ré-enregistre les routes mc-router des serveurs existants (après un redémarrage)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		srv.ReconcileRoutes(ctx)
	}()

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
}
