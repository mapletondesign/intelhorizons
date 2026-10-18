package fetcher

import (
	"context"
	"strings"

	"golang.org/x/net/html"

	"intelhorizons/internal/store"
)

type Fetcher interface {
	Fetch(ctx context.Context, src store.Source) ([]store.Article, error)
}

// stripHTML extracts plain text from an HTML string, collapsing whitespace.
func stripHTML(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.Join(strings.Fields(b.String()), " ")
}
