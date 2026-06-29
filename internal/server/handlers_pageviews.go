package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/redis/go-redis/v9"
)

// Présence « en ligne maintenant » : chaque onglet ouvert envoie un battement (heartbeat)
// toutes les ~20 s. On garde les battements des presenceTTLSec dernières secondes dans un
// sorted set Redis (score = horodatage du dernier battement) → le nombre de membres récents
// = le nombre d'onglets/visiteurs ayant la page ouverte. Le hash meta retient leur page + pseudo.
const (
	presenceZKey    = "presence:online" // sorted set : member=vid, score=last-heartbeat (unix)
	presenceHKey    = "presence:meta"   // hash : field=vid, value="username\x1fpath"
	presenceTTLSec  = 60                 // un vid sans battement depuis 60 s n'est plus « en ligne »
	presenceMetaSep = "\x1f"
)

// extractUsername renvoie le pseudo du visiteur si son cookie JWT est valide, sinon "".
func (s *Server) extractUsername(r *http.Request) string {
	if c, err := r.Cookie("token"); err == nil {
		if claims, err := auth.ValidateToken(c.Value, s.cfg.JWTSecret); err == nil {
			return clip(claims.Username, 100)
		}
	}
	return ""
}

// prunePresence retire les vid dont le dernier battement est trop ancien (> presenceTTLSec).
func (s *Server) prunePresence(r *http.Request, now int64) {
	ctx := r.Context()
	cutoff := fmt.Sprintf("%d", now-presenceTTLSec)
	stale, err := s.rdb.ZRangeByScore(ctx, presenceZKey, &redis.ZRangeBy{Min: "0", Max: cutoff}).Result()
	if err != nil || len(stale) == 0 {
		return
	}
	members := make([]any, len(stale))
	for i, v := range stale {
		members[i] = v
	}
	s.rdb.ZRem(ctx, presenceZKey, members...)
	s.rdb.HDel(ctx, presenceHKey, stale...)
}

// handlePresence : POST /api/v1/presence (public) → battement « je suis encore là ».
// Best-effort : jamais d'erreur visible côté visiteur.
func (s *Server) handlePresence(w http.ResponseWriter, r *http.Request) {
	if s.rdb == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var b struct {
		Vid  string `json:"vid"`
		Path string `json:"path"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<13)).Decode(&b); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	vid := clip(b.Vid, 64)
	if vid == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	now := time.Now().Unix()
	ctx := r.Context()
	s.rdb.ZAdd(ctx, presenceZKey, redis.Z{Score: float64(now), Member: vid})
	s.rdb.HSet(ctx, presenceHKey, vid, s.extractUsername(r)+presenceMetaSep+normalizePath(b.Path))
	s.prunePresence(r, now)
	w.WriteHeader(http.StatusNoContent)
}

// handleTrack : POST /api/v1/track (public) → enregistre une vue de page.
// Le beacon front l'appelle à chaque changement de route. Si le visiteur est connecté
// (cookie JWT « token »), on attache son user_id + pseudo ; sinon visiteur anonyme.
// Volontairement tolérant : jamais d'erreur visible côté visiteur (best-effort analytics).
func (s *Server) handleTrack(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Path    string `json:"path"`
		Referer string `json:"referer"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&b); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	path := normalizePath(b.Path)
	if path == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Identité optionnelle via le cookie JWT (les visiteurs anonymes naviguent aussi).
	var userID, username any
	if c, err := r.Cookie("token"); err == nil {
		if claims, err := auth.ValidateToken(c.Value, s.cfg.JWTSecret); err == nil {
			if claims.UserID != "" {
				userID = claims.UserID
			}
			if claims.Username != "" {
				username = clip(claims.Username, 100)
			}
		}
	}

	var ref any
	if rf := normalizePath(b.Referer); rf != "" {
		ref = rf
	}
	_, _ = s.db.Exec(r.Context(),
		`INSERT INTO page_views (user_id, username, path, referer) VALUES ($1,$2,$3,$4)`,
		userID, username, path, ref,
	)
	w.WriteHeader(http.StatusNoContent)
}

// contains : true si v (déjà en minuscule) est dans la liste.
func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// normalizePath nettoie un chemin (retire l'origine et la query string, garde « / »).
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	// Retire un éventuel schéma/hôte (le front envoie un chemin relatif, mais on durcit).
	if i := strings.Index(p, "://"); i >= 0 {
		if j := strings.IndexByte(p[i+3:], '/'); j >= 0 {
			p = p[i+3+j:]
		} else {
			p = "/"
		}
	}
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	if !strings.HasPrefix(p, "/") {
		return ""
	}
	return clip(p, 200)
}

// handleAdminPageViews : GET /api/v1/admin/pageviews → analytics de navigation pour /admin.
// Exclut les comptes proprio/tests (METRICS_EXCLUDE_USERS, ex. aVaBaTa). Lecture seule.
func (s *Server) handleAdminPageViews(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	// Filtre d'exclusion : on garde les anonymes (username NULL) et tout user dont le
	// pseudo n'est pas dans la liste exclue (insensible à la casse).
	excl := s.cfg.MetricsExcludeUsers // ex. ["avabata"] ; liste vide → rien d'exclu
	const notExcluded = `(username IS NULL OR lower(username) <> ALL($1))`

	type pageRow struct {
		Path     string `json:"path"`
		Views    int    `json:"views"`
		Visitors int    `json:"visitors"` // utilisateurs connectés distincts
	}
	type activeRow struct {
		Username string `json:"username"`
		Path     string `json:"path"`
		LastSeen string `json:"last_seen"`
	}
	type viewRow struct {
		Username  string `json:"username"`
		Path      string `json:"path"`
		CreatedAt string `json:"created_at"`
	}
	type onlineRow struct {
		Username string `json:"username"`
		Path     string `json:"path"`
	}
	out := struct {
		OnlineCount int         `json:"online_count"` // onglets/visiteurs ayant la page ouverte (temps réel)
		Online      []onlineRow `json:"online"`       // détail : page actuelle de chacun
		Views24h    int         `json:"views_24h"`
		Views7d     int         `json:"views_7d"`
		LoggedUsers int         `json:"logged_users_7d"` // utilisateurs connectés distincts (7 j)
		ActiveNow   []activeRow `json:"active_now"`      // connectés vus dans les 30 dernières min
		ByPage      []pageRow   `json:"by_page"`         // pages les + vues (7 j)
		Recent      []viewRow   `json:"recent"`          // flux des dernières vues
	}{Online: []onlineRow{}, ActiveNow: []activeRow{}, ByPage: []pageRow{}, Recent: []viewRow{}}

	// --- En ligne maintenant (présence temps réel, Redis) ---
	if s.rdb != nil {
		now := time.Now().Unix()
		s.prunePresence(r, now)
		if vids, err := s.rdb.ZRange(ctx, presenceZKey, 0, -1).Result(); err == nil && len(vids) > 0 {
			metas, _ := s.rdb.HMGet(ctx, presenceHKey, vids...).Result()
			for _, m := range metas {
				str, _ := m.(string)
				uname, path, _ := strings.Cut(str, presenceMetaSep)
				if uname != "" && contains(s.cfg.MetricsExcludeUsers, strings.ToLower(uname)) {
					continue // proprio/tests exclus du compteur
				}
				name := uname
				if name == "" {
					name = "(anonyme)"
				}
				out.Online = append(out.Online, onlineRow{Username: name, Path: path})
			}
			out.OnlineCount = len(out.Online)
		}
	}

	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM page_views WHERE created_at >= now()-interval '24 hours' AND `+notExcluded, excl).Scan(&out.Views24h)
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM page_views WHERE created_at >= now()-interval '7 days' AND `+notExcluded, excl).Scan(&out.Views7d)
	s.db.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND created_at >= now()-interval '7 days' AND `+notExcluded, excl).Scan(&out.LoggedUsers)

	// Utilisateurs connectés actifs : leur dernière page vue dans les 30 dernières minutes.
	// Horodatage en ISO 8601 UTC (« …Z ») → le front /admin l'affiche dans le fuseau du
	// navigateur (= l'heure locale de l'admin, sans fuseau codé en dur côté serveur).
	if rows, err := s.db.Query(ctx,
		`SELECT DISTINCT ON (user_id) username, path, to_char(created_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM page_views
		 WHERE user_id IS NOT NULL AND created_at >= now()-interval '30 minutes' AND `+notExcluded+`
		 ORDER BY user_id, created_at DESC`, excl); err == nil {
		for rows.Next() {
			var a activeRow
			if rows.Scan(&a.Username, &a.Path, &a.LastSeen) == nil {
				out.ActiveNow = append(out.ActiveNow, a)
			}
		}
		rows.Close()
	}

	// Pages les plus vues sur 7 jours (+ visiteurs connectés distincts).
	if rows, err := s.db.Query(ctx,
		`SELECT path, COUNT(*), COUNT(DISTINCT user_id)
		 FROM page_views
		 WHERE created_at >= now()-interval '7 days' AND `+notExcluded+`
		 GROUP BY path ORDER BY COUNT(*) DESC LIMIT 25`, excl); err == nil {
		for rows.Next() {
			var p pageRow
			if rows.Scan(&p.Path, &p.Views, &p.Visitors) == nil {
				out.ByPage = append(out.ByPage, p)
			}
		}
		rows.Close()
	}

	// Flux des dernières vues (connectées ou anonymes).
	if rows, err := s.db.Query(ctx,
		`SELECT COALESCE(username,'(anonyme)'), path, to_char(created_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM page_views WHERE `+notExcluded+`
		 ORDER BY created_at DESC LIMIT 50`, excl); err == nil {
		for rows.Next() {
			var v viewRow
			if rows.Scan(&v.Username, &v.Path, &v.CreatedAt) == nil {
				out.Recent = append(out.Recent, v)
			}
		}
		rows.Close()
	}

	respond(w, http.StatusOK, out)
}
