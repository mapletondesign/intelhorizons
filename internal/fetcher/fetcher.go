package fetcher

import (
	"context"

	"intelhorizons/internal/store"
)

type Fetcher interface {
	Fetch(ctx context.Context, src store.Source) ([]store.Article, error)
}
