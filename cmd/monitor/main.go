package main

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/monitor"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	port := getEnv("MONITOR_PORT", "8090")
	hostRoot := getEnv("HOST_ROOT", "/host")
	nodes := getEnv("DOCKER_NODES", "node1=unix:///var/run/docker.sock")

	collector, err := monitor.NewCollector(hostRoot, nodes)
	if err != nil {
		slog.Error("failed to init collector", "err", err)
		os.Exit(1)
	}

	// Connexion DB optionnelle (pour la liste des utilisateurs)
	var db *pgxpool.Pool
	if dbURL := getEnv("DATABASE_URL", ""); dbURL != "" {
		if pool, err := pgxpool.New(context.Background(), dbURL); err == nil {
			db = pool
			defer db.Close()
		} else {
			slog.Warn("db unavailable, users list disabled", "err", err)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "db not configured", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		users, err := monitor.ListUsers(ctx, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []monitor.UserInfo{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()

		snap, err := collector.Collect(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snap)
	})

	mux.HandleFunc("POST /api/containers/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := collector.StopContainer(ctx, r.PathValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /api/containers/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := collector.RemoveContainer(ctx, r.PathValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.Handle("GET /", http.FileServer(http.FS(staticFS)))
	// Sert l'index à la racine
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		b, _ := staticFS.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b)
	})

	slog.Info("monitor starting", "port", port, "host_root", hostRoot)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
