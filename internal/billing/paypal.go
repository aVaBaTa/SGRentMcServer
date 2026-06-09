package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PayPal client minimal (Orders API v2).
type PayPal struct {
	clientID string
	secret   string
	base     string // https://api-m.sandbox.paypal.com ou https://api-m.paypal.com
	http     *http.Client
}

func NewPayPal(clientID, secret, env string) *PayPal {
	base := "https://api-m.sandbox.paypal.com"
	if env == "live" || env == "production" {
		base = "https://api-m.paypal.com"
	}
	return &PayPal{
		clientID: clientID,
		secret:   secret,
		base:     base,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *PayPal) Configured() bool { return p.clientID != "" && p.secret != "" }

func (p *PayPal) accessToken(ctx context.Context) (string, error) {
	form := url.Values{"grant_type": {"client_credentials"}}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.base+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.SetBasicAuth(p.clientID, p.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paypal token: %s", b)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

type Order struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// CreateOrder crée une commande PayPal pour un montant donné (USD).
// amount ex: "7.00". Retourne l'ID de commande à approuver côté client.
func (p *PayPal) CreateOrder(ctx context.Context, amount, description string) (*Order, error) {
	token, err := p.accessToken(ctx)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{{
			"amount": map[string]any{
				"currency_code": "USD",
				"value":         amount,
			},
			"description": description,
		}},
	}
	return p.doOrder(ctx, token, http.MethodPost, "/v2/checkout/orders", body)
}

// CaptureOrder capture le paiement d'une commande approuvée.
func (p *PayPal) CaptureOrder(ctx context.Context, orderID string) (*Order, error) {
	token, err := p.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	return p.doOrder(ctx, token, http.MethodPost, "/v2/checkout/orders/"+orderID+"/capture", map[string]any{})
}

func (p *PayPal) doOrder(ctx context.Context, token, method, path string, body map[string]any) (*Order, error) {
	buf, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, method, p.base+path, bytes.NewReader(buf))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("paypal %s: %s", path, raw)
	}
	var o Order
	if err := json.Unmarshal(raw, &o); err != nil {
		return nil, err
	}
	return &o, nil
}
