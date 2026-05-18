package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"intelhorizons/internal/store"
)

// Twitter uses the v2 API with a Bearer token.
// Free tier: user timeline endpoint, staggered to stay under rate limits.
type Twitter struct {
	client      *http.Client
	bearerToken string
}

func NewTwitter(bearerToken string) *Twitter {
	return &Twitter{
		client:      &http.Client{Timeout: 15 * time.Second},
		bearerToken: bearerToken,
	}
}

type twitterTimelineResponse struct {
	Data []twitterTweet `json:"data"`
}

type twitterTweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	AuthorID  string `json:"author_id"`
}

type twitterUserResponse struct {
	Data []twitterUser `json:"data"`
}

type twitterUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (t *Twitter) Fetch(ctx context.Context, src store.Source) ([]store.Article, error) {
	handles, ok := src.Config["handles"]
	if !ok || handles == "" {
		return nil, fmt.Errorf("twitter source %d missing handles config", src.ID)
	}

	usernames := strings.Split(handles, ",")
	userIDs, err := t.resolveUserIDs(ctx, usernames)
	if err != nil {
		return nil, fmt.Errorf("resolve user ids: %w", err)
	}

	var all []store.Article
	for _, uid := range userIDs {
		tweets, err := t.fetchTimeline(ctx, uid, src.ID)
		if err != nil {
			// Log and continue rather than aborting all accounts on one failure
			fmt.Printf("twitter timeline %s: %v\n", uid, err)
			continue
		}
		all = append(all, tweets...)
		// Stagger requests to respect free-tier rate limits
		time.Sleep(500 * time.Millisecond)
	}
	return all, nil
}

func (t *Twitter) resolveUserIDs(ctx context.Context, usernames []string) ([]string, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/by?usernames=%s&user.fields=id,username",
		strings.Join(usernames, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.bearerToken)
	req.Header.Set("User-Agent", userAgent)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users/by status %d", resp.StatusCode)
	}

	var result twitterUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Data))
	for _, u := range result.Data {
		ids = append(ids, u.ID)
	}
	return ids, nil
}

func (t *Twitter) fetchTimeline(ctx context.Context, userID string, srcID int) ([]store.Article, error) {
	url := fmt.Sprintf(
		"https://api.twitter.com/2/users/%s/tweets?max_results=10&tweet.fields=created_at&exclude=retweets,replies",
		userID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.bearerToken)
	req.Header.Set("User-Agent", userAgent)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("timeline status %d", resp.StatusCode)
	}

	var result twitterTimelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	articles := make([]store.Article, 0, len(result.Data))

	for _, tweet := range result.Data {
		pub, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			pub = time.Now()
		}
		if pub.Before(cutoff) {
			continue
		}

		link := fmt.Sprintf("https://x.com/i/web/status/%s", tweet.ID)
		articles = append(articles, store.Article{
			URLHash:     store.HashURL(link),
			URL:         link,
			Title:       truncate(tweet.Text, 200),
			Summary:     tweet.Text,
			SourceID:    srcID,
			PublishedAt: pub,
		})
	}
	return articles, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
