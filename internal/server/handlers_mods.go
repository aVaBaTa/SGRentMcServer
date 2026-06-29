package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
)

// Navigateur de mods/plugins via Modrinth (API publique, sans clé). Couvre les plugins
// Paper ET les mods Fabric/Forge. Recherche → table de choix dans le panel ; install →
// le backend télécharge le bon .jar et l'injecte dans /mods (moddé) ou /plugins (Paper).
const modrinthAPI = "https://api.modrinth.com/v2"

var modHTTP = &http.Client{Timeout: 40 * time.Second}

// effectiveLoader : loader « effectif » pour le navigateur de mods. Un serveur modpack
// dont le loader n'est pas (encore) résolu est traité comme Fabric (cas le plus courant)
// → recherche de mods + dossier /mods plutôt que des plugins Paper.
func effectiveLoader(gs *servers.GameServer) string {
	if servers.IsModded(gs.Loader) {
		return gs.Loader
	}
	if gs.Modpack != "" {
		return "fabric"
	}
	return gs.Loader
}

// modTarget retourne, pour un loader donné : le type de projet Modrinth, la catégorie
// (loader) pour le facet de recherche, les loaders acceptés à l'install, et le dossier cible.
func modTarget(loader string) (projectType, searchCat string, versionLoaders []string, dir string) {
	switch loader {
	case "fabric":
		return "mod", "fabric", []string{"fabric"}, "/mods"
	case "forge":
		return "mod", "forge", []string{"forge"}, "/mods"
	default: // paper
		return "plugin", "paper", []string{"paper", "spigot", "bukkit", "folia"}, "/plugins"
	}
}

func modGetJSON(ctx context.Context, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Playrena/1.0 (mcserver.vbt-prog.com)")
	resp, err := modHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("modrinth: HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// handleModSearch : GET /servers/{id}/mods/search?q= → liste de mods/plugins compatibles.
func (s *Server) handleModSearch(w http.ResponseWriter, r *http.Request) {
	gs, _, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.Game != "minecraft" {
		http.Error(w, "navigateur de mods réservé à Minecraft", http.StatusBadRequest)
		return
	}
	ptype, cat, _, _ := modTarget(effectiveLoader(gs))
	facets := [][]string{{"project_type:" + ptype}, {"categories:" + cat}}
	if gs.Version != "" && gs.Version != "LATEST" {
		facets = append(facets, []string{"versions:" + gs.Version})
	}
	fj, _ := json.Marshal(facets)
	u := fmt.Sprintf("%s/search?limit=20&index=relevance&query=%s&facets=%s",
		modrinthAPI, url.QueryEscape(r.URL.Query().Get("q")), url.QueryEscape(string(fj)))

	var sr struct {
		Hits []struct {
			ProjectID   string `json:"project_id"`
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Downloads   int    `json:"downloads"`
			IconURL     string `json:"icon_url"`
			Author      string `json:"author"`
		} `json:"hits"`
	}
	if err := modGetJSON(r.Context(), u, &sr); err != nil {
		http.Error(w, "Modrinth injoignable", http.StatusBadGateway)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"loader": gs.Loader,
		"kind":   ptype, // "mod" | "plugin"
		"hits":   sr.Hits,
	})
}

// handleModInstall : POST /servers/{id}/mods/install {project_id} → télécharge le .jar
// compatible et l'injecte dans le dossier mods/plugins du serveur (running requis).
func (s *Server) handleModInstall(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.Game != "minecraft" {
		http.Error(w, "réservé à Minecraft", http.StatusBadRequest)
		return
	}
	if gs.ContainerID == "" || gs.Status != "running" {
		http.Error(w, "serveur non démarré", http.StatusConflict)
		return
	}
	var body struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProjectID == "" {
		http.Error(w, "project_id requis", http.StatusBadRequest)
		return
	}

	_, _, loaders, dir := modTarget(effectiveLoader(gs))
	lj, _ := json.Marshal(loaders)
	vu := fmt.Sprintf("%s/project/%s/version?loaders=%s", modrinthAPI, url.PathEscape(body.ProjectID), url.QueryEscape(string(lj)))
	if gs.Version != "" && gs.Version != "LATEST" {
		gj, _ := json.Marshal([]string{gs.Version})
		vu += "&game_versions=" + url.QueryEscape(string(gj))
	}

	var versions []struct {
		Files []struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
			Primary  bool   `json:"primary"`
			Size     int64  `json:"size"`
		} `json:"files"`
	}
	if err := modGetJSON(r.Context(), vu, &versions); err != nil {
		http.Error(w, "Modrinth injoignable", http.StatusBadGateway)
		return
	}
	if len(versions) == 0 {
		http.Error(w, "aucune version compatible avec ce serveur (loader/version)", http.StatusNotFound)
		return
	}
	// Première version (la plus récente) → fichier principal (sinon le 1er).
	files := versions[0].Files
	pick := files[0]
	for _, f := range files {
		if f.Primary {
			pick = f
			break
		}
	}
	if pick.URL == "" || pick.Size <= 0 {
		http.Error(w, "fichier introuvable", http.StatusBadGateway)
		return
	}

	// Télécharge le jar et l'injecte dans le container.
	dl, err := http.NewRequestWithContext(r.Context(), http.MethodGet, pick.URL, nil)
	if err != nil {
		http.Error(w, "requête invalide", http.StatusInternalServerError)
		return
	}
	dl.Header.Set("User-Agent", "Playrena/1.0 (mcserver.vbt-prog.com)")
	resp, err := modHTTP.Do(dl)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		http.Error(w, "téléchargement échoué", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	_ = node.Mkdir(r.Context(), gs.ContainerID, dir) // best-effort (existe déjà en général)
	name := path.Base(pick.Filename)
	dest := path.Join("/", dir, name)
	if err := node.WriteFileReader(r.Context(), gs.ContainerID, dest, pick.Size, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"status": "ok",
		"name":   name,
		"dir":    dir,
		"modded": servers.IsModded(gs.Loader),
	})
}
