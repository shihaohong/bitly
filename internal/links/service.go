package links

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLen = 7

type Link struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID, originalURL string) (*Link, error) {
	code, err := generateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}

	var link Link
	err = s.db.QueryRow(ctx,
		`INSERT INTO links (short_code, original_url, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, short_code, original_url, click_count, created_at`,
		code, originalURL, userID,
	).Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.ClickCount, &link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert link: %w", err)
	}
	return &link, nil
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	var url string
	err := s.db.QueryRow(ctx,
		`UPDATE links SET click_count = click_count + 1
		 WHERE short_code = $1
		 RETURNING original_url`,
		code,
	).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("not found")
	}
	return url, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Link, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, short_code, original_url, click_count, created_at
		 FROM links WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.ShortCode, &l.OriginalURL, &l.ClickCount, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *Service) Delete(ctx context.Context, userID, code string) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM links WHERE short_code = $1 AND user_id = $2`,
		code, userID,
	)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func generateCode() (string, error) {
	b := make([]byte, codeLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
