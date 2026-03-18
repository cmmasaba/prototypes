BEGIN;

SET search_path TO urlshortener;

ALTER TABLE users ADD COLUMN public_id UUID NOT NULL DEFAULT uuidv7();

CREATE UNIQUE INDEX idx_users_public_id ON users(public_id);

COMMIT;
