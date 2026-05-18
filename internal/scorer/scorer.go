package scorer

import (
	"math"
	"strings"
	"time"

	"intelhorizons/internal/store"
)

// Term weights for AI relevance scoring.
// Higher weight = stronger signal of AI-specific content.
var terms = []struct {
	words  []string
	weight float64
}{
	{[]string{"llm", "large language model", "language model"}, 2.5},
	{[]string{"gpt", "claude", "gemini", "llama", "mistral", "grok", "deepseek"}, 2.5},
	{[]string{"agi", "artificial general intelligence"}, 2.0},
	{[]string{"agent", "agentic", "multi-agent"}, 2.0},
	{[]string{"benchmark", "sota", "state of the art", "evaluation", "leaderboard"}, 1.5},
	{[]string{"multimodal", "vision", "reasoning", "chain of thought"}, 1.5},
	{[]string{"fine-tun", "fine tuning", "rlhf", "reinforcement learning"}, 1.5},
	{[]string{"open source", "open-source", "weights", "hugging face"}, 1.2},
	{[]string{"safety", "alignment", "hallucination", "bias"}, 1.2},
	{[]string{"regulation", "policy", "legislation", "eu ai act"}, 1.0},
	{[]string{"model", "neural", "transformer", "diffusion"}, 1.0},
	{[]string{"robot", "robotics", "embodied"}, 1.0},
	{[]string{"openai", "anthropic", "deepmind", "google", "meta ai", "mistral ai"}, 0.8},
	{[]string{"research", "paper", "arxiv", "study", "published"}, 0.5},
	{[]string{"release", "launch", "announce", "unveil", "introduce"}, 0.5},
}

// Score assigns a relevance score in [0, 10] to each article.
func Score(articles []store.Article) []store.Article {
	scored := make([]store.Article, len(articles))
	for i, a := range articles {
		scored[i] = a
		scored[i].Score = scoreOne(a)
	}
	return scored
}

func scoreOne(a store.Article) float64 {
	haystack := strings.ToLower(a.Title + " " + a.Summary)
	var relevance float64
	for _, entry := range terms {
		for _, word := range entry.words {
			if strings.Contains(haystack, word) {
				relevance += entry.weight
				break // only count the group once
			}
		}
	}

	// Recency decay: full score within 6h, halved at 24h, approaching 0 at 7d
	age := time.Since(a.PublishedAt)
	recency := math.Exp(-age.Hours() / 24.0)

	// ArXiv and research papers get a small baseline bump
	if a.Category == "Research" {
		relevance += 0.5
	}

	raw := relevance * (0.4 + 0.6*recency)

	// Clamp to [0, 10]
	if raw > 10 {
		return 10
	}
	if raw < 0 {
		return 0
	}
	return raw
}
