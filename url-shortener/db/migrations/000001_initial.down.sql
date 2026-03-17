BEGIN;

SET search_path TO urlshortener;

DROP INDEX IF EXISTS idx_clicks_link_id_clicked_at;
DROP INDEX IF EXISTS idx_clicks_link_id_country;
DROP INDEX IF EXISTS idx_clicks_country;
ALTER TABLE IF EXISTS clicks DROP CONSTRAINT IF EXISTS clicks_link_id_fkey;
DROP TABLE IF EXISTS clicks;

DROP INDEX IF EXISTS id_links_expires_at;
DROP TABLE IF EXISTS links;

DROP INDEX IF EXISTS idx_users_oauth;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;
DROP TABLE IF EXISTS refresh_tokens;

DROP SCHEMA IF EXISTS urlshortener;

COMMIT;
