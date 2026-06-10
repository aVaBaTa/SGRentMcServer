package server

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/orchestrator"
	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
)

// Format des logs de l'image ghcr.io/terkea/hytale-server (OAuth device-code).
// Les deux étapes (téléchargeur puis serveur) impriment l'URL de vérification
// Hytale ; le code utilisateur est porté par le paramètre user_code.
//
//	Please visit the following URL to authenticate:
//	https://oauth.accounts.hytale.com/oauth2/device/verify?user_code=XXXX
//	...
//	  SERVER AUTHENTICATION REQUIRED
//	  Visit: https://oauth.accounts.hytale.com/oauth2/device/verify?user_code=YYYY
//	  Code:  YYYY
var hytaleVerifyRe = regexp.MustCompile(`https://oauth\.accounts\.hytale\.com/oauth2/device/verify\?user_code=([A-Za-z0-9_-]+)`)

// Marqueurs de fin d'authentification (le serveur démarre réellement ensuite).
var hytaleAuthDoneRe = regexp.MustCompile(`(?i)authenticated and ready|server oauth authorized`)

type authStatus struct {
	Pending bool   `json:"pending"`        // true tant qu'une autorisation est attendue
	Step    int    `json:"step,omitempty"` // 1 = téléchargeur, 2 = serveur
	URL     string `json:"url,omitempty"`  // URL de vérification à visiter
	Code    string `json:"code,omitempty"` // code à entrer
}

// parseAuthFromLogs extrait le dernier prompt OAuth d'un bloc de logs.
// done=true si l'authentification est terminée (plus rien à faire).
func parseAuthFromLogs(logs string) (st authStatus, done bool) {
	if hytaleAuthDoneRe.MatchString(logs) {
		return authStatus{Pending: false}, true
	}
	idxs := hytaleVerifyRe.FindAllStringSubmatchIndex(logs, -1)
	if len(idxs) == 0 {
		return authStatus{Pending: false}, false
	}
	last := idxs[len(idxs)-1]
	st = authStatus{
		Pending: true,
		URL:     logs[last[0]:last[1]],
		Code:    logs[last[2]:last[3]],
		Step:    1,
	}
	// L'étape "serveur" est précédée de la bannière SERVER AUTHENTICATION REQUIRED.
	if strings.Contains(logs[:last[0]], "SERVER AUTHENTICATION REQUIRED") {
		st.Step = 2
	}
	return st, false
}

// watchAuthAndRun surveille les logs d'un serveur à authentification interactive
// (Hytale) après son démarrage : passe le statut à "auth_required" dès qu'un
// prompt OAuth apparaît, puis à "running" une fois l'authentification complétée.
// Détaché du contexte de provisioning (l'utilisateur a besoin de temps).
func (s *Server) watchAuthAndRun(gs *servers.GameServer, node *orchestrator.Node) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	announcedAuth := false
	for {
		select {
		case <-ctx.Done():
			slog.Warn("auth watch: timeout before completion", "id", gs.ID)
			return
		case <-ticker.C:
		}

		logs, err := node.Logs(ctx, gs.ContainerID, 400)
		if err != nil {
			continue
		}
		st, done := parseAuthFromLogs(logs)
		if done {
			s.serverRepo.UpdateStatus(ctx, gs.ID, "running")
			slog.Info("auth watch: authentication complete, server running", "id", gs.ID)
			return
		}
		if st.Pending && !announcedAuth {
			announcedAuth = true
			s.serverRepo.UpdateStatus(ctx, gs.ID, "auth_required")
			slog.Info("auth watch: interactive authorization required", "id", gs.ID, "step", st.Step)
		}
	}
}

// handleServerAuth : GET /servers/{id}/auth → état de l'authentification
// interactive (URL + code à présenter à l'utilisateur). Sans état : relit les
// logs du container à la demande.
func (s *Server) handleServerAuth(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// Jeu sans auth, ou serveur déjà en ligne : rien à faire.
	if gs.ContainerID == "" || gs.Status == "running" {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	if def, err := servers.GetGame(gs.Game); err != nil || !def.NeedsAuth {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	logs, err := node.Logs(r.Context(), gs.ContainerID, 400)
	if err != nil {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	st, _ := parseAuthFromLogs(logs)
	respond(w, http.StatusOK, st)
}
