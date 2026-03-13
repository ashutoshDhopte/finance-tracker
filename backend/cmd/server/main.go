package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/ash/finance-tracker/backend/internal/api"
	"github.com/ash/finance-tracker/backend/internal/config"
	"github.com/ash/finance-tracker/backend/internal/db"
	"github.com/ash/finance-tracker/backend/internal/services/dedup"
	gmailsvc "github.com/ash/finance-tracker/backend/internal/services/gmail"
	"github.com/ash/finance-tracker/backend/internal/services/parser"
	"github.com/ash/finance-tracker/backend/internal/services/scheduler"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Run DB migrations
	migrationsPath := findMigrationsPath()
	if err := db.RunMigrations(cfg.DB.URL, migrationsPath); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Connect to database
	pool, err := db.Connect(cfg.DB.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Seed admin user
	if err := db.SeedAdminUser(pool, cfg.Admin.Username, cfg.Admin.Password); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	// Initialize services
	parserSvc := parser.NewService(cfg.Ollama.URL)
	dedupSvc := dedup.NewService(pool)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Gmail poller if credentials exist
	var gmailService *gmailsvc.Service
	sched := scheduler.New()
	if fileExists(cfg.Gmail.CredentialsFile) && fileExists(cfg.Gmail.TokenFile) {
		svc, err := gmailsvc.NewService(pool, parserSvc, dedupSvc,
			cfg.Gmail.CredentialsFile, cfg.Gmail.TokenFile,
			cfg.Gmail.Query, cfg.Gmail.PollInterval)
		if err != nil {
			log.Printf("gmail service init failed (will run without email polling): %v", err)
		} else {
			gmailService = svc
			sched.Add("gmail-poller", gmailService.Start)
		}
	} else {
		log.Println("gmail credentials not found, email polling disabled")
		log.Println("to enable: place credentials.json and token.json in the working directory")
	}

	go sched.Start(ctx)

	// Start HTTP server
	router := api.NewRouter(cfg, pool, parserSvc, gmailService)
	log.Printf("starting server on :%s", cfg.Server.Port)

	go func() {
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	cancel()
}

func findMigrationsPath() string {
	candidates := []string{
		"migrations",
		"internal/db/migrations",
		"backend/internal/db/migrations",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "migrations"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
