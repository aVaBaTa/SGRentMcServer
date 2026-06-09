package mcrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client pilote l'API REST de itzg/mc-router pour enregistrer/supprimer des routes
// (hostname Minecraft → backend host:port). Permet le routing cross-host (multi-node).
type Client struct {
	base string
	http *http.Client
}

func New(baseURL string) *Client {
	return &Client{base: baseURL, http: &http.Client{Timeout: 8 * time.Second}}
}

func (c *Client) Configured() bool { return c.base != "" }

// Register mappe serverAddress (ex: avabata.servers.vbt-prog.com) → backend (ex: 10.0.0.110:25568).
func (c *Client) Register(ctx context.Context, serverAddress, backend string) error {
	body, _ := json.Marshal(map[string]string{
		"serverAddress": serverAddress,
		"backend":       backend,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/routes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mc-router register %s: status %d", serverAddress, resp.StatusCode)
	}
	return nil
}

// Unregister supprime la route d'un serverAddress.
func (c *Client) Unregister(ctx context.Context, serverAddress string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, c.base+"/routes/"+serverAddress, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
