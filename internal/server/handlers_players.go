package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// validPlayerName : pseudo Minecraft (Java) — 3 à 16 caractères alphanumériques/underscore.
var validPlayerName = regexp.MustCompile(`^[A-Za-z0-9_]{1,16}$`)

// playerActions mappe une action de l'UI vers la/les commande(s) RCON.
// %s est remplacé par le pseudo (déjà validé).
var playerActions = map[string]string{
	"op":               "op %s",
	"deop":             "deop %s",
	"kick":             "kick %s",
	"ban":              "ban %s",
	"pardon":           "pardon %s",
	"whitelist_add":    "whitelist add %s",
	"whitelist_remove": "whitelist remove %s",
}

// handlePlayerAction : POST /servers/{id}/players/action  {action, player}
func (s *Server) handlePlayerAction(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var body struct {
		Action string `json:"action"`
		Player string `json:"player"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	tmpl, ok := playerActions[body.Action]
	if !ok {
		http.Error(w, "action inconnue", http.StatusBadRequest)
		return
	}
	player := strings.TrimSpace(body.Player)
	if !validPlayerName.MatchString(player) {
		http.Error(w, "pseudo invalide", http.StatusBadRequest)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	cmd := fmt.Sprintf(tmpl, player)
	out, err := node.Exec(r.Context(), gs.ContainerID, append([]string{"rcon-cli"}, strings.Fields(cmd)...))
	if err != nil {
		http.Error(w, "command failed", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]string{"output": strings.TrimSpace(out)})
}

// listAfterColonRe : "There are 2 whitelisted players: Alice, Bob" → capture "Alice, Bob"
var listAfterColonRe = regexp.MustCompile(`:\s*(.*)$`)

func parseNameList(out string) []string {
	out = strings.TrimSpace(out)
	m := listAfterColonRe.FindStringSubmatch(out)
	names := []string{}
	if m == nil {
		return names
	}
	for _, p := range strings.Split(m[1], ",") {
		if p = strings.TrimSpace(strings.TrimSuffix(p, ".")); p != "" {
			names = append(names, p)
		}
	}
	return names
}

// handlePlayerLists : GET /servers/{id}/playerlists → whitelist + bannis.
func (s *Server) handlePlayerLists(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	resp := map[string]any{"whitelist": []string{}, "banned": []string{}}
	if gs.ContainerID == "" || gs.Status != "running" {
		respond(w, http.StatusOK, resp)
		return
	}
	if out, err := node.Exec(r.Context(), gs.ContainerID, []string{"rcon-cli", "whitelist", "list"}); err == nil {
		resp["whitelist"] = parseNameList(out)
	}
	if out, err := node.Exec(r.Context(), gs.ContainerID, []string{"rcon-cli", "banlist", "players"}); err == nil {
		resp["banned"] = parseNameList(out)
	}
	respond(w, http.StatusOK, resp)
}
