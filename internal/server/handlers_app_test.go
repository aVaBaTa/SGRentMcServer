package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/aVaBaTa/SGRentMcServer/internal/config"
)

// Le téléchargement de l'application est le seul endroit où un jeton GitHub
// circule : on vérifie ici, sans réseau, que (1) sans session → 401, (2) un
// pseudo hors AdminUsers → 403, (3) l'admin reçoit exactement les octets de
// l'archive avec les bons en-têtes, (4) le jeton part vers l'API GitHub mais
// JAMAIS vers l'hôte de stockage de la redirection, (5) /app/status ne révèle
// la release qu'aux autorisés, (6) ?next= n'accepte qu'un chemin relatif.

const testSecret = "test-jwt-secret"

func newTestServer(t *testing.T) *Server {
	t.Helper()
	return &Server{cfg: &config.Config{
		JWTSecret:   testSecret,
		AdminUsers:  []string{"avabata"},
		GitHubToken: "ghp_test_token",
		Env:         "test",
	}}
}

func sessionCookie(t *testing.T, username string) *http.Cookie {
	t.Helper()
	tok, err := auth.GenerateToken("uid-1", "123", username, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "token", Value: tok}
}

// fakeGitHub : /releases/latest, puis l'asset qui redirige (302) vers un second
// serveur « stockage » qui refuse toute requête portant un Authorization.
func fakeGitHub(t *testing.T, payload []byte) (api *httptest.Server, storage *httptest.Server, authSeenAtStorage *bool) {
	t.Helper()
	seen := false
	storage = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			seen = true
			http.Error(w, "token leaked to storage", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(payload)
	}))
	api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ghp_test_token" {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v0.1.0", "name": "SGCoreLink v0.1.0", "published_at": "2026-09-27T20:29:00Z",
				"assets": []map[string]any{
					{"id": 1, "name": "corelink-viewer.exe", "size": 5},
					{"id": 2, "name": "SGCoreLink-v0.1.0-win64.zip", "size": len(payload)},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/releases/assets/2"):
			if r.Header.Get("Accept") != "application/octet-stream" {
				http.Error(w, "bad accept", http.StatusBadRequest)
				return
			}
			http.Redirect(w, r, storage.URL+"/blob", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	return api, storage, &seen
}

func resetAppCache() {
	appRelCache.Lock()
	appRelCache.rel, appRelCache.at = nil, time.Time{}
	appRelCache.Unlock()
}

func TestAppDownload_AccessControl(t *testing.T) {
	s := newTestServer(t)
	api, storage, _ := fakeGitHub(t, []byte("zip"))
	defer api.Close()
	defer storage.Close()
	githubAPI = api.URL
	resetAppCache()

	cases := []struct {
		name   string
		cookie *http.Cookie
		want   int
	}{
		{"sans session", nil, http.StatusUnauthorized},
		{"pseudo non autorise", sessionCookie(t, "quelquun"), http.StatusForbidden},
		{"admin", sessionCookie(t, "aVaBaTa"), http.StatusOK}, // casse différente : EqualFold
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/app/download", nil)
			if c.cookie != nil {
				req.AddCookie(c.cookie)
			}
			rr := httptest.NewRecorder()
			s.handleAppDownload(rr, req)
			if rr.Code != c.want {
				t.Fatalf("code = %d, attendu %d (%s)", rr.Code, c.want, rr.Body.String())
			}
			if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
				t.Fatalf("Cache-Control = %q : no-store obligatoire (cache Cloudflare)", cc)
			}
		})
	}
}

func TestAppDownload_StreamsArchiveWithoutLeakingToken(t *testing.T) {
	s := newTestServer(t)
	payload := []byte("PK\x03\x04 contenu de test de l archive")
	api, storage, leaked := fakeGitHub(t, payload)
	defer api.Close()
	defer storage.Close()
	githubAPI = api.URL
	resetAppCache()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/app/download", nil)
	req.AddCookie(sessionCookie(t, "avabata"))
	rr := httptest.NewRecorder()
	s.handleAppDownload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code = %d : %s", rr.Code, rr.Body.String())
	}
	if *leaked {
		t.Fatal("le jeton GitHub a été envoyé à l hôte de stockage")
	}
	body, _ := io.ReadAll(rr.Body)
	if string(body) != string(payload) {
		t.Fatalf("corps = %q, attendu l archive", body)
	}
	if cd := rr.Header().Get("Content-Disposition"); cd != `attachment; filename="SGCoreLink-v0.1.0-win64.zip"` {
		t.Fatalf("Content-Disposition = %q", cd)
	}
	if cl := rr.Header().Get("Content-Length"); cl != fmt.Sprintf("%d", len(payload)) {
		t.Fatalf("Content-Length = %q", cl)
	}
}

func TestAppStatus_RevealsReleaseOnlyToAllowed(t *testing.T) {
	s := newTestServer(t)
	api, storage, _ := fakeGitHub(t, []byte("zip"))
	defer api.Close()
	defer storage.Close()
	githubAPI = api.URL
	resetAppCache()

	get := func(c *http.Cookie) map[string]any {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/app/status", nil)
		if c != nil {
			req.AddCookie(c)
		}
		rr := httptest.NewRecorder()
		s.handleAppStatus(rr, req)
		var out map[string]any
		json.Unmarshal(rr.Body.Bytes(), &out)
		return out
	}
	if o := get(nil); o["logged_in"] != false || o["allowed"] != false || o["release"] != nil {
		t.Fatalf("anonyme : %v", o)
	}
	if o := get(sessionCookie(t, "quelquun")); o["logged_in"] != true || o["allowed"] != false || o["release"] != nil {
		t.Fatalf("non autorise : %v", o)
	}
	o := get(sessionCookie(t, "avabata"))
	rel, _ := o["release"].(map[string]any)
	if o["allowed"] != true || rel == nil || rel["tag"] != "v0.1.0" || rel["asset_name"] != "SGCoreLink-v0.1.0-win64.zip" {
		t.Fatalf("admin : %v", o)
	}
}

func TestAppDownload_MissingTokenIsExplicit(t *testing.T) {
	s := newTestServer(t)
	s.cfg.GitHubToken = ""
	resetAppCache()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/app/download", nil)
	req.AddCookie(sessionCookie(t, "avabata"))
	rr := httptest.NewRecorder()
	s.handleAppDownload(rr, req)
	if rr.Code != http.StatusServiceUnavailable || !strings.Contains(rr.Body.String(), "GITHUB_TOKEN") {
		t.Fatalf("code = %d, corps = %q", rr.Code, rr.Body.String())
	}
}

func TestDiscordLogin_NextOnlyRelativePath(t *testing.T) {
	s := newTestServer(t)
	s.discordOAuth = auth.NewDiscordOAuth("id", "secret", "http://localhost/cb")
	for _, c := range []struct {
		next   string
		cookie bool
	}{
		{"/telecharger", true},
		{"/dashboard?x=1", true},
		{"https://evil.example/phish", false},
		{"//evil.example", false},
		{"", false},
	} {
		req := httptest.NewRequest(http.MethodGet, "/auth/discord?next="+c.next, nil)
		rr := httptest.NewRecorder()
		s.handleDiscordLogin(rr, req)
		got := false
		for _, ck := range rr.Result().Cookies() {
			if ck.Name == "oauth_next" && ck.Value == c.next {
				got = true
			}
		}
		if got != c.cookie {
			t.Fatalf("next=%q : cookie pose = %v, attendu %v", c.next, got, c.cookie)
		}
	}
}
