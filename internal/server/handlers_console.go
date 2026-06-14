package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
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
	// On bloque les retours à la ligne (1 commande = 1 ligne ; pas d'injection multi-commande).
	if strings.ContainsAny(cmd, "\r\n") {
		http.Error(w, "invalid command", http.StatusBadRequest)
		return
	}
	// Jeux à console stdin (Hytale) : on écrit directement sur le STDIN du serveur.
	// La sortie apparaît dans les logs (pas de retour direct). Les jeux à RCON
	// (Minecraft) passent par rcon-cli et renvoient la sortie.
	if def, derr := servers.GetGame(gs.Game); derr == nil && def.ConsoleStdin {
		if err := node.SendStdin(r.Context(), gs.ContainerID, cmd); err != nil {
			http.Error(w, "command failed", http.StatusServiceUnavailable)
			return
		}
		respond(w, http.StatusOK, map[string]string{"output": "", "sent": "true"})
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

// discoveryTokenRe : caractères autorisés dans un token de découverte Hytale
// (opaque : base64url + points + '='). Empêche toute injection dans la ligne console.
var discoveryTokenRe = regexp.MustCompile(`^[A-Za-z0-9._=\-]{8,512}$`)

// handleServerDiscovery : POST /servers/{id}/discovery {token} → lie le serveur au
// listing public Hytale via `discovery link <token>` (console stdin). Hytale only.
func (s *Server) handleServerDiscovery(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	def, derr := servers.GetGame(gs.Game)
	if derr != nil || !def.ConsoleStdin {
		http.Error(w, "discovery not supported for this game", http.StatusBadRequest)
		return
	}
	var body struct {
		Token  string `json:"token"`
		Unlink bool   `json:"unlink"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	cmd := "discovery unlink"
	if !body.Unlink {
		token := strings.TrimSpace(body.Token)
		if !discoveryTokenRe.MatchString(token) {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		cmd = "discovery link " + token
	}
	if err := node.SendStdin(r.Context(), gs.ContainerID, cmd); err != nil {
		http.Error(w, "command failed", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "sent"})
}
