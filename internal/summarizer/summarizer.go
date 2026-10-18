package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	apiURL  = "https://api.anthropic.com/v1/messages"
	model   = "claude-haiku-4-5-20251001"
	version = "2023-06-01"
)

type Summarizer struct {
	apiKey string
	client *http.Client
}

func New(apiKey string) *Summarizer {
	return &Summarizer{
		apiKey: apiKey,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *Summarizer) Summarize(ctx context.Context, title, rawSummary string) (string, error) {
	prompt := fmt.Sprintf(
		"Write a clean, factual 2-sentence summary of this AI news article. "+
			"Be concise and specific — include the key finding, model name, or announcement. "+
			"Do not start with phrases like \"The article\" or \"This piece\". Output only the summary, nothing else.\n\n"+
			"Title: %s\nContent: %s",
		title, rawSummary,
	)

	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 150,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", version)
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error struct{ Message string } `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, e.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Content) == 0 || result.Content[0].Text == "" {
		return "", fmt.Errorf("empty response from API")
	}

	return result.Content[0].Text, nil
}
