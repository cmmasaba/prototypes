BEGIN;

SET search_path TO urlshortener;

ALTER TABLE IF EXISTS refresh_tokens DROP COLUMN revoked;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;

COMMIT;
