package main

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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

	// Historique en mémoire : ~1 échantillon / 20 s, gardé sur ~2 h (360 points).
	history := monitor.NewHistory(360)
	go func() {
		tick := time.NewTicker(20 * time.Second)
		defer tick.Stop()
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			if snap, err := collector.Collect(ctx); err == nil {
				history.Record(snap)
			}
			cancel()
			<-tick.C
		}
	}()

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

	// Client admin (proxy vers l'API) : création de serveur pour un user + toggle
	// des droits. nil si API_BASE/ADMIN_TOKEN absents → endpoints renvoient 503.
	admin := monitor.NewAdminClient(getEnv("API_BASE", "http://mcserver-api:8080"), getEnv("ADMIN_TOKEN", ""))

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

	mux.HandleFunc("GET /api/servers", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "db not configured", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		list, err := monitor.ListServers(ctx, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []monitor.ServerInfo{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
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

	mux.HandleFunc("GET /api/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history.Snapshot())
	})

	mux.HandleFunc("GET /api/containers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
		logs, err := collector.ContainerLogs(ctx, r.PathValue("id"), tail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(logs))
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

	mux.HandleFunc("POST /api/containers/{id}/start", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := collector.StartContainer(ctx, r.PathValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /api/containers/{id}/restart", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		if err := collector.RestartContainer(ctx, r.PathValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Suppression complète d'un serveur de jeu : container Docker + ligne DB game_servers.
	// NB : ne libère pas encore volume / route mc-router (cf. #M, à câbler via orchestrateur).
	mux.HandleFunc("POST /api/servers/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "db not configured", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
		defer cancel()
		id := r.PathValue("id")
		cid, _, err := monitor.GetServerContainer(ctx, db, id)
		if err != nil {
			http.Error(w, "serveur introuvable", http.StatusNotFound)
			return
		}
		if cid != "" {
			// Best-effort : le container peut déjà avoir disparu.
			if err := collector.RemoveContainer(ctx, cid); err != nil {
				slog.Warn("delete server: container removal failed", "id", id, "container", cid, "err", err)
			}
		}
		if err := monitor.DeleteServer(ctx, db, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Admin : catalogue, création de serveur pour un user, toggle des droits ---
	relay := func(w http.ResponseWriter, status int, body []byte, err error) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(body)
	}

	mux.HandleFunc("GET /api/catalog", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil {
			http.Error(w, "admin API non configurée", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, body, err := admin.Catalog(ctx)
		relay(w, st, body, err)
	})

	mux.HandleFunc("POST /api/users/{id}/unlimited", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil {
			http.Error(w, "admin API non configurée", http.StatusServiceUnavailable)
			return
		}
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		st, out, err := admin.SetUnlimited(ctx, r.PathValue("id"), body)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/users/{id}/create-server", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil {
			http.Error(w, "admin API non configurée", http.StatusServiceUnavailable)
			return
		}
		var in map[string]any
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		in["user_id"] = r.PathValue("id")
		payload, _ := json.Marshal(in)
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		st, out, err := admin.CreateServer(ctx, payload)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/servers/{id}/resources", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil {
			http.Error(w, "admin API non configurée", http.StatusServiceUnavailable)
			return
		}
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		st, out, err := admin.UpdateResources(ctx, r.PathValue("id"), body)
		relay(w, st, out, err)
	})

	mux.HandleFunc("GET /api/promo", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.GetPromo(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/promo", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.SetPromo(ctx, body)
		relay(w, st, out, err)
	})

	mux.HandleFunc("GET /api/games", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.ListGames(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/games/{id}/config", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.SetGameConfig(ctx, r.PathValue("id"), body)
		relay(w, st, out, err)
	})

	mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.Metrics(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("GET /api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.Feedback(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("GET /api/pageviews", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.PageViews(ctx)
		relay(w, st, out, err)
	})

	// --- Calradia-Coop : accès anticipé (onglet Calradia) ---
	mux.HandleFunc("GET /api/calradia/early-access", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaList(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/calradia/early-access/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<12))
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaDecide(ctx, r.PathValue("id"), body)
		relay(w, st, out, err)
	})

	// Verrou de sortie : date annoncée (informative) + ouverture publique explicite.
	mux.HandleFunc("GET /api/calradia/release", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaGetRelease(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/calradia/release", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<12))
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaSetRelease(ctx, body)
		relay(w, st, out, err)
	})

	// Liste d'accès au téléchargement (utilisateurs Discord connus + ajouts manuels).
	mux.HandleFunc("GET /api/calradia/users", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaUsers(ctx)
		relay(w, st, out, err)
	})

	mux.HandleFunc("POST /api/calradia/allow", func(w http.ResponseWriter, r *http.Request) {
		if admin == nil { http.Error(w, "admin API non configurée", http.StatusServiceUnavailable); return }
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<12))
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		st, out, err := admin.CalradiaAllow(ctx, body)
		relay(w, st, out, err)
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
