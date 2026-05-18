CREATE TABLE IF NOT EXISTS sources (
    id                  SERIAL PRIMARY KEY,
    name                TEXT NOT NULL UNIQUE,
    type                TEXT NOT NULL CHECK (type IN ('rss', 'twitter', 'hn', 'reddit', 'arxiv')),
    url                 TEXT NOT NULL,
    config              JSONB NOT NULL DEFAULT '{}',
    active              BOOLEAN NOT NULL DEFAULT true,
    fetch_interval_mins INT NOT NULL DEFAULT 30,
    last_fetched_at     TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS articles (
    id           BIGSERIAL PRIMARY KEY,
    url_hash     CHAR(64) NOT NULL UNIQUE,
    url          TEXT NOT NULL,
    title        TEXT NOT NULL,
    summary      TEXT NOT NULL DEFAULT '',
    author       TEXT NOT NULL DEFAULT '',
    source_id    INT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    published_at TIMESTAMPTZ NOT NULL,
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    score        DOUBLE PRECISION NOT NULL DEFAULT 0,
    category     TEXT NOT NULL DEFAULT '',
    search_vec   TSVECTOR
);

CREATE INDEX IF NOT EXISTS articles_published_at_idx ON articles (published_at DESC);
CREATE INDEX IF NOT EXISTS articles_score_idx        ON articles (score DESC);
CREATE INDEX IF NOT EXISTS articles_source_id_idx    ON articles (source_id);
CREATE INDEX IF NOT EXISTS articles_search_vec_idx   ON articles USING GIN (search_vec);

CREATE OR REPLACE FUNCTION articles_search_vec_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vec := to_tsvector('english', NEW.title || ' ' || NEW.summary);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER articles_search_vec_trigger
    BEFORE INSERT OR UPDATE ON articles
    FOR EACH ROW EXECUTE FUNCTION articles_search_vec_update();
