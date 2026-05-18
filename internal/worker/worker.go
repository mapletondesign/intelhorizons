package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"intelhorizons/internal/fetcher"
	"intelhorizons/internal/scorer"
	"intelhorizons/internal/store"
)

type Worker struct {
	store    *store.Store
	fetchers map[string]fetcher.Fetcher
}

func New(st *store.Store, twitterToken string) *Worker {
	fetchers := map[string]fetcher.Fetcher{
		"rss":     fetcher.NewRSS(),
		"arxiv":   fetcher.NewArXiv(),
		"hn":      fetcher.NewHN(),
		"reddit":  fetcher.NewReddit(),
	}
	if twitterToken != "" {
		fetchers["twitter"] = fetcher.NewTwitter(twitterToken)
	}
	return &Worker{store: st, fetchers: fetchers}
}

// RunOnce fetches all sources that are due for a refresh.
func (w *Worker) RunOnce(ctx context.Context) {
	sources, err := w.store.ActiveSources(ctx)
	if err != nil {
		log.Printf("worker: load sources: %v", err)
		return
	}

	for _, src := range sources {
		if !isDue(src) {
			continue
		}
		w.fetchSource(ctx, src)
		// Brief pause between sources to be a polite HTTP client
		time.Sleep(500 * time.Millisecond)
	}
}

// RunForever polls on a fixed tick and calls RunOnce each time.
func (w *Worker) RunForever(ctx context.Context, interval time.Duration) {
	log.Printf("worker: starting, poll interval %s", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	w.RunOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("worker: shutting down")
			return
		case <-ticker.C:
			w.RunOnce(ctx)
		}
	}
}

func (w *Worker) fetchSource(ctx context.Context, src store.Source) {
	f, ok := w.fetchers[src.Type]
	if !ok {
		log.Printf("worker: no fetcher for type %q (source %d)", src.Type, src.ID)
		return
	}

	log.Printf("worker: fetching %s (%s)", src.Name, src.Type)
	articles, err := f.Fetch(ctx, src)
	if err != nil {
		log.Printf("worker: fetch %s: %v", src.Name, err)
		return
	}

	articles = scorer.Score(articles)

	saved := 0
	for _, a := range articles {
		if err := w.store.UpsertArticle(ctx, a); err != nil {
			log.Printf("worker: upsert article %s: %v", a.URL, err)
			continue
		}
		saved++
	}
	log.Printf("worker: %s — %d fetched, %d saved", src.Name, len(articles), saved)

	if err := w.store.UpdateSourceFetchedAt(ctx, src.ID); err != nil {
		log.Printf("worker: update source %d: %v", src.ID, err)
	}
}

func isDue(src store.Source) bool {
	if src.LastFetchedAt == nil {
		return true
	}
	return time.Since(*src.LastFetchedAt) >= time.Duration(src.FetchIntervalMins)*time.Minute
}

// Validate that all expected fetcher types are registered.
func (w *Worker) Validate() error {
	required := []string{"rss", "arxiv", "hn", "reddit"}
	for _, t := range required {
		if _, ok := w.fetchers[t]; !ok {
			return fmt.Errorf("missing fetcher for type %q", t)
		}
	}
	return nil
}
