package mailer

import (
	"context"
	"fmt"
	"log"
	"time"

	"intelhorizons/internal/store"
)

func RunDigestScheduler(ctx context.Context, st *store.Store, m *Mailer, hour int) {
	for {
		next := nextOccurrence(hour)
		log.Printf("digest scheduler: next send at %s", next.UTC().Format(time.RFC3339))
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
		}
		if err := sendDailyDigest(ctx, st, m); err != nil {
			log.Printf("digest scheduler: %v", err)
		}
	}
}

func sendDailyDigest(ctx context.Context, st *store.Store, m *Mailer) error {
	articles, err := st.DailyDigest(ctx, 20)
	if err != nil {
		return fmt.Errorf("daily digest: %w", err)
	}
	if len(articles) == 0 {
		log.Println("digest scheduler: no articles today, skipping send")
		return nil
	}

	subscribers, err := st.ActiveSubscribers(ctx)
	if err != nil {
		return fmt.Errorf("active subscribers: %w", err)
	}
	if len(subscribers) == 0 {
		log.Println("digest scheduler: no active subscribers, skipping send")
		return nil
	}

	log.Printf("digest scheduler: sending %d articles to %d subscribers", len(articles), len(subscribers))
	if err := m.SendDigest(ctx, subscribers, articles, time.Now().UTC()); err != nil {
		return fmt.Errorf("send digest: %w", err)
	}
	log.Println("digest scheduler: done")
	return nil
}

func nextOccurrence(hour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
