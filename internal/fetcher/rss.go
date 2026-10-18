package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"

	"intelhorizons/internal/store"
)

const userAgent = "IntelHorizons/1.0 (intelhorizons.com; aggregator)"

type RSS struct {
	client *http.Client
	parser *gofeed.Parser
}

func NewRSS() *RSS {
	client := &http.Client{Timeout: 15 * time.Second}
	parser := gofeed.NewParser()
	parser.Client = client
	parser.UserAgent = userAgent
	return &RSS{client: client, parser: parser}
}

func (r *RSS) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	feed, err := r.parser.ParseURLWithContext(src.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", src.URL, err)
	}

	articles := make([]store.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			continue
		}

		pub := publishedAt(item)
		if pub.Before(time.Now().Add(-7 * 24 * time.Hour)) {
			continue
		}

		summary := stripHTML(item.Description)
		if item.Content != "" {
			clean := stripHTML(item.Content)
			if len(clean) > len(summary) {
				summary = clean
			}
		}
		if r := []rune(summary); len(r) > 1000 {
			summary = string(r[:1000])
		}

		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		articles = append(articles, store.Article{
			URLHash:     store.HashURL(item.Link),
			URL:         item.Link,
			Title:       item.Title,
			Summary:     summary,
			Author:      author,
			SourceID:    src.ID,
			PublishedAt: pub,
		})
	}
	return articles, nil
}

func publishedAt(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	return time.Now()
}
