package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Verrou de téléchargement Calradia-Coop.
//
// ⚠️ Leçon du 2026-07 : la version précédente ouvrait le téléchargement À TOUS
// dès que la date de sortie était passée. La date (17 juillet) est passée sans
// que le mod soit prêt → le build s'est retrouvé public sans action de personne.
// Désormais la date est purement INFORMATIVE (compte à rebours affiché sur le
// site) et l'ouverture publique est un interrupteur explicite (`Public`).
// Tant que `Public` est faux, seuls les admins et la liste d'accès téléchargent.
//
// Stocké dans Redis (comme la promo) → date modifiable depuis /admin sans
// redéploiement, ce qui évite de rebuild l'API à chaque report.

const calradiaReleaseKey = "calradia:release"

// CalradiaRelease : état du verrou de téléchargement.
type CalradiaRelease struct {
	ReleaseAt string `json:"release_at"` // RFC3339 ; "" = aucune date annoncée
	Public    bool   `json:"public"`     // true = ouvert à tous (une fois la date passée)
}

// Released : le téléchargement est-il ouvert à tout le monde ?
// Faux si l'interrupteur public est coupé, si aucune date n'est fixée, ou si la
// date n'est pas encore atteinte.
func (c CalradiaRelease) Released() bool {
	if !c.Public || c.ReleaseAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, c.ReleaseAt)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}

// getCalradiaRelease lit l'état depuis Redis. En l'absence de valeur (ou si
// Redis est indisponible), on retombe sur la date de config avec Public=false :
// le défaut est donc VERROUILLÉ — jamais d'ouverture accidentelle.
func (s *Server) getCalradiaRelease(ctx context.Context) CalradiaRelease {
	fallback := CalradiaRelease{Public: false}
	if !s.cfg.CalradiaReleaseAt.IsZero() {
		fallback.ReleaseAt = s.cfg.CalradiaReleaseAt.Format(time.RFC3339)
	}
	if s.rdb == nil {
		return fallback
	}
	val, err := s.rdb.Get(ctx, calradiaReleaseKey).Result()
	if err != nil || val == "" {
		return fallback
	}
	var c CalradiaRelease
	if err := json.Unmarshal([]byte(val), &c); err != nil {
		return fallback
	}
	return c
}

// setCalradiaRelease écrit l'état (date + interrupteur public) dans Redis.
func (s *Server) setCalradiaRelease(ctx context.Context, c CalradiaRelease) error {
	b, _ := json.Marshal(c)
	return s.rdb.Set(ctx, calradiaReleaseKey, b, 0).Err()
}

// --- Admin : verrou de sortie ---

// handleAdminCalradiaGetRelease : GET /api/v1/admin/calradia/release
func (s *Server) handleAdminCalradiaGetRelease(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	c := s.getCalradiaRelease(r.Context())
	respond(w, http.StatusOK, map[string]any{
		"release_at": c.ReleaseAt,
		"public":     c.Public,
		"released":   c.Released(),
	})
}

// handleAdminCalradiaSetRelease : POST /api/v1/admin/calradia/release
// Corps : {"release_at":"2026-09-15T00:00:00-04:00","public":false}
// release_at vide = aucune date annoncée (le site n'affiche pas de compte à rebours).
func (s *Server) handleAdminCalradiaSetRelease(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if s.rdb == nil {
		http.Error(w, "redis indisponible", http.StatusServiceUnavailable)
		return
	}
	var b struct {
		ReleaseAt string `json:"release_at"`
		Public    bool   `json:"public"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	b.ReleaseAt = strings.TrimSpace(b.ReleaseAt)
	if b.ReleaseAt != "" {
		if _, err := time.Parse(time.RFC3339, b.ReleaseAt); err != nil {
			http.Error(w, "release_at doit être au format RFC3339 (ex. 2026-09-15T00:00:00-04:00)", http.StatusBadRequest)
			return
		}
	}
	c := CalradiaRelease{ReleaseAt: b.ReleaseAt, Public: b.Public}
	if err := s.setCalradiaRelease(r.Context(), c); err != nil {
		http.Error(w, "redis error", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"release_at": c.ReleaseAt,
		"public":     c.Public,
		"released":   c.Released(),
	})
}

// --- Admin : liste d'accès au téléchargement ---

// handleAdminCalradiaUsers : GET /api/v1/admin/calradia/users — utilisateurs
// Discord connus (déjà connectés à Playrena) avec leur état d'autorisation, plus
// les entrées ajoutées à la main qui ne correspondent à aucun compte connu.
func (s *Server) handleAdminCalradiaUsers(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	type entry struct {
		Username  string `json:"username"`
		DiscordID string `json:"discord_id,omitempty"`
		Status    string `json:"status,omitempty"` // approved | rejected | pending | ""
		Allowed   bool   `json:"allowed"`
		Known     bool   `json:"known"` // a déjà ouvert une session Playrena
		Admin     bool   `json:"admin"` // admin → toujours autorisé
	}

	rows, err := s.db.Query(r.Context(),
		`SELECT u.username, u.discord_id, COALESCE(e.status, '')
		   FROM users u
		   LEFT JOIN calradia_early_access e
		          ON lower(e.discord_username) = lower(u.username)
		  ORDER BY lower(u.username) LIMIT 1000`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []entry{}
	seen := map[string]bool{}
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.Username, &e.DiscordID, &e.Status); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		e.Known = true
		e.Admin = s.isAdminUser(e.Username)
		e.Allowed = e.Status == "approved" || e.Admin
		seen[strings.ToLower(e.Username)] = true
		list = append(list, e)
	}

	// Entrées de la liste d'accès sans compte Playrena connu (ajouts manuels).
	extra, err := s.db.Query(r.Context(),
		`SELECT discord_username, status FROM calradia_early_access
		  ORDER BY lower(discord_username) LIMIT 1000`)
	if err == nil {
		defer extra.Close()
		for extra.Next() {
			var e entry
			if err := extra.Scan(&e.Username, &e.Status); err != nil {
				continue
			}
			if seen[strings.ToLower(e.Username)] {
				continue
			}
			e.Admin = s.isAdminUser(e.Username)
			e.Allowed = e.Status == "approved" || e.Admin
			list = append(list, e)
		}
	}

	respond(w, http.StatusOK, map[string]any{"users": list})
}

// handleAdminCalradiaAllow : POST /api/v1/admin/calradia/allow
// Corps : {"username":"Zeddwolff","allowed":true}
// Autorise (ou retire) un pseudo Discord. Fonctionne aussi pour un pseudo qui ne
// s'est jamais connecté à Playrena (ajout manuel) : la ligne est créée au besoin.
func (s *Server) handleAdminCalradiaAllow(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var b struct {
		Username string `json:"username"`
		Allowed  bool   `json:"allowed"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	b.Username = strings.TrimSpace(b.Username)
	if b.Username == "" {
		http.Error(w, "username requis", http.StatusBadRequest)
		return
	}
	status := "rejected"
	if b.Allowed {
		status = "approved"
	}
	_, err := s.db.Exec(r.Context(),
		`INSERT INTO calradia_early_access (discord_username, message, status, decided_at)
		 VALUES ($1, '', $2, now())
		 ON CONFLICT (lower(discord_username)) DO UPDATE
		   SET status = EXCLUDED.status, decided_at = now()`,
		b.Username, status)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"username": b.Username, "allowed": b.Allowed})
}
