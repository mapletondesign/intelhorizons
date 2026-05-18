package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"intelhorizons/internal/store"
)

type Reddit struct {
	client *http.Client
}

func NewReddit() *Reddit {
	return &Reddit{client: &http.Client{Timeout: 15 * time.Second}}
}

type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	URL       string  `json:"url"`
	Permalink string  `json:"permalink"`
	Author    string  `json:"author"`
	Score     int     `json:"score"`
	Selftext  string  `json:"selftext"`
	Created   float64 `json:"created_utc"`
	IsSelf    bool    `json:"is_self"`
}

func (r *Reddit) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	// Reddit requires a non-browser User-Agent for its JSON API
	req.Header.Set("User-Agent", userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch reddit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit api status %d", resp.StatusCode)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	articles := make([]store.Article, 0, len(listing.Data.Children))

	for _, child := range listing.Data.Children {
		post := child.Data
		pub := time.Unix(int64(post.Created), 0).UTC()
		if pub.Before(cutoff) {
			continue
		}
		if post.Title == "" {
			continue
		}

		// For link posts use the external URL; for self-posts use the reddit thread
		link := post.URL
		if post.IsSelf {
			link = "https://www.reddit.com" + post.Permalink
		}

		summary := post.Selftext
		if len(summary) > 800 {
			summary = summary[:800]
		}
		if summary == "" {
			summary = fmt.Sprintf("Reddit score: %d", post.Score)
		}

		articles = append(articles, store.Article{
			URLHash:     store.HashURL(link),
			URL:         link,
			Title:       post.Title,
			Summary:     summary,
			Author:      post.Author,
			SourceID:    src.ID,
			PublishedAt: pub,
		})
	}
	return articles, nil
}
