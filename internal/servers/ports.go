package servers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// 25565 est réservé à mc-router (routing par hostname). Les serveurs prennent
	// un "port de base" 25566+ pour l'accès direct IP:port.
	portRangeStart = 25566
	// Fin de la plage utilisable (incluse). Toute la plage 25566-26565 doit être
	// ouverte au pare-feu (UDP + TCP).
	portRangeEnd = 26565

	// portBlockSize : chaque serveur se voit attribuer un BLOC contigu de ports
	// [base, base+portBlockSize). TOUS les ports d'un jeu (jeu, query, messaging…)
	// se dérivent de `base` à l'intérieur de SON bloc → deux serveurs voisins ne
	// peuvent jamais se chevaucher, même avec des ports dérivés consécutifs
	// (base+1, base+2, …). 16 est large : aucun jeu actuel n'utilise plus de
	// ~3 ports. (26565-25566+1)/16 = 62 blocs ⇒ 62 serveurs simultanés possibles,
	// très au-delà de la capacité RAM de la machine.
	portBlockSize = 16

	// legacySatisfactoryMsgOffset : ancien décalage du port "reliable messaging"
	// de Satisfactory (base+500, hors bloc). Conservé UNIQUEMENT pour réserver ce
	// port lors de l'allocation tant qu'un ancien container Satisfactory tourne
	// encore avec cet offset (le nouvel offset, dans le bloc, est dans games.go).
	legacySatisfactoryMsgOffset = 500
)

// blockBaseOf retourne la base du bloc contenant `port` (alignée sur
// portBlockSize depuis portRangeStart).
func blockBaseOf(port int) int {
	if port < portRangeStart {
		return portRangeStart
	}
	return portRangeStart + ((port-portRangeStart)/portBlockSize)*portBlockSize
}

// usedPort : un port de base déjà attribué et le jeu qui le détient (le jeu sert
// à savoir s'il faut réserver l'ancien port messaging Satisfactory hors bloc).
type usedPort struct {
	base int
	game string
}

// firstFreeBlockBase retourne la base du premier bloc libre, étant donné les
// ports déjà attribués. Fonction pure (testable sans DB) : c'est ici que vit la
// règle anti-collision.
func firstFreeBlockBase(used []usedPort) (int, error) {
	taken := make(map[int]bool)
	for _, u := range used {
		taken[blockBaseOf(u.base)] = true
		// Compat : un ancien serveur Satisfactory peut encore écouter sur base+500
		// (ancien offset, hors bloc) tant que son container n'a pas été recréé →
		// on réserve aussi ce bloc pour éviter tout chevauchement transitoire.
		if u.game == "satisfactory" {
			taken[blockBaseOf(u.base+legacySatisfactoryMsgOffset)] = true
		}
	}
	for base := portRangeStart; base+portBlockSize-1 <= portRangeEnd; base += portBlockSize {
		if !taken[base] {
			return base, nil
		}
	}
	return 0, fmt.Errorf("aucun bloc de %d ports libre dans %d-%d", portBlockSize, portRangeStart, portRangeEnd)
}

// NextAvailableBase retourne le port de base d'un bloc de ports libre. Le serveur
// possède alors tout le bloc [base, base+portBlockSize) ; ses ports dérivés se
// calculent depuis `base` (voir GameDef.Ports / GameDef.Env).
func NextAvailableBase(ctx context.Context, db *pgxpool.Pool) (int, error) {
	const q = `SELECT port, game FROM game_servers WHERE port IS NOT NULL ORDER BY port`
	rows, err := db.Query(ctx, q)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var used []usedPort
	for rows.Next() {
		var u usedPort
		if err := rows.Scan(&u.base, &u.game); err != nil {
			return 0, err
		}
		used = append(used, u)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return firstFreeBlockBase(used)
}
