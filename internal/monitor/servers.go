package monitor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ServerInfo = un serveur de jeu (table game_servers) enrichi du propriétaire.
type ServerInfo struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	Game        string    `json:"game"`
	Plan        string    `json:"plan"`
	Status      string    `json:"status"`
	Subdomain   string    `json:"subdomain"`
	Node        string    `json:"node"`
	ContainerID string    `json:"container_id"`
	RAMMb       int64     `json:"ram_mb"`
	CPUCores    float64   `json:"cpu_cores"`
	Port        int       `json:"port"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListServers retourne tous les serveurs de jeu avec le pseudo du propriétaire.
func ListServers(ctx context.Context, db *pgxpool.Pool) ([]ServerInfo, error) {
	const q = `
		SELECT g.id, COALESCE(u.username,''), g.game, g.plan, g.status, g.subdomain,
		       g.node, COALESCE(g.container_id,''), g.ram_mb, g.cpu_cores,
		       COALESCE(g.port,0), g.created_at
		FROM game_servers g
		LEFT JOIN users u ON u.id = g.user_id
		ORDER BY g.created_at DESC
	`
	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ServerInfo
	for rows.Next() {
		var s ServerInfo
		if err := rows.Scan(&s.ID, &s.Owner, &s.Game, &s.Plan, &s.Status, &s.Subdomain,
			&s.Node, &s.ContainerID, &s.RAMMb, &s.CPUCores, &s.Port, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// GetServerContainer retourne (container_id, node) d'un serveur par son ID.
func GetServerContainer(ctx context.Context, db *pgxpool.Pool, id string) (containerID, node string, err error) {
	const q = `SELECT COALESCE(container_id,''), node FROM game_servers WHERE id = $1`
	err = db.QueryRow(ctx, q, id).Scan(&containerID, &node)
	return
}

// DeleteServer supprime la ligne d'un serveur de la table game_servers (action admin).
// La suppression du container Docker est faite séparément côté collector.
func DeleteServer(ctx context.Context, db *pgxpool.Pool, id string) error {
	_, err := db.Exec(ctx, `DELETE FROM game_servers WHERE id = $1`, id)
	return err
}
