# IntelHorizons — Agent Orientation

## Project State

Live at **intelhorizons.com** (Fly.io, region iad, machine d8d044dce75268).

### Completed stages
- **Stage 1** — Go backend: chi v5, pgx/v5, RSS/Reddit/HN/ArXiv fetchers, keyword scorer, Redis cache
- **Stage 2** — Store layer: Postgres migrations, UpsertArticle, Feed query with DISTINCT ON deduplication, category filter, semantic+ILIKE search
- **Stage 3** — Nuxt 3 SPA: Feed, Digest, Sources pages; category pills, source filter, search, pagination
- **Stage 4** — Email digest: Resend double opt-in, daily scheduler, unsubscribe tokens, subscribe form + banners in Vue
- **Infrastructure** — Fly.io deploy, Postgres + Upstash Redis, intelhorizons.com DNS + SSL, Datadog RUM

### Key env vars (set as Fly secrets)
- `DATABASE_URL`, `REDIS_URL`, `RESEND_API_KEY`, `RESEND_FROM`, `SITE_URL`, `DIGEST_HOUR`

---

## Planned Work

### 1. LLM-Powered Summaries
Replace raw RSS summaries with Claude-generated 2-sentence summaries.

**Approach:**
- Add `ANTHROPIC_API_KEY` as a Fly secret
- New package `internal/summarizer/summarizer.go` — calls Claude API (haiku for cost) with the raw title + summary, returns a clean 2-sentence version
- Add `summarized bool` column to `articles` table (migration 004)
- Worker calls summarizer after fetch, before scoring, only for articles where `summarized = false`
- Rate-limit to avoid burning tokens: process max N articles per worker cycle (e.g. 10), skip if summary already looks clean (len > 100 chars and no HTML)
- API key: not yet set — prompt user before starting

**Files to touch:**
- `migrations/004_summarizer.sql`
- `internal/summarizer/summarizer.go` (new)
- `internal/store/store.go` — add `summarized` field, `MarkSummarized()` method
- `cmd/ainews/main.go` — wire summarizer into worker loop

---

### 2. Semantic Search
Augment/replace keyword search with vector similarity via pgvector.

**Approach:**
- Enable `pgvector` extension in Postgres (migration 005)
- Add `embedding vector(1536)` column to `articles` table
- New package `internal/embedder/embedder.go` — calls an embeddings API (OpenAI `text-embedding-3-small` or Anthropic when available) to generate embeddings for `title + summary`
- Background job embeds new articles after summarization
- Feed query: when `q` param is present, embed the query string and use `<=>` cosine distance operator; fall back to existing text search if no embeddings for an article yet
- Add `ivfflat` index on the embedding column once the table has enough rows

**Files to touch:**
- `migrations/005_pgvector.sql`
- `internal/embedder/embedder.go` (new)
- `internal/store/store.go` — add `embedding` field, `SetEmbedding()`, update `Feed()` to use vector search when query present
- `internal/api/api.go` — embed query string before calling `store.Feed()`
- Requires: `OPENAI_API_KEY` Fly secret (or reuse Anthropic key when embeddings API is available)

---

### 3. Sponsored Feed Slot
Insert a clearly-labelled sponsor card at position 4 in the feed (zero-indexed slot 3).

**Approach:**
- New table `sponsors` (migration 006): `id`, `company`, `headline`, `body`, `cta_label`, `cta_url`, `logo_url`, `active bool`, `start_date`, `end_date`, `created_at`
- New API endpoint `GET /api/sponsor/active` returns the current active sponsor (or 204 if none)
- Feed API response unchanged — sponsor injection happens client-side to keep the cache clean
- Vue: `index.vue` fetches active sponsor on load; if present, splices a `SponsorCard` component into position 3 of the articles array
- New component `web/components/SponsorCard.vue` — visually distinct card with "Sponsored" label, company name, headline, body copy, CTA button
- Sponsors managed by direct DB insert for now (no admin UI needed yet)

**Files to touch:**
- `migrations/006_sponsors.sql`
- `internal/store/sponsors.go` (new)
- `internal/api/api.go` — add `GET /api/sponsor/active` route + handler
- `web/components/SponsorCard.vue` (new)
- `web/pages/index.vue` — fetch sponsor, splice into articles list

---

## Suggested Order

1. **Sponsored slot** — no external API keys needed, ships fast, enables revenue
2. **LLM summaries** — needs `ANTHROPIC_API_KEY`; improves content quality across the board
3. **Semantic search** — depends on summaries being clean first; needs embeddings API key

---

## Deploy

```bash
cd /Users/apollo/Projects/intelhorizons
fly deploy --ha=false
```

Migrations run automatically on boot via `RunMigrations`.
