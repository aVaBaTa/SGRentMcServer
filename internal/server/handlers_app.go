package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Application cliente Playrena — le visionneur SGCoreLink que le client installe
// sur son PC pour jouer en streaming. Distribuée sur playrena.vbt-prog.com/telecharger.
//
// Accès : réservé aux AdminUsers (le proprio, « avabata ») pour l'instant — même
// règle que le verrou Calradia, sans liste early-access. Le fichier n'est pas
// stocké ici : c'est la dernière release GitHub du dépôt (privé) aVaBaTa/SGCoreLink,
// récupérée avec GITHUB_TOKEN et relayée au navigateur. Ainsi une nouvelle
// release côté SGCoreLink est servie sans redéploiement de Playrena.
//
// ⚠️ Cache-Control no-store sur le téléchargement (leçon Calradia) : sinon
// Cloudflare garde le zip à l'edge après un téléchargement autorisé et le
// ressert aux anonymes — le verrou serait contourné.

const (
	appRepo        = "aVaBaTa/SGCoreLink"
	appAssetSuffix = "-win64.zip" // l'archive complète ; corelink-viewer.exe seul est ignoré
	appCacheTTL    = 5 * time.Minute
)

// githubAPI : variable (pas const) pour que les tests pointent sur un serveur simulé.
var githubAPI = "https://api.github.com"

type appRelease struct {
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	AssetName   string `json:"asset_name"`
	AssetSize   int64  `json:"asset_size"`
	assetID     int64
}

// Dernière release mise en cache : /app/status est appelé à chaque visite, on ne
// veut pas taper GitHub (limité à 5000 req/h avec jeton) à chaque fois.
var appRelCache struct {
	sync.Mutex
	rel *appRelease
	at  time.Time
}

// appAllowed : ce pseudo peut-il télécharger l'application ?
func (s *Server) appAllowed(username string) bool {
	return username != "" && s.isAdminUser(username)
}

func (s *Server) githubRequest(ctx context.Context, url, accept string, followRedirects bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "playrena-api")
	if s.cfg.GitHubToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.GitHubToken)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	if !followRedirects {
		// Le jeton ne doit partir QUE vers api.github.com : la redirection vers le
		// stockage (URL signée, autre hôte) est suivie à la main, sans en-tête.
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}
	return client.Do(req)
}

// latestAppRelease : dernière release du dépôt et son archive win64 (cache 5 min).
func (s *Server) latestAppRelease(ctx context.Context) (*appRelease, error) {
	appRelCache.Lock()
	if appRelCache.rel != nil && time.Since(appRelCache.at) < appCacheTTL {
		rel := *appRelCache.rel
		appRelCache.Unlock()
		return &rel, nil
	}
	appRelCache.Unlock()

	if s.cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN manquant")
	}
	resp, err := s.githubRequest(ctx, githubAPI+"/repos/"+appRepo+"/releases/latest", "application/vnd.github+json", true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github releases/latest: HTTP %d", resp.StatusCode)
	}
	var raw struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&raw); err != nil {
		return nil, fmt.Errorf("github releases/latest: %w", err)
	}
	rel := &appRelease{Tag: raw.TagName, Name: raw.Name, PublishedAt: raw.PublishedAt}
	for _, a := range raw.Assets {
		if strings.HasSuffix(a.Name, appAssetSuffix) {
			rel.assetID, rel.AssetName, rel.AssetSize = a.ID, a.Name, a.Size
			break
		}
	}
	if rel.assetID == 0 {
		return nil, fmt.Errorf("release %s : aucune archive *%s", raw.TagName, appAssetSuffix)
	}
	appRelCache.Lock()
	appRelCache.rel, appRelCache.at = rel, time.Now()
	appRelCache.Unlock()
	return rel, nil
}

// handleAppStatus : GET /api/v1/app/status (public) — état pour la page
// /telecharger : connecté ? autorisé ? et, si autorisé, la release disponible.
func (s *Server) handleAppStatus(w http.ResponseWriter, r *http.Request) {
	username := s.extractUsername(r)
	out := struct {
		LoggedIn bool        `json:"logged_in"`
		Username string      `json:"username,omitempty"`
		Allowed  bool        `json:"allowed"`
		Release  *appRelease `json:"release,omitempty"`
		Error    string      `json:"error,omitempty"`
	}{
		LoggedIn: username != "",
		Username: username,
		Allowed:  s.appAllowed(username),
	}
	if out.Allowed {
		rel, err := s.latestAppRelease(r.Context())
		if err != nil {
			out.Error = err.Error()
		} else {
			out.Release = rel
		}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	respond(w, http.StatusOK, out)
}

// handleAppDownload : GET /api/v1/app/download — relaie l'archive de la dernière
// release au navigateur, seulement pour un pseudo autorisé.
func (s *Server) handleAppDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	username := s.extractUsername(r)
	if username == "" {
		http.Error(w, "login required", http.StatusUnauthorized)
		return
	}
	if !s.appAllowed(username) {
		http.Error(w, "access restricted", http.StatusForbidden)
		return
	}
	rel, err := s.latestAppRelease(r.Context())
	if err != nil {
		http.Error(w, "release indisponible : "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 1) l'API répond 302 vers le stockage (sans jeton pour la suite)
	assetURL := fmt.Sprintf("%s/repos/%s/releases/assets/%d", githubAPI, appRepo, rel.assetID)
	first, err := s.githubRequest(r.Context(), assetURL, "application/octet-stream", false)
	if err != nil {
		http.Error(w, "github injoignable", http.StatusBadGateway)
		return
	}
	upstream := first
	if first.StatusCode >= 300 && first.StatusCode < 400 {
		loc := first.Header.Get("Location")
		first.Body.Close()
		if loc == "" {
			http.Error(w, "redirection github sans destination", http.StatusBadGateway)
			return
		}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, loc, nil)
		if err != nil {
			http.Error(w, "github injoignable", http.StatusBadGateway)
			return
		}
		upstream, err = (&http.Client{Timeout: 5 * time.Minute}).Do(req)
		if err != nil {
			http.Error(w, "stockage github injoignable", http.StatusBadGateway)
			return
		}
	}
	defer upstream.Body.Close()
	if upstream.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("github a répondu %d", upstream.StatusCode), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", rel.AssetName))
	if rel.AssetSize > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", rel.AssetSize))
	}
	io.Copy(w, upstream.Body)
}
