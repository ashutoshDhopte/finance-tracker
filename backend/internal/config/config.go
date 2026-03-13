package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server         ServerConfig
	DB             DBConfig
	JWT            JWTConfig
	Ollama         OllamaConfig
	Gmail          GmailConfig
	Admin          AdminConfig
	AllowedOrigins []string
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
}

type OllamaConfig struct {
	URL string
}

type GmailConfig struct {
	CredentialsFile string
	TokenFile       string
	PollInterval    time.Duration
	Query           string
}

type AdminConfig struct {
	Username string
	Password string
}

func Load() (*Config, error) {
	pollInterval, err := time.ParseDuration(getEnv("GMAIL_POLL_INTERVAL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid GMAIL_POLL_INTERVAL: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("DB_USER", "finance"),
			getEnv("DB_PASSWORD", "finance_secret"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "finance_tracker"),
			getEnv("DB_SSLMODE", "disable"),
		)
	}

	var allowedOrigins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			URL: dbURL,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "X2Kc7LNkR8t2GkqZixpUObbOk7WZ5CMivIYO/rmMP8I="),
		},
		Ollama: OllamaConfig{
			URL: getEnv("OLLAMA_URL", "http://localhost:11434"),
		},
		Gmail: GmailConfig{
			CredentialsFile: getEnv("GMAIL_CREDENTIALS_FILE", "credentials.json"),
			TokenFile:       getEnv("GMAIL_TOKEN_FILE", "token.json"),
			PollInterval:    pollInterval,
			Query:           getEnv("GMAIL_QUERY", "from:(pncalerts@visa.com OR no.reply.alerts@chase.com OR pncalerts@pnc.com)"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "changeme123"),
		},
		AllowedOrigins: allowedOrigins,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
