BEGIN;

SET search_path TO urlshortener;

ALTER TABLE refresh_tokens ALTER COLUMN token_hash TYPE VARCHAR(255);

COMMIT;
