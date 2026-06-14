package server

import (
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
)

// handleAdminMetrics : GET /api/v1/admin/metrics → métriques business pour /admin.
// Agrège paiements (revenu réel), serveurs, utilisateurs, conversion. Lecture seule.
func (s *Server) handleAdminMetrics(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	type gameRow struct {
		Game  string `json:"game"`
		Count int    `json:"count"`
		Paid  int    `json:"paid"`
	}
	type planRow struct {
		Plan  string `json:"plan"`
		Count int    `json:"count"`
	}
	type payRow struct {
		Plan        string `json:"plan"`
		AmountCents int    `json:"amount_cents"`
		Currency    string `json:"currency"`
		Status      string `json:"status"`
		CreatedAt   string `json:"created_at"`
	}

	out := struct {
		RevenueTotalCents int64     `json:"revenue_total_cents"`
		RevenueMonthCents int64     `json:"revenue_month_cents"`
		CompletedCount    int       `json:"completed_count"`
		PendingCount      int       `json:"pending_count"`
		PendingCents      int64     `json:"pending_cents"`
		MRRCents          int64     `json:"mrr_cents"` // estimation : serveurs payants actifs × prix plan
		TotalUsers        int       `json:"total_users"`
		PayingUsers       int       `json:"paying_users"`
		TotalServers      int       `json:"total_servers"`
		PaidServers       int       `json:"paid_servers"`
		FreeServers       int       `json:"free_servers"`
		ByGame            []gameRow `json:"by_game"`
		ByPlan            []planRow `json:"by_plan"`
		Recent            []payRow  `json:"recent"`
	}{ByGame: []gameRow{}, ByPlan: []planRow{}, Recent: []payRow{}}

	// Comptes exclus (proprio/tests, ex. aVaBaTa) → leurs serveurs/paiements/users ne
	// comptent pas dans les métriques. excl = leurs user_id (texte). Liste vide → rien d'exclu
	// (`<> ALL('{}')` est vrai pour tous).
	unames := s.cfg.MetricsExcludeUsers
	excl := []string{}
	if len(unames) > 0 {
		if rows, err := s.db.Query(ctx, `SELECT id::text FROM users WHERE lower(username) = ANY($1)`, unames); err == nil {
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					excl = append(excl, id)
				}
			}
			rows.Close()
		}
	}

	// Revenu encaissé (status completed) total + ce mois.
	s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents),0), COUNT(*) FROM payments WHERE status='completed' AND user_id::text <> ALL($1)`, excl).
		Scan(&out.RevenueTotalCents, &out.CompletedCount)
	s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents),0) FROM payments WHERE status='completed' AND created_at >= date_trunc('month', now()) AND user_id::text <> ALL($1)`, excl).
		Scan(&out.RevenueMonthCents)
	// Paiements en attente (créés mais non capturés).
	s.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(amount_cents),0) FROM payments WHERE status='created' AND user_id::text <> ALL($1)`, excl).
		Scan(&out.PendingCount, &out.PendingCents)
	// Utilisateurs (proprio exclu par username).
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE lower(username) <> ALL($1)`, unames).Scan(&out.TotalUsers)
	s.db.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM payments WHERE status='completed' AND user_id::text <> ALL($1)`, excl).Scan(&out.PayingUsers)
	// Serveurs.
	s.db.QueryRow(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE plan <> 'free'), COUNT(*) FILTER (WHERE plan = 'free') FROM game_servers WHERE user_id::text <> ALL($1)`, excl).
		Scan(&out.TotalServers, &out.PaidServers, &out.FreeServers)

	// Répartition par jeu (+ payants).
	if rows, err := s.db.Query(ctx, `SELECT game, COUNT(*), COUNT(*) FILTER (WHERE plan <> 'free') FROM game_servers WHERE user_id::text <> ALL($1) GROUP BY game ORDER BY COUNT(*) DESC`, excl); err == nil {
		for rows.Next() {
			var g gameRow
			if rows.Scan(&g.Game, &g.Count, &g.Paid) == nil {
				out.ByGame = append(out.ByGame, g)
			}
		}
		rows.Close()
	}

	// Répartition par plan (sert aussi à l'estimation MRR).
	if rows, err := s.db.Query(ctx, `SELECT plan, COUNT(*) FROM game_servers WHERE user_id::text <> ALL($1) GROUP BY plan`, excl); err == nil {
		for rows.Next() {
			var p planRow
			if rows.Scan(&p.Plan, &p.Count) == nil {
				out.ByPlan = append(out.ByPlan, p)
				if pl, err := servers.GetPlan(p.Plan); err == nil {
					out.MRRCents += int64(pl.PriceCents) * int64(p.Count)
				}
			}
		}
		rows.Close()
	}

	// Derniers paiements (tous statuts).
	if rows, err := s.db.Query(ctx, `SELECT plan, amount_cents, currency, status, to_char(created_at,'YYYY-MM-DD HH24:MI') FROM payments WHERE user_id::text <> ALL($1) ORDER BY created_at DESC LIMIT 10`, excl); err == nil {
		for rows.Next() {
			var p payRow
			if rows.Scan(&p.Plan, &p.AmountCents, &p.Currency, &p.Status, &p.CreatedAt) == nil {
				out.Recent = append(out.Recent, p)
			}
		}
		rows.Close()
	}

	respond(w, http.StatusOK, out)
}
