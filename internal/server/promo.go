package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Promo = rabais GLOBAL appliqué à tous les prix (affichage + facturation PayPal).
// Stocké dans Redis (clé promoKey) → pas de migration DB. EndsAt=0 → sans fin.
type Promo struct {
	Percent int   `json:"percent"`           // 0-90
	EndsAt  int64 `json:"ends_at,omitempty"` // unix secondes, 0 = permanent
}

const promoKey = "promo:global"

// Active : true si le rabais s'applique (percent>0 et pas expiré).
func (p Promo) Active() bool {
	if p.Percent <= 0 {
		return false
	}
	return p.EndsAt == 0 || time.Now().Unix() < p.EndsAt
}

// Factor : multiplicateur de prix (ex. -50% → 0.5). 1.0 si inactif.
func (p Promo) Factor() float64 {
	if !p.Active() {
		return 1.0
	}
	return float64(100-p.Percent) / 100.0
}

// getPromo lit la promo depuis Redis (zéro-valeur si absente/illisible).
func (s *Server) getPromo(ctx context.Context) Promo {
	var p Promo
	if s.rdb == nil {
		return p
	}
	val, err := s.rdb.Get(ctx, promoKey).Result()
	if err != nil || val == "" {
		return p
	}
	_ = json.Unmarshal([]byte(val), &p)
	return p
}

// setPromo écrit la promo : days>0 → fin dans N jours ; days<=0 → permanent.
func (s *Server) setPromo(ctx context.Context, percent, days int) (Promo, error) {
	p := Promo{Percent: percent}
	if days > 0 {
		p.EndsAt = time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()
	}
	b, _ := json.Marshal(p)
	if err := s.rdb.Set(ctx, promoKey, b, 0).Err(); err != nil {
		return Promo{}, err
	}
	return p, nil
}

// applyPromoCents applique le rabais actif à un montant en cents (arrondi).
func (s *Server) applyPromoCents(ctx context.Context, cents int) int {
	f := s.getPromo(ctx).Factor()
	if f >= 1.0 {
		return cents
	}
	return int(float64(cents)*f + 0.5)
}

// handlePromo : GET /api/v1/promo (public) → rabais actif pour l'affichage des prix.
func (s *Server) handlePromo(w http.ResponseWriter, r *http.Request) {
	p := s.getPromo(r.Context())
	if !p.Active() {
		respond(w, http.StatusOK, Promo{Percent: 0})
		return
	}
	respond(w, http.StatusOK, p)
}

// handleAdminGetPromo : GET /api/v1/admin/promo → promo brute (même expirée).
func (s *Server) handleAdminGetPromo(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	respond(w, http.StatusOK, s.getPromo(r.Context()))
}

// handleAdminSetPromo : POST /api/v1/admin/promo {percent, days} → définit le rabais global.
func (s *Server) handleAdminSetPromo(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Percent int `json:"percent"`
		Days    int `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Percent < 0 || body.Percent > 90 {
		http.Error(w, "percent doit être entre 0 et 90", http.StatusBadRequest)
		return
	}
	p, err := s.setPromo(r.Context(), body.Percent, body.Days)
	if err != nil {
		http.Error(w, "redis error", http.StatusServiceUnavailable)
		return
	}
	respond(w, http.StatusOK, p)
}
