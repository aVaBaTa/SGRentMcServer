package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/aVaBaTa/SGRentMcServer/internal/billing"
	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
)

// GET /api/v1/billing/config — infos publiques pour le SDK PayPal côté client.
func (s *Server) handleBillingConfig(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]any{
		"paypal_client_id": s.cfg.PayPalClientID,
		"paypal_env":       s.cfg.PayPalEnv,
		"enabled":          s.paypal.Configured(),
	})
}

// POST /api/v1/servers/{id}/checkout/paypal  body {plan}
func (s *Server) handleCreatePayPalOrder(w http.ResponseWriter, r *http.Request) {
	if !s.paypal.Configured() {
		http.Error(w, "paiement non configuré", http.StatusServiceUnavailable)
		return
	}

	var req upgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	plan, err := servers.GetPlan(req.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if plan.PriceCents <= 0 {
		http.Error(w, "ce plan est gratuit", http.StatusBadRequest)
		return
	}

	gs, _, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	claims := auth.GetClaims(r)

	order, err := s.paypal.CreateOrder(r.Context(), plan.PriceString(),
		"SGRentMc — plan "+plan.Name+" pour "+gs.Name)
	if err != nil {
		slog.Error("paypal create order", "err", err)
		http.Error(w, "échec création commande PayPal", http.StatusBadGateway)
		return
	}

	if err := s.paymentRepo.Create(r.Context(), &billing.Payment{
		UserID:          claims.UserID,
		ServerID:        gs.ID,
		Plan:            plan.Name,
		AmountCents:     plan.PriceCents,
		Currency:        "USD",
		Provider:        "paypal",
		ProviderOrderID: order.ID,
		Status:          "created",
	}); err != nil {
		slog.Error("save payment", "err", err)
	}

	respond(w, http.StatusOK, map[string]string{"order_id": order.ID})
}

// POST /api/v1/servers/{id}/checkout/paypal/{orderID}/capture
func (s *Server) handleCapturePayPalOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	claims := auth.GetClaims(r)

	payment, err := s.paymentRepo.GetByOrderID(r.Context(), orderID, claims.UserID)
	if err != nil {
		http.Error(w, "paiement introuvable", http.StatusNotFound)
		return
	}

	order, err := s.paypal.CaptureOrder(r.Context(), orderID)
	if err != nil {
		slog.Error("paypal capture", "err", err)
		s.paymentRepo.UpdateStatus(r.Context(), payment.ID, "failed")
		http.Error(w, "échec du paiement", http.StatusBadGateway)
		return
	}
	if !strings.EqualFold(order.Status, "COMPLETED") {
		s.paymentRepo.UpdateStatus(r.Context(), payment.ID, "failed")
		http.Error(w, "paiement non complété ("+order.Status+")", http.StatusPaymentRequired)
		return
	}

	// Paiement réussi → appliquer le plan au serveur.
	s.paymentRepo.UpdateStatus(r.Context(), payment.ID, "completed")

	gs, node, err := s.resolveServer(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	plan, _ := servers.GetPlan(payment.Plan)
	if gs.ContainerID != "" {
		node.UpdateResources(r.Context(), gs.ContainerID, plan.RAMMb, plan.CPUCores)
	}
	s.serverRepo.UpdatePlan(r.Context(), gs.ID, plan.Name, plan.RAMMb, plan.CPUCores)

	gs.Plan = plan.Name
	gs.RAMMb = plan.RAMMb
	gs.CPUCores = plan.CPUCores
	respond(w, http.StatusOK, gs)
}
