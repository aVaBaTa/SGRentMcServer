package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string
	DiscordID string
	Username  string
	Avatar    string
	Email     string
	Plan      string
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
		RETURNING id, discord_id, username, avatar, email, plan
	`
	row := r.db.QueryRow(ctx, q, d.ID, d.Username, d.AvatarURL(), d.Email)

	var u User
	if err := row.Scan(&u.ID, &u.DiscordID, &u.Username, &u.Avatar, &u.Email, &u.Plan); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT id, discord_id, username, avatar, email, plan FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, q, id)

	var u User
	if err := row.Scan(&u.ID, &u.DiscordID, &u.Username, &u.Avatar, &u.Email, &u.Plan); err != nil {
		return nil, err
	}
	return &u, nil
}
