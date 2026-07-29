package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type DockerNode struct {
	ID   string
	Host string
}

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string

	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string

	JWTSecret string

	StripeSecretKey     string
	StripeWebhookSecret string
	BTCPayServerURL     string
	BTCPayAPIKey        string

	PayPalClientID string
	PayPalSecret   string
	PayPalEnv      string // sandbox | live

	// DOCKER_NODES="node1=unix:///var/run/docker.sock,node2=tcp://192.168.1.11:2376"
	DockerNodes []DockerNode

	// PrimaryNode : node où sont ÉPINGLÉS les jeux à IP directe (Hytale, Satisfactory…).
	// Raison : le routeur ne peut rediriger une plage de ports que vers UNE machine ;
	// en épinglant ces jeux sur un seul node, le NAT reste simple (plage → 1 IP LAN).
	// Les jeux mc-router (Minecraft) restent load-balancés (routage par hostname OK cross-node).
	PrimaryNode string

	ServersDomain string // ex: servers.vbt-prog.com (pour mc-router hostname)
	MCNetwork     string // réseau Docker partagé avec mc-router

	// SeedDir : dossier (monté dans le container API) contenant les fichiers de jeu
	// pré-téléchargés, par jeu : <SeedDir>/<gameID>/<fichier>. Ex. Hytale :
	// <SeedDir>/hytale/{HytaleServer.jar,Assets.zip}. Évite le re-téléchargement (et
	// son OAuth downloader) à chaque création — le volume est seedé avant le start.
	// Vide → comportement historique (le container télécharge lui-même).
	SeedDir string

	// MetricsExcludeUsers : usernames (minuscule) exclus des métriques business /admin
	// (compte proprio/tests). Les serveurs/paiements/users de ces comptes ne comptent pas.
	MetricsExcludeUsers []string

	// AdminUsers : usernames (Discord) autorisés à accéder à /admin via leur SESSION
	// (cookie JWT) — nginx les laisse passer sans Basic Auth. Ex. le proprio « avabata ».
	AdminUsers []string

	MCRouterAPI string            // ex: http://mc-router:26666
	NodeAddrs   map[string]string // IP LAN par node pour le routing (node1=10.0.0.2,...)

	// SMTP (SGMail) pour le formulaire de support
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SupportFrom string
	SupportTo   string

	// Ingestion du courrier entrant (Cloudflare Email Worker → backend → mailserver)
	IngestSecret string

	// AdminToken : secret partagé entre le monitor (/admin) et l'API pour les
	// endpoints d'administration internes (création de serveur pour un user,
	// toggle des droits). Vide = endpoints admin désactivés.
	AdminToken string

	// CookieDomain : attribut Domain du cookie de session JWT. Vide = host-only
	// (playrena seulement). « .vbt-prog.com » = session visible par les autres
	// sous-domaines (requis pour le gate de téléchargement calradiacoop).
	CookieDomain string

	// CalradiaReleaseAt : date de sortie publique du mod Calradia-Coop. Avant :
	// téléchargement réservé aux AdminUsers + candidatures early-access approuvées.
	// Après : public. Zéro (parse raté) = toujours restreint.
	CalradiaReleaseAt time.Time

	Env string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sgrentmc?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		DiscordClientID:     getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURL:  getEnv("DISCORD_REDIRECT_URL", "http://localhost:8080/auth/discord/callback"),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),

		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		BTCPayServerURL:     getEnv("BTCPAY_SERVER_URL", ""),
		BTCPayAPIKey:        getEnv("BTCPAY_API_KEY", ""),

		PayPalClientID: getEnv("PAYPAL_CLIENT_ID", ""),
		PayPalSecret:   getEnv("PAYPAL_SECRET", ""),
		PayPalEnv:      getEnv("PAYPAL_ENV", "sandbox"),

		DockerNodes:         parseNodes(getEnv("DOCKER_NODES", "node1=unix:///var/run/docker.sock")),
		PrimaryNode:         getEnv("PRIMARY_NODE", "node1"),
		ServersDomain:       getEnv("SERVERS_DOMAIN", "servers.vbt-prog.com"),
		MCNetwork:           getEnv("MC_NETWORK", "mc-net"),
		SeedDir:             getEnv("SEED_DIR", "/seeds"),
		MetricsExcludeUsers: parseList(getEnv("METRICS_EXCLUDE_USERS", "avabata")),
		AdminUsers:          parseList(getEnv("ADMIN_USERS", "avabata")),
		MCRouterAPI:         getEnv("MC_ROUTER_API", "http://mc-router:26666"),
		NodeAddrs:           parseKV(getEnv("NODE_ADDRS", "node1=10.0.0.2,node2=10.0.0.110")),

		SMTPHost:    getEnv("SMTP_HOST", "mailserver"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", "contact@mcserver.vbt-prog.com"),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		SupportFrom: getEnv("SUPPORT_FROM", "contact@mcserver.vbt-prog.com"),
		SupportTo:   getEnv("SUPPORT_TO", "contact@mcserver.vbt-prog.com"),

		IngestSecret: getEnv("INGEST_SECRET", ""),

		AdminToken: getEnv("ADMIN_TOKEN", ""),

		CookieDomain:      getEnv("COOKIE_DOMAIN", ""),
		CalradiaReleaseAt: parseTime(getEnv("CALRADIA_RELEASE_AT", "2026-07-17T00:00:00-04:00")),

		Env: getEnv("ENV", "development"),
	}
}

// parseTime parse un timestamp RFC3339 ; zéro si invalide (= resté restreint).
func parseTime(raw string) time.Time {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		slog.Warn("invalid time value, using zero", "raw", raw, "err", err)
		return time.Time{}
	}
	return t
}

// parseList parse "a,b,c" en []string (trim + minuscule, vides ignorés).
func parseList(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if v := strings.ToLower(strings.TrimSpace(p)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// parseKV parse "k1=v1,k2=v2" en map.
func parseKV(raw string) map[string]string {
	m := make(map[string]string)
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

// parseNodes parse "node1=host1,node2=host2" en []DockerNode
func parseNodes(raw string) []DockerNode {
	var nodes []DockerNode
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 2)
		if len(parts) == 2 {
			nodes = append(nodes, DockerNode{ID: parts[0], Host: parts[1]})
		}
	}
	return nodes
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
