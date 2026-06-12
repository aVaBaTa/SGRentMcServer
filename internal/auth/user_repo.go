package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID              string `json:"id"`
	DiscordID       string `json:"discord_id"`
	Username        string `json:"username"`
	Avatar          string `json:"avatar"`
	Email           string `json:"email"`
	Plan            string `json:"plan"`
	UnlimitedCreate bool   `json:"unlimited_create"`
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// UpsertFromDiscord crée ou met à jour un user depuis les infos Discord.
func (r *UserRepo) UpsertFromDiscord(ctx context.Context, d *DiscordUser) (*User, error) {
	const q = `
		INSERT INTO users (discord_id, username, avatar, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (discord_id) DO UPDATE
			SET username   = EXCLUDED.username,
			    avatar     = EXCLUDED.avatar,
			    email      = EXCLUDED.email,
			    updated_at = NOW()
		RETURNING id, discord_id, username, avatar, email, plan, unlimited_create
	`
	row := r.db.QueryRow(ctx, q, d.ID, d.Username, d.AvatarURL(), d.Email)

	var u User
	if err := row.Scan(&u.ID, &u.DiscordID, &u.Username, &u.Avatar, &u.Email, &u.Plan, &u.UnlimitedCreate); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT id, discord_id, username, avatar, email, plan, unlimited_create FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, q, id)

	var u User
	if err := row.Scan(&u.ID, &u.DiscordID, &u.Username, &u.Avatar, &u.Email, &u.Plan, &u.UnlimitedCreate); err != nil {
		return nil, err
	}
	return &u, nil
}

// SetUnlimitedCreate active/désactive le droit de création illimité d'un user (admin).
func (r *UserRepo) SetUnlimitedCreate(ctx context.Context, id string, enabled bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET unlimited_create = $2, updated_at = NOW() WHERE id = $1`, id, enabled)
	return err
}
