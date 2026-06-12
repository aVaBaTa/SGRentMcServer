package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
)

// Domaines mail hébergés localement (évite tout relais ouvert).
var hostedMailDomains = []string{"@vbt-prog.com", "@mcserver.vbt-prog.com", "@playrena.vbt-prog.com"}

// handleMailIngest reçoit un courriel brut (RFC822) poussé par le Cloudflare
// Email Worker et le livre localement dans le mailserver. Permet de RECEVOIR du
// courrier alors que le port 25 entrant est bloqué par le FAI : Cloudflare reçoit,
// le Worker POST ici en HTTPS, et on injecte en livraison locale (port 25 interne).
func (s *Server) handleMailIngest(w http.ResponseWriter, r *http.Request) {
	if s.cfg.IngestSecret == "" || r.Header.Get("X-Ingest-Secret") != s.cfg.IngestSecret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	to := strings.TrimSpace(r.Header.Get("X-Mail-To"))
	from := strings.TrimSpace(r.Header.Get("X-Mail-From"))
	if a, err := mail.ParseAddress(to); err == nil {
		to = a.Address
	}
	if a, err := mail.ParseAddress(from); err == nil {
		from = a.Address
	}
	if to == "" {
		http.Error(w, "missing X-Mail-To", http.StatusBadRequest)
		return
	}

	hosted := false
	for _, d := range hostedMailDomains {
		if strings.HasSuffix(strings.ToLower(to), d) {
			hosted = true
			break
		}
	}
	if !hosted {
		http.Error(w, "recipient not hosted", http.StatusBadRequest)
		return
	}

	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 30*1024*1024)) // 30 Mo max
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	if from == "" {
		from = "noreply@playrena.vbt-prog.com"
	}

	if err := injectLocal(s.cfg.SMTPHost+":25", from, to, raw); err != nil {
		slog.Error("mail ingest: injection échouée", "err", err, "to", to)
		http.Error(w, "delivery failed", http.StatusBadGateway)
		return
	}

	slog.Info("mail ingest: livré", "to", to, "from", from, "bytes", len(raw))
	respond(w, http.StatusOK, map[string]string{"status": "delivered"})
}

// injectLocal remet le message brut au postfix local (port 25 interne, sans auth :
// destination finale = domaine hébergé, donc accepté comme livraison locale).
func injectLocal(addr, from, to string, raw []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer c.Close()
	if err := c.Hello("playrena.vbt-prog.com"); err != nil {
		return fmt.Errorf("hello: %w", err)
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := wc.Write(raw); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return c.Quit()
}
