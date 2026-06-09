package billing

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	ServerID        string `json:"server_id"`
	Plan            string `json:"plan"`
	AmountCents     int    `json:"amount_cents"`
	Currency        string `json:"currency"`
	Provider        string `json:"provider"`
	ProviderOrderID string `json:"provider_order_id"`
	Status          string `json:"status"`
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p *Payment) error {
	const q = `
		INSERT INTO payments (user_id, server_id, plan, amount_cents, currency, provider, provider_order_id, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id
	`
	return r.db.QueryRow(ctx, q,
		p.UserID, nullable(p.ServerID), p.Plan, p.AmountCents, p.Currency,
		p.Provider, p.ProviderOrderID, p.Status,
	).Scan(&p.ID)
}

// GetByOrderID retourne le paiement lié à une commande PayPal (et vérifie le propriétaire).
func (r *Repo) GetByOrderID(ctx context.Context, orderID, userID string) (*Payment, error) {
	const q = `
		SELECT id, user_id, COALESCE(server_id::text,''), plan, amount_cents, currency, provider, provider_order_id, status
		FROM payments WHERE provider_order_id=$1 AND user_id=$2
	`
	p := &Payment{}
	err := r.db.QueryRow(ctx, q, orderID, userID).Scan(
		&p.ID, &p.UserID, &p.ServerID, &p.Plan, &p.AmountCents, &p.Currency,
		&p.Provider, &p.ProviderOrderID, &p.Status,
	)
	return p, err
}

func (r *Repo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE payments SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
