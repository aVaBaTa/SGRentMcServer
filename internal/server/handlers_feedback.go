package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

// handleFeedback : POST /api/v1/feedback (public) → enregistre une réponse au sondage.
func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Rating  int    `json:"rating"`
		Source  string `json:"source"`
		Usecase string `json:"usecase"`
		Game    string `json:"game"`
		Message string `json:"message"`
		Email   string `json:"email"`
		Page    string `json:"page"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	// Au moins une info utile (note ou commentaire ou un choix).
	if b.Rating == 0 && b.Message == "" && b.Source == "" && b.Usecase == "" && b.Game == "" {
		http.Error(w, "réponse vide", http.StatusBadRequest)
		return
	}
	var ratingPtr any
	if b.Rating >= 1 && b.Rating <= 5 {
		ratingPtr = b.Rating
	}
	_, err := s.db.Exec(r.Context(),
		`INSERT INTO feedback (rating, source, usecase, game, message, email, page)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		ratingPtr, clip(b.Source, 40), clip(b.Usecase, 40), clip(b.Game, 40),
		clip(b.Message, 2000), clip(b.Email, 200), clip(b.Page, 200),
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleAdminFeedback : GET /api/v1/admin/feedback → agrégats + dernières réponses.
func (s *Server) handleAdminFeedback(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := r.Context()
	type kv struct {
		Key   string `json:"key"`
		Count int    `json:"count"`
	}
	type respRow struct {
		Rating    int    `json:"rating"`
		Source    string `json:"source"`
		Usecase   string `json:"usecase"`
		Game      string `json:"game"`
		Message   string `json:"message"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}
	out := struct {
		Total     int       `json:"total"`
		AvgRating float64   `json:"avg_rating"`
		BySource  []kv      `json:"by_source"`
		ByUsecase []kv      `json:"by_usecase"`
		ByGame    []kv      `json:"by_game"`
		Recent    []respRow `json:"recent"`
	}{BySource: []kv{}, ByUsecase: []kv{}, ByGame: []kv{}, Recent: []respRow{}}

	s.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(ROUND(AVG(rating)::numeric,2),0) FROM feedback`).Scan(&out.Total, &out.AvgRating)

	groupBy := func(col string) []kv {
		res := []kv{}
		rows, err := s.db.Query(ctx, `SELECT COALESCE(NULLIF(`+col+`,''),'(non précisé)'), COUNT(*) FROM feedback GROUP BY 1 ORDER BY 2 DESC`)
		if err != nil {
			return res
		}
		defer rows.Close()
		for rows.Next() {
			var k kv
			if rows.Scan(&k.Key, &k.Count) == nil {
				res = append(res, k)
			}
		}
		return res
	}
	out.BySource = groupBy("source")
	out.ByUsecase = groupBy("usecase")
	out.ByGame = groupBy("game")

	if rows, err := s.db.Query(ctx, `SELECT COALESCE(rating,0), COALESCE(source,''), COALESCE(usecase,''), COALESCE(game,''), COALESCE(message,''), COALESCE(email,''), to_char(created_at,'YYYY-MM-DD HH24:MI') FROM feedback ORDER BY created_at DESC LIMIT 30`); err == nil {
		for rows.Next() {
			var p respRow
			if rows.Scan(&p.Rating, &p.Source, &p.Usecase, &p.Game, &p.Message, &p.Email, &p.CreatedAt) == nil {
				out.Recent = append(out.Recent, p)
			}
		}
		rows.Close()
	}
	respond(w, http.StatusOK, out)
}
