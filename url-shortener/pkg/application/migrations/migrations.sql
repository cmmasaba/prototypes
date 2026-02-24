CREATE SCHEMA 'urlshortener';

SET search_path TO 'urlshortener';

CREATE TABLE links (
	id BIGSERIAL PRIMARY KEY,
	short_code VARCHAR(50) UNIQUE NOT NULL,
	original_url TEXT NOT NULL,
	ownership_token VARCHAR(64) NOT NULL
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	expires_at TIMESTAMPTZ NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX idx_links_short_code ON links USING HASH(short_code);
CREATE INDEX id_links_expires_at ON links(short_code) WHERE expires_at IS NOT NULL;

CREATE TABLE clicks (
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
	city VARCHAR(100) NULL
)

CREATE INDEX idx_clicks_link_id_clicked_at ON clicks(link_id, clicked_at);
CREATE INDEX idx_clicks_country ON clicks(country);
