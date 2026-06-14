package server

import (
	"net/http"

	"github.com/aVaBaTa/SGRentMcServer/internal/auth"
	"github.com/aVaBaTa/SGRentMcServer/internal/billing"
	"github.com/aVaBaTa/SGRentMcServer/internal/config"
	"github.com/aVaBaTa/SGRentMcServer/internal/mcrouter"
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
	mcRouter     *mcrouter.Client
}

func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client, orch *orchestrator.Orchestrator) *Server {
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
		mcRouter:     mcrouter.New(cfg.MCRouterAPI),
	}
	s.router = chi.NewRouter()
	s.mountMiddleware()
	s.mountRoutes()
	return s
}

// ServeHTTP permet à *Server d'être utilisé directement comme http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
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
		// --- Routes publiques (sans authentification) ---
		r.Post("/support", s.handleSupport)
		r.Post("/mail/ingest", s.handleMailIngest) // Cloudflare Email Worker → livraison locale (protégé par secret)
		r.Get("/promo", s.handlePromo)             // rabais global actif (affichage prix)
		r.Post("/feedback", s.handleFeedback)      // sondage visiteurs (widget « Ton avis ? »)

		// --- Routes admin internes (monitor /admin → API, protégées par X-Admin-Token) ---
		r.Route("/admin", func(r chi.Router) {
			r.Get("/catalog", s.handleAdminCatalog)
			r.Post("/servers", s.handleAdminCreateServer)
			r.Post("/servers/{id}/resources", s.handleAdminUpdateResources)
			r.Post("/users/{id}/unlimited", s.handleAdminSetUnlimited)
			r.Get("/promo", s.handleAdminGetPromo)
			r.Post("/promo", s.handleAdminSetPromo)
			r.Get("/games", s.handleAdminListGames)
			r.Post("/games/{id}/config", s.handleAdminSetGameConfig)
			r.Get("/metrics", s.handleAdminMetrics)
			r.Get("/feedback", s.handleAdminFeedback)
		})

		// --- Routes authentifiées ---
		r.Group(func(r chi.Router) {
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
				r.Get("/{id}/players", s.handleServerPlayers)
				r.Get("/{id}/playerlists", s.handlePlayerLists)
				r.Post("/{id}/players/action", s.handlePlayerAction)
				r.Get("/{id}/logs", s.handleServerLogs)
				r.Get("/{id}/auth", s.handleServerAuth)
				r.Post("/{id}/command", s.handleServerCommand)
				r.Post("/{id}/discovery", s.handleServerDiscovery)
				r.Get("/{id}/files", s.handleListFiles)
				r.Get("/{id}/files/content", s.handleReadFile)
				r.Put("/{id}/files/content", s.handleWriteFile)
				r.Get("/{id}/files/download", s.handleDownloadFile)
				r.Post("/{id}/files/upload", s.handleUploadFile)
				r.Post("/{id}/files/mkdir", s.handleFileMkdir)
				r.Delete("/{id}/files", s.handleDeleteFile)
				r.Post("/{id}/checkout/paypal", s.handleCreatePayPalOrder)
				r.Post("/{id}/checkout/paypal/{orderID}/capture", s.handleCapturePayPalOrder)
			})
		})
	})
}
