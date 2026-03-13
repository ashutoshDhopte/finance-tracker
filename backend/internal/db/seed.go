package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdminUser(pool *pgxpool.Pool, username, password string) error {
	var exists bool
	err := pool.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check admin user: %w", err)
	}

	if exists {
		log.Printf("admin user %q already exists", username)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = pool.Exec(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username, string(hash),
	)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}

	log.Printf("admin user %q created", username)
	return nil
}
