package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/aVaBaTa/SGRentMcServer/internal/mailer"
)

type supportRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// handleSupport reçoit le formulaire de support (public) et envoie un courriel
// vers la boîte de support (SGMail).
func (s *Server) handleSupport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024) // 16 Ko max
	var req supportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Message = strings.TrimSpace(req.Message)

	if req.Name == "" || req.Email == "" || req.Message == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 || len(req.Subject) > 150 || len(req.Message) > 5000 {
		http.Error(w, "field too long", http.StatusBadRequest)
		return
	}
	addr, err := mail.ParseAddress(req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if req.Subject == "" {
		req.Subject = "(sans objet)"
	}

	mc := mailer.Config{
		Host: s.cfg.SMTPHost, Port: s.cfg.SMTPPort,
		User: s.cfg.SMTPUser, Pass: s.cfg.SMTPPass,
		From: s.cfg.SupportFrom, To: s.cfg.SupportTo,
	}
	if !mc.Enabled() {
		slog.Error("support: SMTP non configuré")
		http.Error(w, "support unavailable", http.StatusServiceUnavailable)
		return
	}

	if err := mc.SendSupport(req.Name, addr.Address, req.Subject, req.Message); err != nil {
		slog.Error("support: échec envoi", "err", err)
		http.Error(w, "send failed", http.StatusBadGateway)
		return
	}

	slog.Info("support: message envoyé", "from", addr.Address, "subject", req.Subject)
	respond(w, http.StatusOK, map[string]string{"status": "sent"})
}
