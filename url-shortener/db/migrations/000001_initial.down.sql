BEGIN;

DROP INDEX IF EXISTS urlshortener.idx_clicks_link_id_clicked_at;
DROP INDEX IF EXISTS urlshortener.idx_clicks_link_id_country;
DROP INDEX IF EXISTS urlshortener.idx_clicks_country;
ALTER TABLE IF EXISTS urlshortener.clicks DROP CONSTRAINT IF EXISTS clicks_link_id_fkey;
DROP TABLE IF EXISTS urlshortener.clicks;

DROP INDEX IF EXISTS urlshortener.id_links_expires_at;
DROP TABLE IF EXISTS urlshortener.links;

DROP INDEX IF EXISTS urlshortener.idx_users_oauth;
DROP TABLE IF EXISTS urlshortener.users;

DROP INDEX IF EXISTS urlshortener.idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS urlshortener.idx_refresh_tokens_expires_at;
ALTER TABLE IF EXISTS urlshortener.users DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;
DROP TABLE IF EXISTS urlshortener.refresh_tokens;

DROP SCHEMA IF EXISTS urlshortener;

COMMIT;
