package servers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// 25565 est réservé à mc-router (routing par hostname). Les serveurs
	// prennent un port "de base" 25566+ pour l'accès direct IP:port.
	portRangeStart = 25566
	// La plage des ports de base s'arrête à 26065 ; la moitié haute
	// (26066-26565) est réservée aux ports dérivés (ex. reliable messaging
	// Satisfactory = port de base + 500). Ainsi un port dérivé ne peut jamais
	// entrer en collision avec le port de base d'un autre serveur. Toute la
	// plage 25566-26565 doit être ouverte au pare-feu (UDP + TCP).
	portRangeEnd = 26065
)

// NextAvailablePort retourne le prochain port de base libre.
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
