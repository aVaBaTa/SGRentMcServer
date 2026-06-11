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

// Parsing tolérant des logs d'auth OAuth (image Hytale type ghcr.io/terkea).
// On ne dépend pas d'un format exact : on cherche une URL de vérification, un
// code, et des marqueurs de fin — avec repli sur les lignes brutes pertinentes
// affichées à l'utilisateur s'il faut faire l'auth à la main.
var (
	anyURLRe        = regexp.MustCompile(`https?://[^\s"'<>]+`)
	userCodeParamRe = regexp.MustCompile(`(?i)user_code=([A-Za-z0-9-]+)`)
	codeLineRe      = regexp.MustCompile(`(?i)\b(?:user[_ ]?code|code)\b\s*[:=]?\s*([A-Z0-9][A-Z0-9-]{3,})`)
	authReadyRe     = regexp.MustCompile(`(?i)(authenticated and ready|server oauth authorized|server (?:is )?(?:started|ready|listening)|ready for connections|world (?:loaded|ready)|listening on)`)
	authHintRe      = regexp.MustCompile(`(?i)(verify|device|oauth|authenticate|authoriz|user_code|sign in|enter the code)`)
	authErrRe       = regexp.MustCompile(`(?i)(authentication failed|oauth (?:error|failed)|download failed|fatal|panic:)`)
)

type authStatus struct {
	Pending bool     `json:"pending"`        // une autorisation est attendue
	Step    int      `json:"step,omitempty"` // 1 = téléchargeur, 2 = serveur
	URL     string   `json:"url,omitempty"`  // URL de vérification à visiter
	Code    string   `json:"code,omitempty"` // code à entrer
	Raw     []string `json:"raw,omitempty"`  // lignes de log pertinentes (repli manuel)
}

// relevantAuthLines garde les dernières lignes mentionnant auth / URL / code.
func relevantAuthLines(logs string) []string {
	var out []string
	for _, ln := range strings.Split(logs, "\n") {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		if authHintRe.MatchString(t) || strings.Contains(t, "http") || codeLineRe.MatchString(t) {
			out = append(out, t)
		}
	}
	if len(out) > 12 {
		out = out[len(out)-12:]
	}
	return out
}

// parseAuthFromLogs extrait l'état d'auth d'un bloc de logs.
// done=true si le serveur est authentifié/prêt (plus rien à faire).
func parseAuthFromLogs(logs string) (st authStatus, done bool) {
	if authReadyRe.MatchString(logs) {
		return authStatus{Pending: false}, true
	}

	var url, code string
	for _, u := range anyURLRe.FindAllString(logs, -1) {
		lu := strings.ToLower(u)
		if strings.Contains(lu, "hytale") || strings.Contains(lu, "oauth") ||
			strings.Contains(lu, "device") || strings.Contains(lu, "verify") {
			url = strings.TrimRight(u, `.,)]}"'`)
			if m := userCodeParamRe.FindStringSubmatch(u); m != nil {
				code = m[1]
			}
		}
	}
	if code == "" {
		if m := codeLineRe.FindStringSubmatch(logs); m != nil {
			code = m[1]
		}
	}

	raw := relevantAuthLines(logs)
	st = authStatus{
		Pending: url != "" || code != "",
		URL:     url,
		Code:    code,
		Raw:     raw,
	}
	if strings.Contains(logs, "SERVER AUTHENTICATION REQUIRED") {
		st.Step = 2
	} else if st.Pending {
		st.Step = 1
	}
	return st, false
}

// watchAuthAndRun surveille un serveur à authentification interactive (Hytale)
// après son démarrage. Robuste : passe vite en "auth_required" (pour que le
// panel et les logs soient visibles), détecte la fin d'auth → "running", et
// l'arrêt/échec du container → "error". Ne reste jamais bloqué en "creating".
func (s *Server) watchAuthAndRun(gs *servers.GameServer, node *orchestrator.Node) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()

	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	announced := false
	for {
		select {
		case <-ctx.Done():
			slog.Warn("auth watch: timeout before completion", "id", gs.ID)
			return
		case <-ticker.C:
		}

		// Container arrêté/mort (échec du download ou crash) → erreur.
		if state, err := node.GetContainerStatus(ctx, gs.ContainerID); err == nil {
			if state == "exited" || state == "dead" {
				s.serverRepo.UpdateStatus(ctx, gs.ID, "error")
				slog.Warn("auth watch: container not running", "id", gs.ID, "state", state)
				return
			}
		}

		logs, err := node.Logs(ctx, gs.ContainerID, 1500)
		if err != nil {
			continue
		}
		_, done := parseAuthFromLogs(logs)
		if done {
			s.serverRepo.UpdateStatus(ctx, gs.ID, "running")
			s.registerRoute(ctx, gs)
			slog.Info("auth watch: server ready, running", "id", gs.ID)
			return
		}
		if authErrRe.MatchString(logs) {
			slog.Warn("auth watch: error marker in logs", "id", gs.ID)
		}
		// Jeu à auth : dès qu'il n'est pas "prêt", on expose le statut
		// auth_required (le panel affiche l'URL/code parsés OU les logs bruts).
		if !announced {
			announced = true
			s.serverRepo.UpdateStatus(ctx, gs.ID, "auth_required")
			slog.Info("auth watch: awaiting authorization", "id", gs.ID)
		}
	}
}

// handleServerAuth : GET /servers/{id}/auth → état de l'auth interactive
// (URL + code + logs bruts de secours). Sans état : relit les logs à la demande.
func (s *Server) handleServerAuth(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status == "running" {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	if def, err := servers.GetGame(gs.Game); err != nil || !def.NeedsAuth {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	logs, err := node.Logs(r.Context(), gs.ContainerID, 1500)
	if err != nil {
		respond(w, http.StatusOK, authStatus{Pending: false})
		return
	}
	st, _ := parseAuthFromLogs(logs)
	respond(w, http.StatusOK, st)
}
