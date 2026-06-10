package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// playersListRe capture la sortie de `rcon-cli list` :
// "There are 2 of a max of 5 players online: Alice, Bob"
var playersListRe = regexp.MustCompile(`There are (\d+) of a max of (\d+) players online:?\s*(.*)`)

type playersResp struct {
	Online  int      `json:"online"`
	Max     int      `json:"max"`
	Players []string `json:"players"`
}

// parsePlayers transforme la sortie de `rcon-cli list` en structure.
func parsePlayers(out string) playersResp {
	out = strings.TrimSpace(out)
	m := playersListRe.FindStringSubmatch(out)
	if m == nil {
		return playersResp{Players: []string{}}
	}
	online, _ := strconv.Atoi(m[1])
	max, _ := strconv.Atoi(m[2])
	var names []string
	if rest := strings.TrimSpace(m[3]); rest != "" {
		for _, p := range strings.Split(rest, ",") {
			if p = strings.TrimSpace(p); p != "" {
				names = append(names, p)
			}
		}
	}
	if names == nil {
		names = []string{}
	}
	return playersResp{Online: online, Max: max, Players: names}
}

// handleServerPlayers : GET /servers/{id}/players → joueurs connectés (via RCON).
func (s *Server) handleServerPlayers(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		respond(w, http.StatusOK, playersResp{Players: []string{}})
		return
	}
	out, err := node.Exec(r.Context(), gs.ContainerID, []string{"rcon-cli", "list"})
	if err != nil {
		// Serveur pas encore prêt (RCON down) : 0 joueur plutôt qu'une erreur.
		respond(w, http.StatusOK, playersResp{Players: []string{}})
		return
	}
	respond(w, http.StatusOK, parsePlayers(out))
}

// handleServerLogs : GET /servers/{id}/logs?tail=N → console lecture seule.
func (s *Server) handleServerLogs(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	tail := 200
	if v := r.URL.Query().Get("tail"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			tail = n
		}
	}
	if gs.ContainerID == "" {
		respond(w, http.StatusOK, map[string]string{"logs": ""})
		return
	}
	logs, err := node.Logs(r.Context(), gs.ContainerID, tail)
	if err != nil {
		http.Error(w, "logs unavailable", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]string{"logs": logs})
}

// handleServerCommand : POST /servers/{id}/command {command} → exécute via RCON.
func (s *Server) handleServerCommand(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	cmd := strings.TrimSpace(body.Command)
	if cmd == "" {
		http.Error(w, "empty command", http.StatusBadRequest)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	// On retire un éventuel '/' initial (les commandes RCON n'en prennent pas).
	cmd = strings.TrimPrefix(cmd, "/")
	out, err := node.Exec(r.Context(), gs.ContainerID, []string{"rcon-cli", cmd})
	if err != nil {
		http.Error(w, "command failed", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]string{"output": out})
}
