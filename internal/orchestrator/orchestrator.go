package orchestrator

import (
	"context"
	"fmt"
	"sync"
)

// Orchestrator gère plusieurs nodes Docker et répartit les serveurs.
type Orchestrator struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

func New() *Orchestrator {
	return &Orchestrator{nodes: make(map[string]*Node)}
}

func (o *Orchestrator) AddNode(id, host string) error {
	node, err := newNode(id, host)
	if err != nil {
		return err
	}
	o.mu.Lock()
	o.nodes[id] = node
	o.mu.Unlock()
	return nil
}

// BestNode retourne le node avec le plus de RAM libre (load balancing).
func (o *Orchestrator) BestNode(ctx context.Context, requiredRAMMb int64) (*Node, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var best *Node
	var bestFree int64

	for _, node := range o.nodes {
		stats, err := node.Stats(ctx)
		if err != nil {
			continue // node inaccessible, on skip
		}
		if stats.FreeRAMMB >= requiredRAMMb && stats.FreeRAMMB > bestFree {
			best = node
			bestFree = stats.FreeRAMMB
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no node has enough free RAM (%d MB required)", requiredRAMMb)
	}
	return best, nil
}

// NodeWithCapacity retourne un node précis (épinglage) s'il a assez de RAM libre.
// Utilisé pour les jeux à IP directe qu'on veut sur un node fixe (NAT simple).
func (o *Orchestrator) NodeWithCapacity(ctx context.Context, id string, requiredRAMMb int64) (*Node, error) {
	node, err := o.NodeByID(id)
	if err != nil {
		return nil, err
	}
	stats, err := node.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("node %q unavailable: %w", id, err)
	}
	if stats.FreeRAMMB < requiredRAMMb {
		return nil, fmt.Errorf("pinned node %q full: %d MB free < %d MB required", id, stats.FreeRAMMB, requiredRAMMb)
	}
	return node, nil
}

func (o *Orchestrator) NodeByID(id string) (*Node, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	node, ok := o.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %q not found", id)
	}
	return node, nil
}

// AllStats retourne les stats de tous les nodes.
func (o *Orchestrator) AllStats(ctx context.Context) ([]*NodeStats, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var results []*NodeStats
	for _, node := range o.nodes {
		stats, err := node.Stats(ctx)
		if err != nil {
			continue
		}
		results = append(results, stats)
	}
	return results, nil
}
