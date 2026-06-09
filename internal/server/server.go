package server

import (
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/aVaBaTa/SGRentMcServer/internal/billing"
	"github.com/aVaBaTa/SGRentMcServer/internal/config"
	"github.com/aVaBaTa/SGRentMcServer/internal/orchestrator"
	"github.com/aVaBaTa/SGRentMcServer/internal/servers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type Server struct {
	cfg          *config.Config
	db           *pgxpool.Pool
	rdb          *redis.Client
	router       *chi.Mux
	discordOAuth *oauth2.Config
	userRepo     *auth.UserRepo
	orch         *orchestrator.Orchestrator
	serverRepo   *servers.Repo
	paypal       *billing.PayPal
	paymentRepo  *billing.Repo
}

func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client, orch *orchestrator.Orchestrator) http.Handler {
	s := &Server{
		cfg:          cfg,
		db:           db,
		rdb:          rdb,
		discordOAuth: auth.NewDiscordOAuth(cfg.DiscordClientID, cfg.DiscordClientSecret, cfg.DiscordRedirectURL),
		userRepo:     auth.NewUserRepo(db),
		orch:         orch,
		serverRepo:   servers.NewRepo(db),
		paypal:       billing.NewPayPal(cfg.PayPalClientID, cfg.PayPalSecret, cfg.PayPalEnv),
		paymentRepo:  billing.NewRepo(db),
	}
	s.router = chi.NewRouter()
	s.mountMiddleware()
	s.mountRoutes()
	return s.router
}

func (s *Server) mountMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
}

func (s *Server) mountRoutes() {
	s.router.Get("/health", s.handleHealth)

	s.router.Route("/auth", func(r chi.Router) {
		r.Get("/discord", s.handleDiscordLogin)
		r.Get("/discord/callback", s.handleDiscordCallback)
		r.Post("/logout", s.handleLogout)
	})

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(s.cfg.JWTSecret))

		r.Route("/user", func(r chi.Router) {
			r.Get("/me", s.handleGetMe)
		})

		r.Get("/billing/config", s.handleBillingConfig)

		r.Route("/servers", func(r chi.Router) {
			r.Get("/", s.handleListServers)
			r.Post("/", s.handleCreateServer)
			r.Get("/{id}", s.handleGetServer)
			r.Delete("/{id}", s.handleDeleteServer)
			r.Post("/{id}/start", s.handleStartServer)
			r.Post("/{id}/stop", s.handleStopServer)
			r.Post("/{id}/restart", s.handleRestartServer)
			r.Post("/{id}/upgrade", s.handleUpgradeServer)
			r.Post("/{id}/version", s.handleChangeVersion)
			r.Post("/{id}/checkout/paypal", s.handleCreatePayPalOrder)
			r.Post("/{id}/checkout/paypal/{orderID}/capture", s.handleCapturePayPalOrder)
		})
	})
}
