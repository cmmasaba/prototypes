BEGIN;

CREATE SCHEMA urlshortener;

SET search_path TO urlshortener;

CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NULL,
	oauth_provider VARCHAR(20) NULL,
	oauth_provider_id VARCHAR(255) NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_oauth ON users(oauth_provider, oauth_provider_id) WHERE oauth_provider_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS links (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id),
	short_code VARCHAR(50) UNIQUE NOT NULL,
	original_url TEXT NOT NULL,
	ownership_token VARCHAR(64) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_links_expires_at ON links(expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS clicks (
	id BIGSERIAL PRIMARY KEY,
	link_id BIGINT NOT NULL REFERENCES links(id),
	clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	ip_hash VARCHAR(64) NOT NULL,
	referrer TEXT NULL,
	user_agent TEXT NULL,
	device_type VARCHAR(20) NULL,
	browser VARCHAR(50) NULL,
	os VARCHAR(50) NULL,
	country VARCHAR(2) NULL,
	city VARCHAR(100) NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clicks_link_id_clicked_at ON clicks(link_id, clicked_at);
CREATE INDEX idx_clicks_link_id_country ON clicks(link_id, country);
CREATE INDEX idx_clicks_country ON clicks USING HASH(country);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash VARCHAR(64) UNIQUE NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

COMMIT;
