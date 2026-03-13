package dedup

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func GenerateHash(amount float64, date string, merchant string) string {
	normalized := strings.ToLower(strings.TrimSpace(merchant))
	input := fmt.Sprintf("%.2f|%s|%s", amount, date, normalized)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func (s *Service) Exists(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM transactions WHERE source_hash = $1)",
		hash,
	).Scan(&exists)
	return exists, err
}
