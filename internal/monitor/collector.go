package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/sys/unix"
)

// ContainerMetric = métriques d'un container Docker.
type ContainerMetric struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Image      string  `json:"image"`
	Status     string  `json:"status"`
	State      string  `json:"state"`
	Managed    bool    `json:"managed"` // géré par SGRent (label sgrent.managed)
	CPUPercent float64 `json:"cpu_percent"`
	MemUsageMB float64 `json:"mem_usage_mb"`
	MemLimitMB float64 `json:"mem_limit_mb"`
	MemPercent float64 `json:"mem_percent"`
	CPULimit   float64 `json:"cpu_limit"` // cœurs alloués (0 = illimité)
	DiskRwMB   float64 `json:"disk_rw_mb"`
	NetRxMB    float64 `json:"net_rx_mb"`
	NetTxMB    float64 `json:"net_tx_mb"`
	UptimeSecs int64   `json:"uptime_secs"`
}

// HostMetric = métriques de la machine hôte.
type HostMetric struct {
	Hostname       string  `json:"hostname"`
	CPUCount       int     `json:"cpu_count"`
	Load1          float64 `json:"load1"`
	Load5          float64 `json:"load5"`
	Load15         float64 `json:"load15"`
	MemTotalMB     float64 `json:"mem_total_mb"`
	MemUsedMB      float64 `json:"mem_used_mb"`
	MemPercent     float64 `json:"mem_percent"`
	DiskTotalGB    float64 `json:"disk_total_gb"`
	DiskUsedGB     float64 `json:"disk_used_gb"`
	DiskPercent    float64 `json:"disk_percent"`
	ContainerTotal int     `json:"container_total"`
	ContainerUp    int     `json:"container_up"`
	NetRxMBs       float64 `json:"net_rx_mbs"`    // débit entrant MB/s
	NetTxMBs       float64 `json:"net_tx_mbs"`    // débit sortant MB/s
	NetRxTotalGB   float64 `json:"net_rx_total_gb"`
	NetTxTotalGB   float64 `json:"net_tx_total_gb"`
}

type Snapshot struct {
	Host       HostMetric        `json:"host"`
	Containers []ContainerMetric `json:"containers"`
	Timestamp  time.Time         `json:"timestamp"`
}

type Collector struct {
	cli      *client.Client
	hostRoot string // chemin du FS hôte monté (pour statfs disque)

	mu      sync.Mutex
	prevNet *netSample // dernier échantillon réseau (pour calcul du débit)
}

type netSample struct {
	rx, tx uint64
	t      time.Time
}

func NewCollector(hostRoot string) (*Collector, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if hostRoot == "" {
		hostRoot = "/"
	}
	if _, err := os.Stat(hostRoot); err != nil {
		hostRoot = "/"
	}
	return &Collector{cli: cli, hostRoot: hostRoot}, nil
}

// StopContainer force l'arrêt d'un container (action admin).
func (c *Collector) StopContainer(ctx context.Context, id string) error {
	timeout := 10
	return c.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}

// RemoveContainer supprime un container de force (action admin).
func (c *Collector) RemoveContainer(ctx context.Context, id string) error {
	return c.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (c *Collector) Collect(ctx context.Context) (*Snapshot, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("container list: %w", err)
	}

	metrics := make([]ContainerMetric, len(containers))
	var wg sync.WaitGroup
	up := 0

	for i, ct := range containers {
		if ct.State == "running" {
			up++
		}
		wg.Add(1)
		go func(i int, ct types.Container) {
			defer wg.Done()
			metrics[i] = c.containerMetric(ctx, ct)
		}(i, ct)
	}
	wg.Wait()

	// Tri : gérés SGRent d'abord, puis par CPU décroissant
	sort.Slice(metrics, func(a, b int) bool {
		if metrics[a].Managed != metrics[b].Managed {
			return metrics[a].Managed
		}
		return metrics[a].CPUPercent > metrics[b].CPUPercent
	})

	host := c.hostMetric(ctx)
	host.ContainerTotal = len(containers)
	host.ContainerUp = up
	host.NetRxMBs, host.NetTxMBs, host.NetRxTotalGB, host.NetTxTotalGB = c.netRate(metrics)

	return &Snapshot{Host: host, Containers: metrics, Timestamp: time.Now()}, nil
}

func (c *Collector) containerMetric(ctx context.Context, ct types.Container) ContainerMetric {
	name := "—"
	if len(ct.Names) > 0 {
		name = strings.TrimPrefix(ct.Names[0], "/")
	}
	_, managed := ct.Labels["sgrent.managed"]

	m := ContainerMetric{
		ID:      ct.ID[:12],
		Name:    name,
		Image:   ct.Image,
		Status:  ct.Status,
		State:   ct.State,
		Managed: managed,
	}

	if ct.State != "running" {
		return m
	}

	// Disque (couche writable) via inspect avec size
	if info, _, err := c.cli.ContainerInspectWithRaw(ctx, ct.ID, true); err == nil {
		if info.SizeRw != nil {
			m.DiskRwMB = float64(*info.SizeRw) / 1024 / 1024
		}
		if info.HostConfig != nil && info.HostConfig.NanoCPUs > 0 {
			m.CPULimit = float64(info.HostConfig.NanoCPUs) / 1e9
		}
		if t, err := time.Parse(time.RFC3339Nano, info.State.StartedAt); err == nil {
			m.UptimeSecs = int64(time.Since(t).Seconds())
		}
	}

	// Stats live (CPU, RAM, réseau)
	resp, err := c.cli.ContainerStats(ctx, ct.ID, false)
	if err != nil {
		return m
	}
	defer resp.Body.Close()

	var s dockerStats
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return m
	}

	m.CPUPercent = s.cpuPercent()
	used := s.memUsedBytes()
	m.MemUsageMB = float64(used) / 1024 / 1024
	m.MemLimitMB = float64(s.MemoryStats.Limit) / 1024 / 1024
	if s.MemoryStats.Limit > 0 {
		m.MemPercent = float64(used) / float64(s.MemoryStats.Limit) * 100
	}
	rx, tx := s.netBytes()
	m.NetRxMB = float64(rx) / 1024 / 1024
	m.NetTxMB = float64(tx) / 1024 / 1024
	return m
}

func (c *Collector) hostMetric(ctx context.Context) HostMetric {
	h := HostMetric{CPUCount: runtime.NumCPU()}
	// Préfère le hostname de l'hôte (monté) plutôt que celui du container
	if b, err := os.ReadFile(c.hostRoot + "/etc/hostname"); err == nil {
		h.Hostname = strings.TrimSpace(string(b))
	}
	if h.Hostname == "" {
		h.Hostname, _ = os.Hostname()
	}

	// RAM via /proc/meminfo (valeurs hôte vues depuis le container)
	if total, avail, ok := readMemInfo(); ok {
		h.MemTotalMB = float64(total) / 1024
		h.MemUsedMB = float64(total-avail) / 1024
		if total > 0 {
			h.MemPercent = float64(total-avail) / float64(total) * 100
		}
	}

	// Load average
	if l1, l5, l15, ok := readLoadAvg(); ok {
		h.Load1, h.Load5, h.Load15 = l1, l5, l15
	}

	// Disque via statfs sur le FS hôte monté
	var st unix.Statfs_t
	if err := unix.Statfs(c.hostRoot, &st); err == nil {
		bs := float64(st.Bsize)
		total := float64(st.Blocks) * bs
		free := float64(st.Bavail) * bs
		h.DiskTotalGB = total / 1e9
		h.DiskUsedGB = (total - free) / 1e9
		if total > 0 {
			h.DiskPercent = (total - free) / total * 100
		}
	}

	return h
}

// netRate agrège le trafic réseau cumulé de tous les containers et calcule
// le débit (MB/s) entre deux relevés. Représente le trafic des serveurs de jeux.
func (c *Collector) netRate(metrics []ContainerMetric) (rxMBs, txMBs, rxTotalGB, txTotalGB float64) {
	var sumRxMB, sumTxMB float64
	for _, m := range metrics {
		sumRxMB += m.NetRxMB
		sumTxMB += m.NetTxMB
	}
	rxTotalGB = sumRxMB / 1024
	txTotalGB = sumTxMB / 1024

	rxBytes := uint64(sumRxMB * 1048576)
	txBytes := uint64(sumTxMB * 1048576)
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.prevNet != nil {
		dt := now.Sub(c.prevNet.t).Seconds()
		if dt > 0 {
			if rxBytes >= c.prevNet.rx {
				rxMBs = float64(rxBytes-c.prevNet.rx) / 1048576 / dt
			}
			if txBytes >= c.prevNet.tx {
				txMBs = float64(txBytes-c.prevNet.tx) / 1048576 / dt
			}
		}
	}
	c.prevNet = &netSample{rx: rxBytes, tx: txBytes, t: now}
	return
}

// --- Décodage des stats Docker (struct minimale) ---

type dockerStats struct {
	CPUStats    cpuStats `json:"cpu_stats"`
	PreCPUStats cpuStats `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64            `json:"usage"`
		Limit uint64            `json:"limit"`
		Stats map[string]uint64 `json:"stats"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

type cpuStats struct {
	CPUUsage struct {
		TotalUsage  uint64   `json:"total_usage"`
		PercpuUsage []uint64 `json:"percpu_usage"`
	} `json:"cpu_usage"`
	SystemCPUUsage uint64 `json:"system_cpu_usage"`
	OnlineCPUs     uint32 `json:"online_cpus"`
}

func (s *dockerStats) cpuPercent() float64 {
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)
	if sysDelta <= 0 || cpuDelta < 0 {
		return 0
	}
	ncpu := float64(s.CPUStats.OnlineCPUs)
	if ncpu == 0 {
		ncpu = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if ncpu == 0 {
		ncpu = float64(runtime.NumCPU())
	}
	return (cpuDelta / sysDelta) * ncpu * 100
}

// memUsedBytes : usage moins le cache fichier inactif (cgroup v2: inactive_file).
func (s *dockerStats) memUsedBytes() uint64 {
	usage := s.MemoryStats.Usage
	if v, ok := s.MemoryStats.Stats["inactive_file"]; ok && v < usage {
		return usage - v
	}
	if v, ok := s.MemoryStats.Stats["total_inactive_file"]; ok && v < usage {
		return usage - v
	}
	return usage
}

func (s *dockerStats) netBytes() (rx, tx uint64) {
	for _, n := range s.Networks {
		rx += n.RxBytes
		tx += n.TxBytes
	}
	return
}

// --- Lecture /proc ---

func readMemInfo() (totalKB, availKB uint64, ok bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	var haveTotal, haveAvail bool
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			totalKB, haveTotal = v, true
		case "MemAvailable:":
			availKB, haveAvail = v, true
		}
		if haveTotal && haveAvail {
			return totalKB, availKB, true
		}
	}
	return totalKB, availKB, haveTotal
}

func readLoadAvg() (l1, l5, l15 float64, ok bool) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, false
	}
	fields := strings.Fields(string(b))
	if len(fields) < 3 {
		return 0, 0, 0, false
	}
	l1, _ = strconv.ParseFloat(fields[0], 64)
	l5, _ = strconv.ParseFloat(fields[1], 64)
	l15, _ = strconv.ParseFloat(fields[2], 64)
	return l1, l5, l15, true
}
