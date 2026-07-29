package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
)

// Accès anticipé Calradia-Coop. Le mod (build client-only) est distribué sur
// calradiacoop.vbt-prog.com/download ; le fichier est réservé aux AdminUsers
// (aVaBaTa) et aux pseudos autorisés dans /admin, tant que l'ouverture publique
// n'est pas activée explicitement (cf. calradia_release.go). nginx (bloc
// calradiacoop) protège /files/ via auth_request → handleCalradiaDownloadAuth.

// calradiaAccessKey : clé d'accès d'une instance louée, dérivée de l'ID du
// serveur + secret interne (déterministe → jamais stockée : la même valeur est
// injectée dans CALRADIA_ACCESS_KEY à la création du container et affichée au
// propriétaire dans le panel). Format lisible/partageable : XXXX-XXXX (hex).
func (s *Server) calradiaAccessKey(serverID string) string {
	sum := sha256.Sum256([]byte("calradia-access|" + serverID + "|" + s.cfg.JWTSecret))
	hexs := strings.ToUpper(fmt.Sprintf("%x", sum[:4]))
	return hexs[:4] + "-" + hexs[4:8]
}

// fillCalradiaAccessKey : renseigne AccessKey sur les serveurs Calradia-Coop
// d'une réponse au propriétaire (champ calculé, absent de la base).
func (s *Server) fillCalradiaAccessKey(list ...*servers.GameServer) {
	for _, gs := range list {
		if gs != nil && gs.Game == "calradia-coop" {
			gs.AccessKey = s.calradiaAccessKey(gs.ID)
		}
	}
}

// isAdminUser : le pseudo appartient-il à AdminUsers (insensible à la casse) ?
func (s *Server) isAdminUser(username string) bool {
	for _, a := range s.cfg.AdminUsers {
		if strings.EqualFold(username, a) {
			return true
		}
	}
	return false
}

// calradiaReleased : le téléchargement est-il ouvert au grand public ?
// ⚠️ Ce n'est PAS « la date est passée » : l'ouverture demande l'interrupteur
// explicite `public` réglé dans /admin (cf. calradia_release.go). Une date
// dépassée sans cet interrupteur laisse le mod verrouillé.
func (s *Server) calradiaReleased(ctx context.Context) bool {
	return s.getCalradiaRelease(ctx).Released()
}

// calradiaApproved : le pseudo a-t-il une candidature early-access approuvée ?
func (s *Server) calradiaApproved(r *http.Request, username string) bool {
	var ok bool
	err := s.db.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM calradia_early_access
		  WHERE lower(discord_username) = lower($1) AND status = 'approved')`,
		username).Scan(&ok)
	return err == nil && ok
}

// handleCalradiaDownloadAuth : GET /api/v1/calradia/download-auth — cible de
// l'auth_request nginx qui garde /files/ sur calradiacoop.vbt-prog.com.
// 200 si : release publique passée, OU session admin, OU candidature approuvée.
func (s *Server) handleCalradiaDownloadAuth(w http.ResponseWriter, r *http.Request) {
	if s.calradiaReleased(r.Context()) {
		w.WriteHeader(http.StatusOK)
		return
	}
	if u := s.extractUsername(r); u != "" && (s.isAdminUser(u) || s.calradiaApproved(r, u)) {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "early access only", http.StatusUnauthorized)
}

// handleCalradiaApply : POST /api/v1/calradia/early-access (public) — dépose une
// candidature d'accès anticipé. Session Discord requise (pseudo fiable, anti-spam) ;
// re-candidater après un refus repasse la demande en « pending ».
func (s *Server) handleCalradiaApply(w http.ResponseWriter, r *http.Request) {
	username := s.extractUsername(r)
	if username == "" {
		http.Error(w, "login required", http.StatusUnauthorized)
		return
	}
	var b struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<15)).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	_, err := s.db.Exec(r.Context(),
		`INSERT INTO calradia_early_access (discord_username, message)
		 VALUES ($1, $2)
		 ON CONFLICT (lower(discord_username)) DO UPDATE
		   SET message = EXCLUDED.message,
		       status = CASE WHEN calradia_early_access.status = 'approved'
		                     THEN 'approved' ELSE 'pending' END,
		       decided_at = CASE WHEN calradia_early_access.status = 'approved'
		                     THEN calradia_early_access.decided_at ELSE NULL END`,
		username, clip(b.Message, 1000))
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "submitted"})
}

// handleCalradiaStatus : GET /api/v1/calradia/early-access/status — état vu par le
// visiteur : release publique ? connecté ? candidature (pending/approved/rejected) ?
// téléchargement autorisé ? Pilote l'UI de la page /download.
func (s *Server) handleCalradiaStatus(w http.ResponseWriter, r *http.Request) {
	username := s.extractUsername(r)
	rel := s.getCalradiaRelease(r.Context())
	out := struct {
		Released    bool   `json:"released"`
		ReleaseAt   string `json:"release_at"`
		Username    string `json:"username,omitempty"`
		Application string `json:"application,omitempty"` // pending | approved | rejected
		CanDownload bool   `json:"can_download"`
	}{
		Released:  rel.Released(),
		ReleaseAt: rel.ReleaseAt,
		Username:  username,
	}
	if username != "" {
		var status string
		if err := s.db.QueryRow(r.Context(),
			`SELECT status FROM calradia_early_access WHERE lower(discord_username) = lower($1)`,
			username).Scan(&status); err == nil {
			out.Application = status
		}
	}
	out.CanDownload = out.Released ||
		(username != "" && (s.isAdminUser(username) || out.Application == "approved"))
	respond(w, http.StatusOK, out)
}

// --- Admin (X-Admin-Token, via le monitor /admin) ---

// handleAdminCalradiaList : GET /api/v1/admin/calradia/early-access — toutes les
// candidatures, plus récentes d'abord.
func (s *Server) handleAdminCalradiaList(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := s.db.Query(r.Context(),
		`SELECT id, discord_username, message, status, created_at, decided_at
		   FROM calradia_early_access ORDER BY created_at DESC LIMIT 500`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type app struct {
		ID        int64      `json:"id"`
		Username  string     `json:"username"`
		Message   string     `json:"message"`
		Status    string     `json:"status"`
		CreatedAt time.Time  `json:"created_at"`
		DecidedAt *time.Time `json:"decided_at,omitempty"`
	}
	apps := []app{}
	for rows.Next() {
		var a app
		if err := rows.Scan(&a.ID, &a.Username, &a.Message, &a.Status, &a.CreatedAt, &a.DecidedAt); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		apps = append(apps, a)
	}
	rel := s.getCalradiaRelease(r.Context())
	respond(w, http.StatusOK, map[string]any{
		"release_at":   rel.ReleaseAt,
		"public":       rel.Public,
		"released":     rel.Released(),
		"applications": apps,
	})
}

// handleAdminCalradiaDecide : POST /api/v1/admin/calradia/early-access/{id}/status
// — approuve/rejette une candidature ({"status":"approved"|"rejected"|"pending"}).
func (s *Server) handleAdminCalradiaDecide(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var b struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&b); err != nil ||
		(b.Status != "approved" && b.Status != "rejected" && b.Status != "pending") {
		http.Error(w, "status must be approved, rejected or pending", http.StatusBadRequest)
		return
	}
	tag, err := s.db.Exec(r.Context(),
		`UPDATE calradia_early_access
		    SET status = $1,
		        decided_at = CASE WHEN $1 = 'pending' THEN NULL ELSE now() END
		  WHERE id = $2`,
		b.Status, chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": b.Status})
}
