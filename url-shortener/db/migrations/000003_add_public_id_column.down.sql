BEGIN;

SET search_path TO urlshortener;

ALTER TABLE users DROP COLUMN IF EXISTS public_id;

COMMIT;
