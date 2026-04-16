BEGIN;

CREATE SCHEMA urlshortener;

SET search_path TO urlshortener;

-- users table
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	public_id UUID NOT NULL DEFAULT uuidv7(),
	email VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NULL,
	oauth_provider VARCHAR(20) NULL,
	oauth_provider_id VARCHAR(255) NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_provider_id) WHERE oauth_provider_id IS NOT NULL;
CREATE UNIQUE INDEX idx_users_public_id ON users(public_id);

-- links table
CREATE TABLE IF NOT EXISTS links (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	short_code VARCHAR(50) UNIQUE NOT NULL,
	original_url TEXT NOT NULL,
	ownership_token VARCHAR(64) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_links_ownership_token ON links(ownership_token);
CREATE INDEX idx_links_user_id ON links(user_id);
CREATE INDEX idx_links_expires_at ON links(expires_at);

-- clicks table
CREATE TABLE IF NOT EXISTS clicks (
	id BIGSERIAL PRIMARY KEY,
	link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
	clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	ip_hash VARCHAR(64) NULL,
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
CREATE INDEX idx_clicks_clicked_at ON clicks(clicked_at) WHERE clicked_at IS NOT NULL;

-- refresh_token table
CREATE TABLE IF NOT EXISTS refresh_tokens (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token VARCHAR(255) NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	revoked BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token);

-- otp table
CREATE TABLE IF NOT EXISTS otp (
	id BIGSERIAL PRIMARY KEY,
	user_public_id UUID NOT NULL,
	purpose VARCHAR(25) NOT NULL CHECK ( purpose IN ('LOGIN', 'EMAIL_VERIFICATION', 'PASSWORD_RESET')),
	code CHAR(64) NOT NULL,
	revoked BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NOT NULL,
	CONSTRAINT unique_userid_code_purpose UNIQUE (user_public_id, code, purpose)
);
CREATE INDEX idx_otp_user_code_purpose ON otp(user_public_id, code, purpose);

-- login_attempts table
CREATE TABLE login_attempts (
	user_id BIGINT PRIMARY KEY REFERENCES users(id),
	fail_count INT NOT NULL DEFAULT 0,
	tier INT NOT NULL DEFAULT 0,
	locked_until TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
