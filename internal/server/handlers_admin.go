package server

import (
	"encoding/json"
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
)

// adminAuthorized vérifie le secret partagé entre le monitor (/admin) et l'API.
// Ces endpoints sont internes (réseau Docker) ; ils ne passent PAS par le JWT
// utilisateur — c'est l'admin (derrière Basic Auth nginx) qui agit.
func (s *Server) adminAuthorized(r *http.Request) bool {
	tok := s.cfg.AdminToken
	return tok != "" && r.Header.Get("X-Admin-Token") == tok
}

// handleAdminCatalog : liste des jeux et plans disponibles (pour le formulaire
// de création de serveur dans /admin).
func (s *Server) handleAdminCatalog(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	type planInfo struct {
		Name       string  `json:"name"`
		RAMMb      int64   `json:"ram_mb"`
		CPUCores   float64 `json:"cpu_cores"`
		PriceCents int     `json:"price_cents"`
		Free       bool    `json:"free"`
	}
	var pl []planInfo
	for _, name := range servers.PlanNames() {
		if p, err := servers.GetPlan(name); err == nil {
			pl = append(pl, planInfo{p.Name, p.RAMMb, p.CPUCores, p.PriceCents, p.Free})
		}
	}
	respond(w, http.StatusOK, map[string]any{
		"games": servers.GameIDs(),
		"plans": pl,
	})
}

// handleAdminSetUnlimited : active/désactive le droit de création illimité d'un user.
func (s *Server) handleAdminSetUnlimited(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if _, err := s.userRepo.GetByID(r.Context(), id); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err := s.userRepo.SetUnlimitedCreate(r.Context(), id, body.Enabled); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"id": id, "unlimited_create": body.Enabled})
}

// handleAdminUpdateResources : l'admin change la RAM/CPU d'un serveur (override).
// Applique à chaud sur le container (si présent) via l'orchestrateur + persiste en DB.
func (s *Server) handleAdminUpdateResources(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		RAMMb    int64   `json:"ram_mb"`
		CPUCores float64 `json:"cpu_cores"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// Bornes de sécurité (évite 0 / valeurs absurdes qui tueraient le container).
	if body.RAMMb < 512 || body.RAMMb > 131072 || body.CPUCores < 0.5 || body.CPUCores > 64 {
		http.Error(w, "ram_mb (512–131072) ou cpu_cores (0.5–64) hors bornes", http.StatusBadRequest)
		return
	}
	gs, err := s.serverRepo.GetByIDAny(r.Context(), id)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	// Applique à chaud sur le container si présent (sinon : DB seulement, prend effet au prochain start).
	if gs.ContainerID != "" {
		node, err := s.orch.NodeByID(gs.Node)
		if err != nil {
			http.Error(w, "node unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := node.UpdateResources(r.Context(), gs.ContainerID, body.RAMMb, body.CPUCores); err != nil {
			http.Error(w, "docker update failed: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
	}
	if err := s.serverRepo.UpdateResources(r.Context(), id, body.RAMMb, body.CPUCores); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"id": id, "ram_mb": body.RAMMb, "cpu_cores": body.CPUCores})
}

// handleAdminCreateServer : l'admin crée un serveur AU NOM d'un utilisateur, sur
// n'importe quel plan, sans paiement (autorité admin).
func (s *Server) handleAdminCreateServer(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		UserID string `json:"user_id"`
		createServerRequest
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	user, err := s.userRepo.GetByID(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	// allowPaid = true : l'admin peut attribuer n'importe quel plan gratuitement.
	gs, status, err := s.createServerForUser(r.Context(), user.ID, user.Username, req.createServerRequest, true)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	respond(w, http.StatusCreated, gs)
}
