package fetcher

import (
	"context"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"

	"intelhorizons/internal/store"
)

// redditTagRe matches Reddit post-type tags like [D], [P], [R], [N], [Q], [Project]
var redditTagRe = regexp.MustCompile(`(?i)\s*\[[^\]]{1,20}\]\s*$`)

type Reddit struct {
	parser *gofeed.Parser
}

func NewReddit() *Reddit {
	p := gofeed.NewParser()
	p.UserAgent = userAgent
	return &Reddit{parser: p}
}

func (r *Reddit) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	feed, err := r.parser.ParseURLWithContext(src.URL, ctx)
	if err != nil {
		return nil, err
	}

	articles := make([]store.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			continue
		}

		pub := publishedAt(item)

		summary := stripHTML(item.Content)
		if summary == "" {
			summary = stripHTML(item.Description)
		}
		if r := []rune(summary); len(r) > 800 {
			summary = string(r[:800])
		}

		title := strings.TrimSpace(redditTagRe.ReplaceAllString(item.Title, ""))

		author := ""
		if item.Author != nil {
			author = strings.TrimPrefix(item.Author.Name, "/u/")
		}

		articles = append(articles, store.Article{
			URLHash:     store.HashURL(item.Link),
			URL:         item.Link,
			Title:       title,
			Summary:     summary,
			Author:      author,
			SourceID:    src.ID,
			PublishedAt: pub,
		})
	}
	return articles, nil
}
