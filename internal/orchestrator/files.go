package orchestrator

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

// dataUID/dataGID : l'image itzg/minecraft-server tourne en uid/gid 1000.
// Les fichiers écrits doivent appartenir à cet utilisateur pour être lisibles.
const (
	dataUID  = 1000
	dataGID  = 1000
	dataRoot = "/data"
)

// FileEntry décrit une entrée d'un répertoire du serveur.
type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime"` // epoch seconds
}

// safeDataPath nettoie un chemin relatif et garantit qu'il reste sous /data
// (anti directory traversal). Retourne le chemin absolu dans le container.
func safeDataPath(rel string) (string, error) {
	clean := path.Clean("/" + strings.TrimSpace(rel)) // force absolu, résout ".."
	full := path.Clean(path.Join(dataRoot, clean))
	if full != dataRoot && !strings.HasPrefix(full, dataRoot+"/") {
		return "", fmt.Errorf("chemin invalide")
	}
	return full, nil
}

// ListFiles liste un répertoire du container (sous /data).
func (n *Node) ListFiles(ctx context.Context, containerID, rel string) ([]FileEntry, error) {
	dir, err := safeDataPath(rel)
	if err != nil {
		return nil, err
	}
	// -lAp : détails, cachés (hors . et ..), '/' sur les dossiers ; mtime epoch.
	out, err := n.Exec(ctx, containerID, []string{"sh", "-c",
		fmt.Sprintf("ls -lAp --time-style=+%%s %s 2>/dev/null", shellQuote(dir))})
	if err != nil {
		return nil, err
	}
	entries := []FileEntry{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "total ") {
			continue
		}
		// perms links owner group size mtime name...
		f := strings.Fields(line)
		if len(f) < 7 {
			continue
		}
		size, _ := strconv.ParseInt(f[4], 10, 64)
		mtime, _ := strconv.ParseInt(f[5], 10, 64)
		name := strings.Join(f[6:], " ")
		isDir := strings.HasPrefix(f[0], "d") || strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")
		// Ignore les liens symboliques cibles (" -> ")
		if i := strings.Index(name, " -> "); i >= 0 {
			name = name[:i]
		}
		if name == "" {
			continue
		}
		entries = append(entries, FileEntry{Name: name, IsDir: isDir, Size: size, MTime: mtime})
	}
	return entries, nil
}

// ReadFile lit jusqu'à maxBytes octets d'un fichier sous /data.
func (n *Node) ReadFile(ctx context.Context, containerID, rel string, maxBytes int64) ([]byte, bool, error) {
	full, err := safeDataPath(rel)
	if err != nil {
		return nil, false, err
	}
	rc, _, err := n.cli.CopyFromContainer(ctx, containerID, full)
	if err != nil {
		return nil, false, err
	}
	defer rc.Close()
	tr := tar.NewReader(rc)
	hdr, err := tr.Next()
	if err != nil {
		return nil, false, fmt.Errorf("fichier introuvable")
	}
	if hdr.FileInfo().IsDir() {
		return nil, false, fmt.Errorf("c'est un dossier")
	}
	limited := io.LimitReader(tr, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	truncated := int64(len(data)) > maxBytes
	if truncated {
		data = data[:maxBytes]
	}
	return data, truncated, nil
}

// tarFileReader lit le contenu du fichier dans un tar et ferme la source à la fin.
type tarFileReader struct {
	tr  *tar.Reader
	src io.Closer
}

func (t *tarFileReader) Read(p []byte) (int, error) { return t.tr.Read(p) }
func (t *tarFileReader) Close() error               { return t.src.Close() }

// DownloadFile retourne un flux du fichier sous /data (téléchargement direct).
func (n *Node) DownloadFile(ctx context.Context, containerID, rel string) (io.ReadCloser, int64, string, error) {
	full, err := safeDataPath(rel)
	if err != nil {
		return nil, 0, "", err
	}
	rc, _, err := n.cli.CopyFromContainer(ctx, containerID, full)
	if err != nil {
		return nil, 0, "", err
	}
	tr := tar.NewReader(rc)
	hdr, err := tr.Next()
	if err != nil {
		rc.Close()
		return nil, 0, "", fmt.Errorf("fichier introuvable")
	}
	if hdr.FileInfo().IsDir() {
		rc.Close()
		return nil, 0, "", fmt.Errorf("c'est un dossier")
	}
	return &tarFileReader{tr: tr, src: rc}, hdr.Size, path.Base(full), nil
}

// WriteFile écrit (crée/écrase) un petit fichier sous /data (configs).
func (n *Node) WriteFile(ctx context.Context, containerID, rel string, content []byte) error {
	return n.WriteFileReader(ctx, containerID, rel, int64(len(content)), bytes.NewReader(content))
}

// WriteFileReader écrit un fichier sous /data en streamant depuis r (taille connue).
// Évite de tout bufferiser en mémoire (utile pour l'upload de mondes volumineux).
func (n *Node) WriteFileReader(ctx context.Context, containerID, rel string, size int64, r io.Reader) error {
	full, err := safeDataPath(rel)
	if err != nil {
		return err
	}
	if full == dataRoot {
		return fmt.Errorf("chemin invalide")
	}
	dir, base := path.Split(full)

	pr, pw := io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		hdr := &tar.Header{
			Name: base, Mode: 0o644, Size: size,
			Uid: dataUID, Gid: dataGID, ModTime: time.Now(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(tw, r); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(tw.Close())
	}()
	return n.cli.CopyToContainer(ctx, containerID, dir, pr, container.CopyToContainerOptions{})
}

// DeletePath supprime un fichier ou dossier (récursif) sous /data.
func (n *Node) DeletePath(ctx context.Context, containerID, rel string) error {
	full, err := safeDataPath(rel)
	if err != nil {
		return err
	}
	if full == dataRoot {
		return fmt.Errorf("impossible de supprimer la racine")
	}
	_, err = n.Exec(ctx, containerID, []string{"rm", "-rf", full})
	return err
}

// Mkdir crée un dossier (récursif) sous /data avec le bon propriétaire.
func (n *Node) Mkdir(ctx context.Context, containerID, rel string) error {
	full, err := safeDataPath(rel)
	if err != nil {
		return err
	}
	if full == dataRoot {
		return fmt.Errorf("chemin invalide")
	}
	_, err = n.Exec(ctx, containerID, []string{"sh", "-c",
		fmt.Sprintf("mkdir -p %s && chown %d:%d %s", shellQuote(full), dataUID, dataGID, shellQuote(full))})
	return err
}

// shellQuote entoure une valeur de quotes simples pour un usage shell sûr.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
