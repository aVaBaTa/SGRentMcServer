package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

const (
	maxEditBytes   = 1 << 20        // 1 Mo : taille max éditable dans l'éditeur texte
	maxUploadBytes = 512 << 20      // 512 Mo : taille max d'un upload (monde, etc.)
)

// handleListFiles : GET /servers/{id}/files?path=/
func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	rel := r.URL.Query().Get("path")
	entries, err := node.ListFiles(r.Context(), gs.ContainerID, rel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]any{"path": cleanRel(rel), "entries": entries})
}

// handleReadFile : GET /servers/{id}/files/content?path=...  (éditeur texte)
func (s *Server) handleReadFile(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	rel := r.URL.Query().Get("path")
	data, truncated, err := node.ReadFile(r.Context(), gs.ContainerID, rel, maxEditBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"path": cleanRel(rel), "content": string(data), "truncated": truncated,
	})
}

// handleWriteFile : PUT /servers/{id}/files/content  {path, content}
func (s *Server) handleWriteFile(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, maxEditBytes+4096)).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := node.WriteFile(r.Context(), gs.ContainerID, body.Path, []byte(body.Content)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleUploadFile : POST /servers/{id}/files/upload?path=DIR  (multipart "file")
func (s *Server) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+(8<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, "fichier trop volumineux ou requête invalide", http.StatusBadRequest)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "fichier manquant", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if hdr.Size > maxUploadBytes {
		http.Error(w, "fichier trop volumineux (max 512 Mo)", http.StatusRequestEntityTooLarge)
		return
	}
	dir := r.URL.Query().Get("path")
	name := path.Base(hdr.Filename) // évite les chemins dans le nom
	dest := path.Join("/", dir, name)
	if err := node.WriteFileReader(r.Context(), gs.ContainerID, dest, hdr.Size, file); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]any{"status": "ok", "name": name, "size": hdr.Size})
}

// handleDownloadFile : GET /servers/{id}/files/download?path=...
// handleDownloadWorld : GET /servers/{id}/world/download → archive .tar.gz du monde
// Minecraft (dossier /data/world). Marche aussi serveur arrêté (CopyFromContainer lit le
// FS). Pour une sauvegarde 100% cohérente, arrêter le serveur d'abord (MC sauvegarde en live).
func (s *Server) handleDownloadWorld(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.Game != "minecraft" {
		http.Error(w, "sauvegarde du monde réservée à Minecraft", http.StatusBadRequest)
		return
	}
	if gs.ContainerID == "" {
		http.Error(w, "serveur introuvable", http.StatusConflict)
		return
	}
	rc, err := node.DownloadDir(r.Context(), gs.ContainerID, "/world")
	if err != nil {
		http.Error(w, "monde introuvable (pas encore généré ?)", http.StatusNotFound)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", gs.Subdomain+"-world.tar.gz"))
	gz := gzip.NewWriter(w)
	defer gz.Close()
	io.Copy(gz, rc)
}

func (s *Server) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	rc, size, name, err := node.DownloadFile(r.Context(), gs.ContainerID, r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	io.Copy(w, rc)
}

// handleFileMkdir : POST /servers/{id}/files/mkdir  {path}
func (s *Server) handleFileMkdir(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := node.Mkdir(r.Context(), gs.ContainerID, body.Path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleDeleteFile : DELETE /servers/{id}/files  {path}
func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "server not running", http.StatusConflict)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := node.DeletePath(r.Context(), gs.ContainerID, body.Path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func cleanRel(rel string) string {
	c := path.Clean("/" + strings.TrimSpace(rel))
	return c
}
