BEGIN;

SET search_path TO urlshortener;

DROP INDEX IF EXISTS idx_otp_user_code_purpose;
ALTER TABLE IF EXISTS otp DROP CONSTRAINT IF EXISTS unique_userid_code_purpose;
ALTER TABLE IF EXISTS otp DROP CONSTRAINT IF EXISTS otp_purpose_check;
DROP TABLE IF EXISTS otp;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP TABLE IF EXISTS refresh_tokens;

DROP INDEX IF EXISTS idx_clicks_link_id_clicked_at;
DROP INDEX IF EXISTS idx_clicks_link_id_country;
DROP INDEX IF EXISTS idx_clicks_clicked_at;
DROP TABLE IF EXISTS clicks;

DROP INDEX IF EXISTS idx_links_ownership_token;
DROP INDEX IF EXISTS idx_links_clicked_at;
DROP INDEX IF EXISTS idx_links_user_id;
DROP TABLE IF EXISTS links;

DROP INDEX IF EXISTS idx_users_oauth;
DROP INDEX IF EXISTS idx_users_public_id;
DROP TABLE IF EXISTS users;

DROP SCHEMA IF EXISTS urlshortener;

COMMIT;
