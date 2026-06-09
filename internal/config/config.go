package config

import (
	"log/slog"
	"os"
	"strings"

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

	ServersDomain string // ex: servers.vbt-prog.com (pour mc-router hostname)
	MCNetwork     string // réseau Docker partagé avec mc-router

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

		DockerNodes:   parseNodes(getEnv("DOCKER_NODES", "node1=unix:///var/run/docker.sock")),
		ServersDomain: getEnv("SERVERS_DOMAIN", "servers.vbt-prog.com"),
		MCNetwork:     getEnv("MC_NETWORK", "mc-net"),
		Env:           getEnv("ENV", "development"),
	}
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
