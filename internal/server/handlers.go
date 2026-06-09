package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"golang.org/x/oauth2"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Auth ---

func (s *Server) handleDiscordLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Env == "production",
		MaxAge:   300,
		SameSite: http.SameSiteLaxMode,
	})
	url := s.discordOAuth.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	// Valider le state anti-CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	// Échanger le code contre un token Discord
	code := r.URL.Query().Get("code")
	token, err := s.discordOAuth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Récupérer les infos utilisateur depuis Discord
	discordUser, err := auth.FetchDiscordUser(r.Context(), token, s.discordOAuth)
	if err != nil {
		http.Error(w, "failed to fetch discord user", http.StatusInternalServerError)
		return
	}

	// Upsert en DB
	user, err := s.userRepo.UpsertFromDiscord(r.Context(), discordUser)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Générer JWT
	jwtToken, err := auth.GenerateToken(user.ID, user.DiscordID, user.Username, s.cfg.JWTSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Env == "production",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	// Rediriger vers le dashboard (frontend)
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	respond(w, http.StatusOK, map[string]string{"msg": "logged out"})
}

// --- User ---

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	user, err := s.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	respond(w, http.StatusOK, user)
}

// --- Helpers ---

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
