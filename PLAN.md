# IntelHorizons — AI News Aggregator

Pulls articles from RSS feeds, ArXiv, Hacker News, and Reddit. Scores by AI relevance + recency. Serves a ranked feed at intelhorizons.com.

---

## Locked Decisions

| Decision | Choice |
|---|---|
| Backend | Go, chi v5, pgx/v5, go-redis |
| Deployment | Fly.io (shared-cpu-1x, 512MB, iad) |
| Scoring | Term-weight relevance × exponential recency decay (half-life ~24h) |
| Sources v1 | RSS, ArXiv, HN, Reddit (Twitter deferred — needs bearer token) |
| Frontend | Vue 3 (Nuxt.js) — not started |
| Auth | None for now — read-only public API is acceptable for Stage 3 |
| Caching | Redis — feed and digest results cached (TTL configurable) |

---

## Stages

| # | Stage | Status | Notes |
|---|---|---|---|
| 1 | Backend — ingestion, scorer, API, deploy | Complete | All fetchers live except Twitter |
| 2 | Backend cleanup | Not started | Fix known gaps before shipping frontend |
| 3 | Vue 3 frontend | Not started | Feed, digest, source filter, search |
| 4 | Email digest | Not started | Daily top-20 email via Resend or Postmark |
| 5 | Auth + personalization | Not started | Saved articles, custom source weights |
| 6 | Source management UI | Not started | Admin UI to add/remove/configure sources |
| 7 | Twitter integration | Blocked | Needs `TWITTER_BEARER_TOKEN` |

---

## Stage 2 — Backend Cleanup

These gaps must be resolved before the frontend ships. None require schema changes.

| Gap | Fix |
|---|---|
| CORS hardcoded to `https://intelhorizons.com` | Load from `PORTAL_ORIGIN` env var; fall back to that string for production |
| Migrations re-run every boot | Add `schema_migrations` table; skip files already recorded |
| Article scores frozen after first insert | Change `UpsertArticle` to `ON CONFLICT (url_hash) DO UPDATE SET score = EXCLUDED.score` |
| `/healthz` returns plain text | Return `{"status":"ok","db":true,"redis":true}` with 503 if either is down — match AdLite pattern |
| No security headers | Add `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `HSTS` middleware |
| No request ID middleware | Add `middleware.RequestID` to chi router |

---

## Stage 3 — Vue 3 Frontend (`web/`)

Stack: Nuxt.js (Vue 3, Composition API, `<script setup>`), Tailwind CSS.

**Pages:**
- `/` — ranked feed; infinite scroll or pagination; filter by source; full-text search box
- `/digest` — today's top 20; shareable permalink
- `/sources` — list of active sources with type and fetch frequency

**Behaviour:**
- Reads from existing API (`/api/feed`, `/api/digest`, `/api/sources`) — no new endpoints needed for Stage 3
- No login required — fully public

**`web/dist`** is served by the Go server at `/*`. Run `nuxt build` and commit the dist output, or serve separately behind a CDN.

---

## Stage 4 — Email Digest

- Daily email at configurable time (default 08:00 UTC) — top 20 articles from `/api/digest`
- Subscriber list stored in new `subscribers` table (email, confirmed, created_at)
- Double opt-in confirmation flow
- Unsubscribe link in every email (one-click, signed token)
- Provider: Resend or Postmark (decide when implementing)

---

## Stage 5 — Auth + Personalization

- Email / password accounts (bcrypt ≥ 12, RS256 JWT — 15 min access / 7 day refresh)
- Saved articles per user
- Custom source-weight overrides (e.g., upweight ArXiv, downweight Reddit)
- Personal feed endpoint that applies user weights on top of global scores

---

## API Reference (current)

| Method | Path | Description |
|---|---|---|
| GET | `/api/feed` | Ranked article feed. Params: `limit` (max 100), `offset`, `source` (ID), `q` (full-text search) |
| GET | `/api/digest` | Top 20 articles published in the last 24h |
| GET | `/api/sources` | All active sources |
| GET | `/healthz` | Health check (plain text today — fix in Stage 2) |

---

## Scorer Notes

Relevance is computed from term-weight matches against `title + summary`:

| Term group | Weight |
|---|---|
| LLM / language model / specific model names | 2.5 |
| AGI, agent, agentic, multi-agent | 2.0 |
| Benchmarks, SOTA, evaluation | 1.5 |
| Multimodal, reasoning, fine-tuning, RLHF | 1.5 |
| Safety, alignment, regulation | 1.0–1.2 |
| Research / paper baseline bump | +0.5 |

Recency multiplier: `exp(-age_hours / 24)` — full weight at 0h, ~37% at 24h, near-zero at 7d.

Final score clamped to [0, 10].

---

## Infrastructure

| Item | Detail |
|---|---|
| Database | PostgreSQL 16 — `sources` + `articles` tables, tsvector GIN index for full-text |
| Cache | Redis 7 — feed TTL 5 min, digest TTL 15 min |
| Deploy | Fly.io `fly.toml` — `MODE=both` runs worker + server in one process |
| Separate worker | Set `MODE=worker` and `MODE=server` on separate machines when traffic warrants it |

---

## Open Questions

1. **Monetization** — display ads, sponsored slots, Pro tier with personalization, or API access for B2B?
2. **Twitter** — obtain bearer token and activate the Twitter fetcher
3. **Content freshness** — consider a separate re-scorer job to update scores on articles older than 1h
4. **CDN for frontend** — serve `web/dist` from Cloudflare rather than the Go process for better perf
