package dedup

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var storeNumberRe = regexp.MustCompile(`[#]\d+`)
var trailingDigitsRe = regexp.MustCompile(`\s+\d+$`)
var multiSpaceRe = regexp.MustCompile(`\s{2,}`)

const merchantHashLen = 20

// NormalizeMerchant strips store numbers, trailing digits, and extra
// whitespace, then truncates to a fixed length so that minor LLM
// variations (e.g. "CHUTNEYS INDIAN EXPR" vs "CHUTNEYS INDIAN EXPRE")
// produce the same value.
func NormalizeMerchant(merchant string) string {
	s := strings.ToLower(strings.TrimSpace(merchant))
	s = storeNumberRe.ReplaceAllString(s, "")
	s = trailingDigitsRe.ReplaceAllString(s, "")
	s = multiSpaceRe.ReplaceAllString(strings.TrimSpace(s), " ")
	if len(s) > merchantHashLen {
		s = s[:merchantHashLen]
	}
	return s
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func GenerateHash(amount float64, date string, merchant string, txnType string, accountLastFour string) string {
	normalized := NormalizeMerchant(merchant)
	input := fmt.Sprintf("%.2f|%s|%s|%s|%s", amount, date, normalized, strings.ToLower(txnType), accountLastFour)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func RehashTransaction(amount float64, date string, merchant string, txnType string, accountLastFour string) string {
	return GenerateHash(amount, date, merchant, txnType, accountLastFour)
}

func (s *Service) Exists(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM transactions WHERE source_hash = $1)",
		hash,
	).Scan(&exists)
	return exists, err
}
