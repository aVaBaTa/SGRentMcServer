package server

import (
	"encoding/json"
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
)

// GET /api/v1/servers/{id}/settings — défs éditables du jeu + valeurs actuelles.
// Un jeu sans réglages renvoie defs=[] : le frontend masque alors la carte.
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	gs, _, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defs := servers.SettingsFor(gs.Game)
	if defs == nil {
		defs = []servers.SettingDef{}
	}
	values := gs.Settings
	if values == nil {
		values = map[string]string{}
	}
	respond(w, http.StatusOK, map[string]any{"defs": defs, "values": values})
}

type settingsRequest struct {
	Values map[string]string `json:"values"`
}

// POST /api/v1/servers/{id}/settings — valide, persiste, puis RECRÉE le
// container (l'env est immuable ; le volume — donc le monde — est conservé,
// même mécanique que le changement de version/plan).
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Values == nil {
		http.Error(w, "values is required", http.StatusBadRequest)
		return
	}

	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(servers.SettingsFor(gs.Game)) == 0 {
		http.Error(w, "ce jeu n'a pas de réglages éditables", http.StatusBadRequest)
		return
	}
	if err := servers.ValidateSettings(gs.Game, req.Values); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	plan, err := servers.GetPlan(gs.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.serverRepo.UpdateSettings(r.Context(), gs.ID, req.Values); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	gs.Settings = req.Values

	if gs.ContainerID != "" {
		s.serverRepo.UpdateStatus(r.Context(), gs.ID, "creating")
		go s.recreateServer(gs, node, plan, gs.Version)
		gs.Status = "creating"
	}
	respond(w, http.StatusOK, gs)
}
