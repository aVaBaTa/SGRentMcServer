package orchestrator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
)

// Node représente un serveur physique avec son daemon Docker.
type Node struct {
	ID   string
	Host string // ex: "unix:///var/run/docker.sock", "tcp://IP:2376", "ssh://user@IP"
	cli  *client.Client
}

func newNode(id, host string) (*Node, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}

	// Transport SSH (ssh://user@host) : pilote un daemon Docker distant via SSH,
	// sans exposer le daemon en TCP. Nécessite le binaire ssh + une clé dans le container.
	if len(host) > 6 && host[:6] == "ssh://" {
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			return nil, fmt.Errorf("node %s ssh helper: %w", id, err)
		}
		opts = append(opts,
			client.WithHTTPClient(&http.Client{Transport: &http.Transport{DialContext: helper.Dialer}}),
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	} else {
		opts = append(opts, client.WithHost(host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("node %s: %w", id, err)
	}
	return &Node{ID: id, Host: host, cli: cli}, nil
}

// Stats retourne la RAM libre et les CPU disponibles sur le node.
func (n *Node) Stats(ctx context.Context) (*NodeStats, error) {
	info, err := n.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("node %s info: %w", n.ID, err)
	}

	containers, err := n.cli.ContainerList(ctx, containerListOpts())
	if err != nil {
		return nil, fmt.Errorf("node %s container list: %w", n.ID, err)
	}

	var usedRAM int64
	for _, c := range containers {
		if ram, ok := c.Labels["sgrent.ram_mb"]; ok {
			var v int64
			fmt.Sscanf(ram, "%d", &v)
			usedRAM += v
		}
	}

	totalRAMMB := info.MemTotal / 1024 / 1024
	return &NodeStats{
		NodeID:      n.ID,
		TotalRAMMB:  totalRAMMB,
		UsedRAMMB:   usedRAM,
		FreeRAMMB:   totalRAMMB - usedRAM,
		TotalCPUs:   int64(info.NCPU),
		ActiveCount: int64(len(containers)),
	}, nil
}

type NodeStats struct {
	NodeID      string
	TotalRAMMB  int64
	UsedRAMMB   int64
	FreeRAMMB   int64
	TotalCPUs   int64
	ActiveCount int64
}
