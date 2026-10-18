CREATE TABLE IF NOT EXISTS subscribers (
    id              BIGSERIAL   PRIMARY KEY,
    email           TEXT        NOT NULL UNIQUE,
    token           CHAR(64)    NOT NULL UNIQUE,
    confirmed       BOOLEAN     NOT NULL DEFAULT false,
    confirmed_at    TIMESTAMPTZ,
    unsubscribed_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS subscribers_token_idx ON subscribers (token);
