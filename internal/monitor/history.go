package monitor

import (
	"sync"
	"time"
)

// Sample = un point d'historique agrégé (cluster) pour les graphiques admin.
type Sample struct {
	T            time.Time `json:"t"`
	CPUPercent   float64   `json:"cpu_percent"`   // charge CPU agrégée / cœurs totaux
	MemUsedMB    float64   `json:"mem_used_mb"`   // RAM utilisée (somme nodes)
	MemTotalMB   float64   `json:"mem_total_mb"`  // RAM totale (somme nodes)
	Players      int       `json:"players"`       // joueurs en ligne (serveurs gérés)
	ContainersUp int       `json:"containers_up"` // containers running
	NetRxMBs     float64   `json:"net_rx_mbs"`    // débit entrant MB/s
	NetTxMBs     float64   `json:"net_tx_mbs"`    // débit sortant MB/s
}

// History = ring buffer borné des derniers échantillons (séries temporelles courtes).
// Suffisant pour des tendances court terme sans table DB.
type History struct {
	mu      sync.RWMutex
	samples []Sample
	max     int
}

// NewHistory crée un buffer conservant au plus max échantillons.
func NewHistory(max int) *History {
	if max <= 0 {
		max = 360
	}
	return &History{samples: make([]Sample, 0, max), max: max}
}

// Record dérive un échantillon agrégé d'un snapshot et l'ajoute au buffer.
func (h *History) Record(snap *Snapshot) {
	if snap == nil {
		return
	}
	var totCores int
	var memUsed, memTotal float64
	for _, n := range snap.Nodes {
		totCores += n.CPUCount
		memUsed += n.MemUsedMB
		memTotal += n.MemTotalMB
	}
	var aggCPU float64
	var players int
	for _, c := range snap.Containers {
		aggCPU += c.CPUPercent
		if c.HasPlayers {
			players += c.PlayersOn
		}
	}
	cpuPct := 0.0
	if totCores > 0 {
		cpuPct = aggCPU / float64(totCores)
		if cpuPct > 100 {
			cpuPct = 100
		}
	}

	s := Sample{
		T:            snap.Timestamp,
		CPUPercent:   cpuPct,
		MemUsedMB:    memUsed,
		MemTotalMB:   memTotal,
		Players:      players,
		ContainersUp: snap.Host.ContainerUp,
		NetRxMBs:     snap.Host.NetRxMBs,
		NetTxMBs:     snap.Host.NetTxMBs,
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.samples = append(h.samples, s)
	if len(h.samples) > h.max {
		// décale en supprimant les plus anciens (garde la capacité)
		copy(h.samples, h.samples[len(h.samples)-h.max:])
		h.samples = h.samples[:h.max]
	}
}

// Snapshot retourne une copie des échantillons (du plus ancien au plus récent).
func (h *History) Snapshot() []Sample {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Sample, len(h.samples))
	copy(out, h.samples)
	return out
}
