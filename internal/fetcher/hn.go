package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"intelhorizons/internal/store"
)

type HN struct {
	client *http.Client
}

func NewHN() *HN {
	return &HN{client: &http.Client{Timeout: 15 * time.Second}}
}

type hnResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID    string `json:"objectID"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	StoryURL    string `json:"story_url"`
	Author      string `json:"author"`
	CreatedAtI  int64  `json:"created_at_i"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
}

func (h *HN) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch hn: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hn api status %d", resp.StatusCode)
	}

	var result hnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	articles := make([]store.Article, 0, len(result.Hits))

	for _, hit := range result.Hits {
		pub := time.Unix(hit.CreatedAtI, 0).UTC()
		if pub.Before(cutoff) {
			continue
		}

		// Prefer the external URL; fall back to the HN thread itself
		link := hit.URL
		if link == "" {
			link = hit.StoryURL
		}
		if link == "" {
			link = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		}
		if hit.Title == "" {
			continue
		}

		summary := fmt.Sprintf("HN points: %d | comments: %d", hit.Points, hit.NumComments)

		articles = append(articles, store.Article{
			URLHash:     store.HashURL(link),
			URL:         link,
			Title:       hit.Title,
			Summary:     summary,
			Author:      hit.Author,
			SourceID:    src.ID,
			PublishedAt: pub,
		})
	}
	return articles, nil
}
