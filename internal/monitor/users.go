package monitor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserInfo struct {
	ID              string    `json:"id"`
	DiscordID       string    `json:"discord_id"`
	Username        string    `json:"username"`
	Avatar          string    `json:"avatar"`
	Email           string    `json:"email"`
	Plan            string    `json:"plan"`
	UnlimitedCreate bool      `json:"unlimited_create"`
	ServerCount     int       `json:"server_count"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListUsers retourne tous les utilisateurs avec leur nombre de serveurs.
func ListUsers(ctx context.Context, db *pgxpool.Pool) ([]UserInfo, error) {
	const q = `
		SELECT u.id, u.discord_id, u.username, COALESCE(u.avatar,''), COALESCE(u.email,''),
		       u.plan, COALESCE(u.unlimited_create, FALSE), COUNT(g.id), u.created_at
		FROM users u
		LEFT JOIN game_servers g ON g.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC
	`
	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UserInfo
	for rows.Next() {
		var u UserInfo
		if err := rows.Scan(&u.ID, &u.DiscordID, &u.Username, &u.Avatar, &u.Email,
			&u.Plan, &u.UnlimitedCreate, &u.ServerCount, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}
