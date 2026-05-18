package fetcher

import (
	"context"
	"fmt"
	"strings"

	"intelhorizons/internal/store"
)

// ArXiv uses Atom XML via its own API endpoint; gofeed handles Atom natively.
type ArXiv struct {
	rss *RSS
}

func NewArXiv() *ArXiv {
	return &ArXiv{rss: NewRSS()}
}

func (a *ArXiv) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	feed, err := a.rss.parser.ParseURLWithContext(src.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("parse arxiv %s: %w", src.URL, err)
	}

	articles := make([]store.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			continue
		}

		// Use the abstract URL (abs link) as canonical; prefer https
		link := item.Link
		for _, l := range item.Links {
			if strings.Contains(l, "abs/") {
				link = l
				break
			}
		}

		summary := stripWhitespace(item.Description)
		if len(summary) > 800 {
			summary = summary[:800]
		}

		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		pub := publishedAt(item)

		articles = append(articles, store.Article{
			URLHash:     store.HashURL(link),
			URL:         link,
			Title:       strings.TrimSpace(item.Title),
			Summary:     summary,
			Author:      author,
			SourceID:    src.ID,
			PublishedAt: pub,
			Category:    "Research",
		})
	}
	return articles, nil
}

func stripWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var parts []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

var _ Fetcher = (*ArXiv)(nil)
