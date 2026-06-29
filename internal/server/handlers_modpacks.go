package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
)

// Modpacks « tout-en-un » (façon launcher FTB/CurseForge) : un pack complet (loader +
// version + mods) installé par l'image itzg. Sources sans clé API : Modrinth (TYPE=MODRINTH)
// et Feed-The-Beast (TYPE=FTBA). Installer un modpack = (re)créer le container du serveur.
const ftbAPI = "https://api.modpacks.ch/public"

const modpackRAMFloorMb = 4096 // les modpacks ont besoin de RAM ; on relève à 4 Go mini.

type modpackHit struct {
	Source      string `json:"source"` // "modrinth" | "ftb"
	ID          string `json:"id"`
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Icon        string `json:"icon"`
	Downloads   int    `json:"downloads"`
	VersionID   string `json:"version_id"`
	VersionName string `json:"version_name"`
}

// handleModpackSearch : GET /servers/{id}/modpacks/search?q=&source=all|modrinth|ftb
func (s *Server) handleModpackSearch(w http.ResponseWriter, r *http.Request) {
	gs, _, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.Game != "minecraft" {
		http.Error(w, "modpacks réservés à Minecraft", http.StatusBadRequest)
		return
	}
	q := r.URL.Query().Get("q")
	src := r.URL.Query().Get("source")
	ctx := r.Context()
	hits := []modpackHit{}

	if src == "" || src == "all" || src == "modrinth" {
		hits = append(hits, s.searchModrinthModpacks(ctx, q)...)
	}
	if src == "" || src == "all" || src == "ftb" {
		hits = append(hits, s.searchFTBModpacks(ctx, q)...)
	}
	respond(w, http.StatusOK, map[string]any{"hits": hits, "current": gs.Modpack})
}

func (s *Server) searchModrinthModpacks(ctx context.Context, q string) []modpackHit {
	index := "relevance"
	if q == "" {
		index = "downloads"
	}
	fj, _ := json.Marshal([][]string{{"project_type:modpack"}})
	u := fmt.Sprintf("%s/search?limit=12&index=%s&query=%s&facets=%s",
		modrinthAPI, index, url.QueryEscape(q), url.QueryEscape(string(fj)))
	var sr struct {
		Hits []struct {
			ProjectID   string `json:"project_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Downloads   int    `json:"downloads"`
			IconURL     string `json:"icon_url"`
		} `json:"hits"`
	}
	if err := modGetJSON(ctx, u, &sr); err != nil {
		return nil
	}
	out := make([]modpackHit, 0, len(sr.Hits))
	for _, h := range sr.Hits {
		out = append(out, modpackHit{
			Source: "modrinth", ID: h.ProjectID, Name: h.Title,
			Summary: h.Description, Icon: h.IconURL, Downloads: h.Downloads,
		})
	}
	return out
}

func (s *Server) searchFTBModpacks(ctx context.Context, q string) []modpackHit {
	// IDs : recherche par terme, sinon les plus populaires.
	listURL := ftbAPI + "/modpack/popular/installs/12"
	if q != "" {
		listURL = fmt.Sprintf("%s/modpack/search/12?term=%s", ftbAPI, url.QueryEscape(q))
	}
	var list struct {
		Packs []int `json:"packs"`
	}
	if err := modGetJSON(ctx, listURL, &list); err != nil {
		return nil
	}
	out := []modpackHit{}
	for i, id := range list.Packs {
		if i >= 6 { // borne le nombre d'appels détails (N+1)
			break
		}
		var d struct {
			Name     string `json:"name"`
			Synopsis string `json:"synopsis"`
			Installs int    `json:"installs"`
			Art      []struct {
				URL  string `json:"url"`
				Type string `json:"type"`
			} `json:"art"`
			Versions []struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				Updated int64  `json:"updated"`
			} `json:"versions"`
		}
		if err := modGetJSON(ctx, fmt.Sprintf("%s/modpack/%d", ftbAPI, id), &d); err != nil || len(d.Versions) == 0 {
			continue
		}
		// Version la plus récente.
		sort.Slice(d.Versions, func(a, b int) bool { return d.Versions[a].Updated > d.Versions[b].Updated })
		icon := ""
		for _, a := range d.Art {
			if a.Type == "square" {
				icon = a.URL
				break
			}
		}
		if icon == "" && len(d.Art) > 0 {
			icon = d.Art[0].URL
		}
		out = append(out, modpackHit{
			Source: "ftb", ID: fmt.Sprintf("%d", id), Name: d.Name,
			Summary: d.Synopsis, Icon: icon, Downloads: d.Installs,
			VersionID: fmt.Sprintf("%d", d.Versions[0].ID), VersionName: d.Versions[0].Name,
		})
	}
	return out
}

// normalizeLoaderName mappe un nom de loader (modrinth/FTB) vers paper/fabric/forge.
func normalizeLoaderName(l string) string {
	l = strings.ToLower(l)
	switch {
	case strings.Contains(l, "fabric"), strings.Contains(l, "quilt"):
		return "fabric"
	case strings.Contains(l, "forge"): // inclut neoforge
		return "forge"
	}
	return ""
}

// resolveModpackMeta détermine le loader réel (fabric/forge) ET la version de Minecraft
// d'un modpack → bon dossier /mods + recherche de mods + bonne image Java. Valeurs vides
// si indéterminées.
func (s *Server) resolveModpackMeta(ctx context.Context, source, id string) (loader, mcVersion string) {
	switch source {
	case "modrinth":
		var versions []struct {
			Loaders      []string `json:"loaders"`
			GameVersions []string `json:"game_versions"`
		}
		if err := modGetJSON(ctx, modrinthAPI+"/project/"+url.PathEscape(id)+"/version", &versions); err == nil {
			for _, v := range versions {
				for _, l := range v.Loaders {
					if loader == "" {
						loader = normalizeLoaderName(l)
					}
				}
				if mcVersion == "" {
					for _, gv := range v.GameVersions {
						if isReleaseVersion(gv) {
							mcVersion = gv
							break
						}
					}
				}
				if loader != "" && mcVersion != "" {
					break
				}
			}
		}
	case "ftb":
		var d struct {
			Versions []struct {
				Targets []struct {
					Type    string `json:"type"`
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"targets"`
			} `json:"versions"`
		}
		if err := modGetJSON(ctx, fmt.Sprintf("%s/modpack/%s", ftbAPI, url.PathEscape(id)), &d); err == nil {
			for _, v := range d.Versions {
				for _, tg := range v.Targets {
					if tg.Type == "modloader" && loader == "" {
						loader = normalizeLoaderName(tg.Name)
					}
					if tg.Type == "game" && tg.Name == "minecraft" && mcVersion == "" {
						mcVersion = tg.Version
					}
				}
				if loader != "" && mcVersion != "" {
					break
				}
			}
		}
	}
	return loader, mcVersion
}

// isReleaseVersion : true pour une version MC « stable » (ex. 1.21.1, 26.1.2), false pour
// les snapshots/pre (24w..., 1.21-rc1) qu'on évite pour choisir l'image Java.
func isReleaseVersion(v string) bool {
	if v == "" {
		return false
	}
	for _, c := range v {
		if !(c >= '0' && c <= '9') && c != '.' {
			return false
		}
	}
	return true
}

// handleModpackInstall : POST /servers/{id}/modpacks/install {source,id,version_id}
// → (re)crée le serveur depuis le modpack. source="none" → retire le modpack (retour Paper).
func (s *Server) handleModpackInstall(w http.ResponseWriter, r *http.Request) {
	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if gs.Game != "minecraft" {
		http.Error(w, "réservé à Minecraft", http.StatusBadRequest)
		return
	}
	var b struct {
		Source    string `json:"source"`
		ID        string `json:"id"`
		VersionID string `json:"version_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "corps invalide", http.StatusBadRequest)
		return
	}

	var modpack string
	switch b.Source {
	case "none", "":
		modpack = "" // retour à un serveur classique (loader actuel)
	case "ftb", "modrinth":
		if b.ID == "" {
			http.Error(w, "id requis", http.StatusBadRequest)
			return
		}
		modpack = b.Source + ":" + b.ID + ":" + b.VersionID
	default:
		http.Error(w, "source inconnue", http.StatusBadRequest)
		return
	}

	plan, err := servers.GetPlan(gs.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Modpacks : plancher RAM (gourmands) + plancher CPU (génération de chunks multi-cœur).
	if modpack != "" {
		changed := false
		if gs.RAMMb < modpackRAMFloorMb {
			gs.RAMMb = modpackRAMFloorMb
			changed = true
		}
		if gs.CPUCores < servers.MinCPUForModded {
			gs.CPUCores = servers.MinCPUForModded
			changed = true
		}
		if changed {
			s.serverRepo.UpdateResources(r.Context(), gs.ID, gs.RAMMb, gs.CPUCores)
		}
		// Loader + version MC réels du pack → bon dossier /mods, recherche de mods, et
		// surtout la bonne image Java (1.21 pack → java21 ; pack 26.x → java25).
		lo, mcv := s.resolveModpackMeta(r.Context(), b.Source, b.ID)
		if lo != "" {
			gs.Loader = lo
			s.serverRepo.UpdateLoader(r.Context(), gs.ID, lo)
		}
		if mcv != "" {
			gs.Version = mcv
			s.serverRepo.UpdateVersion(r.Context(), gs.ID, mcv)
		}
	}

	gs.Modpack = modpack
	s.serverRepo.UpdateModpack(r.Context(), gs.ID, modpack)
	s.serverRepo.UpdateStatus(r.Context(), gs.ID, "creating")
	gs.Status = "creating"
	go s.recreateServer(gs, node, plan, gs.Version)

	respond(w, http.StatusOK, gs)
}
