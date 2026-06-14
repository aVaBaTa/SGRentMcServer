package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
)

// GameCfg = override des planchers de ressources d'un jeu (RAM/CPU), éditable depuis
// /admin. Stocké dans Redis (clé gameCfgKey). Absent → on utilise les valeurs codées
// dans games.go (GameDef.MinRAMMb / MinCPUCores). Permet de changer la config de base
// d'un jeu (ex. Hytale 10 Go / 4 cœurs) sans rebuild.
type GameCfg struct {
	MinRAMMb    int64   `json:"min_ram_mb"`
	MinCPUCores float64 `json:"min_cpu_cores"`
}

func gameCfgKey(id string) string { return "gamecfg:" + id }

// getGameCfg lit l'override Redis d'un jeu (ok=false si aucun).
func (s *Server) getGameCfg(ctx context.Context, gameID string) (GameCfg, bool) {
	var c GameCfg
	if s.rdb == nil {
		return c, false
	}
	val, err := s.rdb.Get(ctx, gameCfgKey(gameID)).Result()
	if err != nil || val == "" {
		return c, false
	}
	if json.Unmarshal([]byte(val), &c) != nil {
		return c, false
	}
	return c, true
}

// effectiveFloor : planchers à appliquer (override Redis si présent, sinon games.go).
func (s *Server) effectiveFloor(ctx context.Context, def servers.GameDef) (int64, float64) {
	ram, cpu := def.MinRAMMb, def.MinCPUCores
	if c, ok := s.getGameCfg(ctx, def.ID); ok {
		if c.MinRAMMb > 0 {
			ram = c.MinRAMMb
		}
		if c.MinCPUCores > 0 {
			cpu = c.MinCPUCores
		}
	}
	return ram, cpu
}

func (s *Server) setGameCfg(ctx context.Context, gameID string, c GameCfg) error {
	b, _ := json.Marshal(c)
	return s.rdb.Set(ctx, gameCfgKey(gameID), b, 0).Err()
}

type gameCfgView struct {
	ID           string  `json:"id"`
	MinRAMMb     int64   `json:"min_ram_mb"`        // effectif (override ou défaut)
	MinCPUCores  float64 `json:"min_cpu_cores"`     // effectif
	DefaultRAMMb int64   `json:"default_ram_mb"`    // valeur codée (games.go)
	DefaultCPU   float64 `json:"default_cpu_cores"` // valeur codée
	Overridden   bool    `json:"overridden"`        // true si valeur custom en Redis
	UsesMCRouter bool    `json:"uses_mc_router"`
}

// handleAdminListGames : GET /api/v1/admin/games → config de base (planchers) par jeu,
// valeur effective + défaut codé. Sert l'éditeur de configs dans /admin.
func (s *Server) handleAdminListGames(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var out []gameCfgView
	for _, id := range servers.GameIDs() {
		def, err := servers.GetGame(id)
		if err != nil {
			continue
		}
		ram, cpu := s.effectiveFloor(r.Context(), def)
		_, ok := s.getGameCfg(r.Context(), id)
		out = append(out, gameCfgView{
			ID: id, MinRAMMb: ram, MinCPUCores: cpu,
			DefaultRAMMb: def.MinRAMMb, DefaultCPU: def.MinCPUCores,
			Overridden: ok, UsesMCRouter: def.UsesMCRouter,
		})
	}
	respond(w, http.StatusOK, out)
}

// handleAdminSetGameConfig : POST /api/v1/admin/games/{id}/config {min_ram_mb, min_cpu_cores}
// → override les planchers d'un jeu (s'applique aux NOUVEAUX serveurs de ce jeu).
func (s *Server) handleAdminSetGameConfig(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if _, err := servers.GetGame(id); err != nil {
		http.Error(w, "jeu inconnu", http.StatusBadRequest)
		return
	}
	var c GameCfg
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if c.MinRAMMb < 256 || c.MinRAMMb > 131072 || c.MinCPUCores < 0.5 || c.MinCPUCores > 64 {
		http.Error(w, "RAM (256–131072 Mo) ou CPU (0.5–64) hors bornes", http.StatusBadRequest)
		return
	}
	if err := s.setGameCfg(r.Context(), id, c); err != nil {
		http.Error(w, "redis error", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]any{"id": id, "min_ram_mb": c.MinRAMMb, "min_cpu_cores": c.MinCPUCores})
}
