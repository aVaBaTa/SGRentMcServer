package monitor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AdminClient appelle les endpoints d'administration internes de l'API
// (`/api/v1/admin/*`) avec le secret partagé. Utilisé par le monitor pour, depuis
// /admin, créer un serveur au nom d'un user ou (dés)activer ses droits.
type AdminClient struct {
	base  string // ex: http://mcserver-api:8080
	token string
	hc    *http.Client
}

// NewAdminClient retourne nil si la base ou le token sont absents (feature off).
func NewAdminClient(base, token string) *AdminClient {
	if base == "" || token == "" {
		return nil
	}
	return &AdminClient{base: base, token: token, hc: &http.Client{Timeout: 30 * time.Second}}
}

// do envoie une requête vers l'API admin et relaie statut + corps bruts.
func (c *AdminClient) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("X-Admin-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("API injoignable: %w", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, out, nil
}

func (c *AdminClient) Catalog(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/admin/catalog", nil)
}

func (c *AdminClient) SetUnlimited(ctx context.Context, userID string, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/admin/users/"+userID+"/unlimited", body)
}

func (c *AdminClient) CreateServer(ctx context.Context, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/admin/servers", body)
}

func (c *AdminClient) UpdateResources(ctx context.Context, serverID string, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/admin/servers/"+serverID+"/resources", body)
}

func (c *AdminClient) GetPromo(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/admin/promo", nil)
}

func (c *AdminClient) SetPromo(ctx context.Context, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/admin/promo", body)
}

func (c *AdminClient) ListGames(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/admin/games", nil)
}

func (c *AdminClient) SetGameConfig(ctx context.Context, gameID string, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/admin/games/"+gameID+"/config", body)
}

func (c *AdminClient) Metrics(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/admin/metrics", nil)
}

func (c *AdminClient) Feedback(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/admin/feedback", nil)
}
