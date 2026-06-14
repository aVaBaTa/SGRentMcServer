package servers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameServer struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Subdomain   string    `json:"subdomain"`
	Game        string    `json:"game"`
	Plan        string    `json:"plan"`
	Version     string    `json:"version"`
	Node        string    `json:"node"`
	ContainerID string    `json:"container_id"`
	Status      string    `json:"status"`
	RAMMb       int64     `json:"ram_mb"`
	CPUCores    float64   `json:"cpu_cores"`
	Port        int       `json:"port"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, s *GameServer) error {
	const q = `
		INSERT INTO game_servers (user_id, name, subdomain, game, plan, version, node, container_id, status, ram_mb, cpu_cores, port)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, q,
		s.UserID, s.Name, s.Subdomain, s.Game, s.Plan, s.Version,
		s.Node, s.ContainerID, s.Status, s.RAMMb, s.CPUCores, s.Port,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *Repo) ListByUser(ctx context.Context, userID string) ([]*GameServer, error) {
	const q = `
		SELECT id, user_id, name, subdomain, game, plan, version, node, container_id, status, ram_mb, cpu_cores, port, created_at
		FROM game_servers WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*GameServer
	for rows.Next() {
		s := &GameServer{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Subdomain, &s.Game, &s.Plan, &s.Version,
			&s.Node, &s.ContainerID, &s.Status, &s.RAMMb, &s.CPUCores, &s.Port, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// ListAll retourne tous les serveurs (pour la réconciliation des routes au démarrage).
func (r *Repo) ListAll(ctx context.Context) ([]*GameServer, error) {
	const q = `
		SELECT id, user_id, name, subdomain, game, plan, version, node, container_id, status, ram_mb, cpu_cores, port, created_at
		FROM game_servers
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*GameServer
	for rows.Next() {
		s := &GameServer{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Subdomain, &s.Game, &s.Plan, &s.Version,
			&s.Node, &s.ContainerID, &s.Status, &s.RAMMb, &s.CPUCores, &s.Port, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *Repo) GetByID(ctx context.Context, id, userID string) (*GameServer, error) {
	const q = `
		SELECT id, user_id, name, subdomain, game, plan, version, node, container_id, status, ram_mb, cpu_cores, port, created_at
		FROM game_servers WHERE id = $1 AND user_id = $2
	`
	s := &GameServer{}
	err := r.db.QueryRow(ctx, q, id, userID).Scan(
		&s.ID, &s.UserID, &s.Name, &s.Subdomain, &s.Game, &s.Plan, &s.Version,
		&s.Node, &s.ContainerID, &s.Status, &s.RAMMb, &s.CPUCores, &s.Port, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}
	return s, nil
}

// GetByIDAny récupère un serveur par id SANS contrôle de propriétaire (usage admin).
func (r *Repo) GetByIDAny(ctx context.Context, id string) (*GameServer, error) {
	const q = `
		SELECT id, user_id, name, subdomain, game, plan, version, node, container_id, status, ram_mb, cpu_cores, port, created_at
		FROM game_servers WHERE id = $1
	`
	s := &GameServer{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.UserID, &s.Name, &s.Subdomain, &s.Game, &s.Plan, &s.Version,
		&s.Node, &s.ContainerID, &s.Status, &s.RAMMb, &s.CPUCores, &s.Port, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}
	return s, nil
}

// UpdateResources met à jour UNIQUEMENT la RAM/CPU (override admin), sans toucher au plan.
func (r *Repo) UpdateResources(ctx context.Context, id string, ramMB int64, cpuCores float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_servers SET ram_mb=$1, cpu_cores=$2, updated_at=NOW() WHERE id=$3`,
		ramMB, cpuCores, id,
	)
	return err
}

func (r *Repo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_servers SET status=$1, updated_at=NOW() WHERE id=$2`,
		status, id,
	)
	return err
}

func (r *Repo) UpdateContainerID(ctx context.Context, id, containerID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_servers SET container_id=$1, updated_at=NOW() WHERE id=$2`,
		containerID, id,
	)
	return err
}

func (r *Repo) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM game_servers WHERE subdomain=$1)`, subdomain,
	).Scan(&exists)
	return exists, err
}

func (r *Repo) CountByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM game_servers WHERE user_id=$1`, userID,
	).Scan(&n)
	return n, err
}

func (r *Repo) UpdateVersion(ctx context.Context, id, version string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_servers SET version=$1, updated_at=NOW() WHERE id=$2`,
		version, id,
	)
	return err
}

func (r *Repo) UpdatePlan(ctx context.Context, id, plan string, ramMB int64, cpuCores float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_servers SET plan=$1, ram_mb=$2, cpu_cores=$3, updated_at=NOW() WHERE id=$4`,
		plan, ramMB, cpuCores, id,
	)
	return err
}

func (r *Repo) Delete(ctx context.Context, id, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM game_servers WHERE id=$1 AND user_id=$2`,
		id, userID,
	)
	return err
}
