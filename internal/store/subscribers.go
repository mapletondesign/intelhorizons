package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Subscriber struct {
	ID             int64
	Email          string
	Token          string
	Confirmed      bool
	ConfirmedAt    *time.Time
	UnsubscribedAt *time.Time
	CreatedAt      time.Time
}

// CreateSubscriber upserts a subscriber and reports whether a confirmation
// email should be sent. It handles four cases:
//   - new email → insert, send confirm
//   - pending confirmation → return existing token, resend confirm
//   - already confirmed and active → no-op, no send
//   - previously unsubscribed → generate new token, reset, send confirm
func (s *Store) CreateSubscriber(ctx context.Context, email string) (token string, needsConfirm bool, err error) {
	var existing struct {
		token          string
		confirmed      bool
		unsubscribedAt *time.Time
	}
	scanErr := s.pool.QueryRow(ctx,
		`SELECT token, confirmed, unsubscribed_at FROM subscribers WHERE email = $1`, email,
	).Scan(&existing.token, &existing.confirmed, &existing.unsubscribedAt)

	if scanErr != nil && !errors.Is(scanErr, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("query subscriber: %w", scanErr)
	}

	if errors.Is(scanErr, pgx.ErrNoRows) {
		tok := generateToken()
		if _, err := s.pool.Exec(ctx,
			`INSERT INTO subscribers (email, token) VALUES ($1, $2)`, email, tok,
		); err != nil {
			return "", false, fmt.Errorf("insert subscriber: %w", err)
		}
		return tok, true, nil
	}

	if existing.unsubscribedAt != nil {
		tok := generateToken()
		if _, err := s.pool.Exec(ctx, `
			UPDATE subscribers
			SET token=$1, confirmed=false, confirmed_at=NULL, unsubscribed_at=NULL
			WHERE email=$2
		`, tok, email); err != nil {
			return "", false, fmt.Errorf("resubscribe: %w", err)
		}
		return tok, true, nil
	}

	if existing.confirmed {
		return existing.token, false, nil
	}

	return existing.token, true, nil
}

func (s *Store) ConfirmSubscriber(ctx context.Context, token string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE subscribers
		SET confirmed=true, confirmed_at=NOW()
		WHERE token=$1 AND confirmed=false AND unsubscribed_at IS NULL
	`, token)
	if err != nil {
		return fmt.Errorf("confirm subscriber: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("token not found or already confirmed")
	}
	return nil
}

func (s *Store) UnsubscribeByToken(ctx context.Context, token string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE subscribers
		SET unsubscribed_at=NOW()
		WHERE token=$1 AND unsubscribed_at IS NULL
	`, token)
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("token not found or already unsubscribed")
	}
	return nil
}

func (s *Store) ActiveSubscribers(ctx context.Context) ([]Subscriber, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, token, confirmed, confirmed_at, unsubscribed_at, created_at
		FROM subscribers
		WHERE confirmed=true AND unsubscribed_at IS NULL
		ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("query active subscribers: %w", err)
	}
	defer rows.Close()

	var subs []Subscriber
	for rows.Next() {
		var sub Subscriber
		if err := rows.Scan(
			&sub.ID, &sub.Email, &sub.Token, &sub.Confirmed,
			&sub.ConfirmedAt, &sub.UnsubscribedAt, &sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscriber: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	return hex.EncodeToString(b)
}
