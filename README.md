# IntelHorizons

An AI news aggregator. It pulls articles from RSS feeds, ArXiv, Hacker News and Reddit, scores them by AI relevance and recency, and serves a ranked feed and a daily email digest.

Live at [intelhorizons.com](https://intelhorizons.com).

## How it works

- A **worker** fetches from each configured source on a schedule, deduplicates articles and stores them in Postgres.
- A **scorer** ranks each article by term-weighted relevance multiplied by an exponential recency decay with a half-life of about 24 hours.
- An **API server** serves the ranked feed, the daily top-20 digest and the source list, with results cached in Redis. It also serves the built frontend.
- A **mailer** sends the daily digest through Resend, with double opt-in and signed unsubscribe links.

Worker and server ship in one binary. `MODE` selects which run.

## Repository layout

| Path | What it is |
|---|---|
| `cmd/ainews/` | Entry point for the worker and API server. |
| `internal/fetcher/` | Source fetchers: RSS, ArXiv, Hacker News, Reddit, Twitter. |
| `internal/scorer/` | Relevance and recency scoring. |
| `internal/store/` | Postgres access and migrations runner. |
| `internal/cache/` | Redis cache. |
| `internal/api/` | HTTP handlers and SPA file serving. |
| `internal/mailer/` | Email digest and subscription emails. |
| `internal/summarizer/` | Optional article summaries via the Claude API. |
| `migrations/` | SQL migrations, applied automatically on boot. |
| `web/` | Nuxt 3 frontend (Vue 3, Tailwind CSS): feed, digest and sources pages. |
| `scripts/` | TypeScript scripts that render social cards and profile images. |
| `PLAN.md` | Stages, decisions and open work. |

## Requirements

- Go 1.26+
- Node.js 22+
- Postgres and Redis

## Configuration

Configuration is read from environment variables. `.env.example` lists them; the binary does not load a `.env` file itself, so export the values into your shell before running.

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | none (required) | Postgres connection string. |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string. |
| `MODE` | `both` | `worker`, `server` or `both`. |
| `PORT` | `8080` | HTTP port. |
| `CORS_ORIGIN` | `https://intelhorizons.com` | Origin allowed to call the API. |
| `SITE_URL` | `https://intelhorizons.com` | Base URL used in email links. |
| `RESEND_API_KEY` | | Enables the email digest. |
| `RESEND_FROM` | `IntelHorizons <digest@intelhorizons.com>` | Digest sender address. |
| `DIGEST_HOUR` | `8` | Hour (UTC) the daily digest is sent. |
| `ANTHROPIC_API_KEY` | | Enables article summaries. |
| `TWITTER_BEARER_TOKEN` | | Enables the Twitter fetcher. |

## Running locally

Build the frontend first, since the Go server serves it from `web/dist`:

```bash
cd web
npm install
npm run build
cd ..
```

Then start the server and worker:

```bash
cp .env.example .env   # then fill in the values
set -a; source .env; set +a
go run ./cmd/ainews
```

The site is at `http://localhost:8080`. For frontend work with hot reload, run `npm run dev` in `web/` instead.

## API

| Route | Description |
|---|---|
| `GET /api/feed` | Ranked feed. Accepts `q` (search) and `category`. |
| `GET /api/digest` | Today's top articles. |
| `GET /api/sources` | Active sources. |
| `POST /api/subscribe` | Start a digest subscription (sends a confirmation email). |
| `GET /api/subscribe/confirm` | Confirm a subscription. |
| `GET /api/unsubscribe` | Unsubscribe via signed token. |
| `GET /healthz` | Health check. |

## Deployment

The app runs on Fly.io as a single machine built from the `Dockerfile`, which compiles the Go binary and the frontend into one image.

```bash
fly deploy --ha=false
```

Secrets are set with `fly secrets set`. Migrations run on boot.

## Status

Ingestion, scoring, the frontend and the email digest are complete. Accounts and personalization, a source management UI and Twitter integration are not yet built. See `PLAN.md`.
