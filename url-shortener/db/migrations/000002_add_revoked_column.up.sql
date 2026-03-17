BEGIN;

SET search_path TO urlshortener;

ALTER TABLE IF EXISTS refresh_tokens ADD revoked BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

COMMIT;
