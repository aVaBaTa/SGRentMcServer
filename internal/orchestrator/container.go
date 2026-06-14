package orchestrator

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

// PortBinding décrit un port à publier : port hôte → port interne du container,
// pour un protocole donné.
type PortBinding struct {
	HostPort int
	Internal int
	Proto    string // "tcp" | "udp"
}

// ServerSpec décrit un serveur de jeu à créer.
type ServerSpec struct {
	ContainerName string
	Image         string // ex: "itzg/minecraft-server:latest"
	RAMMb         int64
	CPUCores      float64
	Ports         []PortBinding // ports à publier (accès direct IP:port)
	DataPath      string        // point de montage du volume nommé (ex: "/data", "/config")
	Subdomain     string        // sous-domaine
	RouterHost    string        // hostname complet pour mc-router ; vide → pas de label mc-router
	RouterPort    int           // port interne ciblé par mc-router (ex: 25565)
	Network       string        // réseau Docker partagé avec mc-router (ex: mc-net)
	EnvVars       []string
	OpenStdin     bool          // garder stdin ouvert (console interactive, ex. Hytale)
}

func (n *Node) CreateServer(ctx context.Context, spec ServerSpec) (string, error) {
	// Pull l'image si absente
	if err := n.pullImage(ctx, spec.Image); err != nil {
		return "", fmt.Errorf("pull image: %w", err)
	}

	// Construit les ports exposés + bindings hôte à partir de spec.Ports.
	// Chaque jeu décide de ses ports/protocoles (cf. internal/servers/games.go).
	exposed := nat.PortSet{}
	bindings := nat.PortMap{}
	for _, p := range spec.Ports {
		cp := nat.Port(fmt.Sprintf("%d/%s", p.Internal, p.Proto))
		exposed[cp] = struct{}{}
		bindings[cp] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(p.HostPort)}}
	}

	labels := map[string]string{
		"sgrent.managed":   "true",
		"sgrent.ram_mb":    strconv.FormatInt(spec.RAMMb, 10),
		"sgrent.subdomain": spec.Subdomain,
	}
	// Label mc-router : route hostname → ce container (Minecraft uniquement).
	if spec.RouterHost != "" {
		labels["mc-router.host"] = spec.RouterHost
		labels["mc-router.port"] = strconv.Itoa(spec.RouterPort)
	}

	cfg := &container.Config{
		Image:        spec.Image,
		Labels:       labels,
		Env:          spec.EnvVars,
		ExposedPorts: exposed,
		// Console interactive (Hytale) : on garde stdin ouvert pour pouvoir envoyer
		// des commandes au serveur (ex. `discovery link <token>`) via attach.
		OpenStdin: spec.OpenStdin,
		StdinOnce: false,
	}

	// Volume nommé persistant pour les données du serveur (monde, configs, mods).
	// Conservé même si le container est supprimé (cf. RemoveServer).
	dataVolume := spec.ContainerName + "-data"

	hostCfg := &container.HostConfig{
		PortBindings: bindings,
		Binds:        []string{dataVolume + ":" + spec.DataPath},
		Resources: container.Resources{
			Memory:     containerMemBytes(spec.RAMMb),
			MemorySwap: containerMemBytes(spec.RAMMb),
			NanoCPUs:   int64(spec.CPUCores * 1e9),
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	// Rejoint le réseau partagé avec mc-router (pour le routing par hostname).
	netCfg := &network.NetworkingConfig{}
	if spec.Network != "" {
		netCfg.EndpointsConfig = map[string]*network.EndpointSettings{
			spec.Network: {},
		}
	}

	resp, err := n.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, spec.ContainerName)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}
	return resp.ID, nil
}

// CopyFileToContainer copie un fichier de l'hôte (API) dans le container, sous destDir.
// Utilisé pour seeder un volume avec des fichiers de jeu pré-téléchargés AVANT le
// démarrage (cf. GameDef.SeedFiles). Fonctionne sur node local ou distant (le client
// Docker transporte le flux). On stream via tar + io.Pipe (pas de gros buffer mémoire,
// Assets.zip peut peser lourd).
func (n *Node) CopyFileToContainer(ctx context.Context, containerID, hostPath, destDir string) error {
	f, err := os.Open(hostPath)
	if err != nil {
		return fmt.Errorf("open seed %s: %w", hostPath, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat seed %s: %w", hostPath, err)
	}

	pr, pw := io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		hdr := &tar.Header{
			Name: filepath.Base(hostPath),
			Mode: 0o644,
			Size: fi.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(tw, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(tw.Close()) // nil si OK → EOF propre
	}()

	if err := n.cli.CopyToContainer(ctx, containerID, destDir, pr, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy %s → %s: %w", hostPath, destDir, err)
	}
	return nil
}

// UpdateResources change les limites RAM/CPU d'un container à chaud (upgrade/downgrade).
// MemorySwap doit être >= Memory ; on le met égal à Memory (pas de swap) pour éviter
// l'erreur Docker lors d'une augmentation de RAM.
// containerMemBytes calcule la limite mémoire du cgroup à partir du tas JVM
// annoncé (ramMB). La limite dépasse le tas pour laisser de la marge au non-heap
// (metaspace, threads, buffers directs, GC), sinon le cgroup tue la JVM (OOM)
// au démarrage / pendant la génération du monde. Marge = +50 %, minimum +512 Mo.
func containerMemBytes(ramMB int64) int64 {
	overhead := ramMB / 2
	if overhead < 512 {
		overhead = 512
	}
	return (ramMB + overhead) * 1024 * 1024
}

func (n *Node) UpdateResources(ctx context.Context, containerID string, ramMB int64, cpuCores float64) error {
	memBytes := containerMemBytes(ramMB)
	_, err := n.cli.ContainerUpdate(ctx, containerID, container.UpdateConfig{
		Resources: container.Resources{
			Memory:     memBytes,
			MemorySwap: memBytes,
			NanoCPUs:   int64(cpuCores * 1e9),
		},
	})
	return err
}

func (n *Node) StartServer(ctx context.Context, containerID string) error {
	return n.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (n *Node) StopServer(ctx context.Context, containerID string) error {
	timeout := 30
	return n.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (n *Node) RestartServer(ctx context.Context, containerID string) error {
	timeout := 30
	return n.cli.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (n *Node) RemoveServer(ctx context.Context, containerID string) error {
	return n.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: false, // garder les données du monde
	})
}

func (n *Node) GetContainerStatus(ctx context.Context, containerID string) (string, error) {
	info, err := n.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}
	return info.State.Status, nil
}

// Exec lance une commande dans le container et retourne sa sortie combinée
// (stdout+stderr). Utilisé pour rcon-cli (liste des joueurs, console).
func (n *Node) Exec(ctx context.Context, containerID string, cmd []string) (string, error) {
	execID, err := n.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	att, err := n.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach: %w", err)
	}
	defer att.Close()

	var outBuf, errBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&outBuf, &errBuf, att.Reader); err != nil {
		return "", fmt.Errorf("exec read: %w", err)
	}
	out := outBuf.String()
	if out == "" {
		out = errBuf.String()
	}
	return out, nil
}

// SendStdin écrit une commande sur le STDIN du serveur (console interactive, ex.
// Hytale → `discovery link <token>`). Le container doit avoir été créé avec OpenStdin.
// La sortie de la commande apparaît dans les LOGS du serveur (pas de retour direct).
// On n'envoie PAS d'EOF (pas de CloseWrite) : le stdin reste ouvert pour les commandes
// suivantes (StdinOnce=false).
func (n *Node) SendStdin(ctx context.Context, containerID, command string) error {
	resp, err := n.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return fmt.Errorf("attach stdin: %w", err)
	}
	defer resp.Close()
	if _, err := resp.Conn.Write([]byte(command + "\n")); err != nil {
		return fmt.Errorf("write stdin: %w", err)
	}
	return nil
}

// Logs retourne les dernières lignes de logs du container (console lecture seule).
func (n *Node) Logs(ctx context.Context, containerID string, tail int) (string, error) {
	rc, err := n.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(tail),
	})
	if err != nil {
		return "", err
	}
	defer rc.Close()
	var outBuf, errBuf bytes.Buffer
	// Les logs MC peuvent être multiplexés (pas de TTY) ; on démultiplexe.
	if _, err := stdcopy.StdCopy(&outBuf, &errBuf, rc); err != nil {
		// Fallback : flux brut (container avec TTY).
		return outBuf.String() + errBuf.String(), nil
	}
	return outBuf.String() + errBuf.String(), nil
}

func (n *Node) pullImage(ctx context.Context, img string) error {
	rc, err := n.cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	io.Copy(io.Discard, rc) // attendre la fin du pull
	return nil
}

func containerListOpts() container.ListOptions {
	return container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "sgrent.managed=true")),
	}
}
