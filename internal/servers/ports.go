package servers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// 25565 est réservé à mc-router (routing par hostname). Les serveurs
	// prennent 25566+ pour l'accès direct IP:port.
	portRangeStart = 25566
	portRangeEnd   = 26565
)

// NextAvailablePort retourne le prochain port libre dans la plage MC.
func NextAvailablePort(ctx context.Context, db *pgxpool.Pool) (int, error) {
	const q = `SELECT port FROM game_servers WHERE port IS NOT NULL ORDER BY port`
	rows, err := db.Query(ctx, q)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	used := make(map[int]bool)
	for rows.Next() {
		var p int
		if err := rows.Scan(&p); err != nil {
			return 0, err
		}
		used[p] = true
	}

	for p := portRangeStart; p <= portRangeEnd; p++ {
		if !used[p] {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", portRangeStart, portRangeEnd)
}
