-- name: SaveShortLink :one
INSERT INTO links (
	short_code, original_url, ownership_token, expires_at
) VALUES (
	$1, $2, $3, $4
)
RETURNING *;

-- name: GetShortLinkByCode :one
SELECT * FROM links
WHERE
	short_code = $1;

-- name: GetShortLinkByExpiresAt :many
SELECT * FROM links
WHERE
	expires_at IS NOT NULL;

-- name: SaveNewClick :one
INSERT INTO clicks (
	link_id, clicked_at, ip_hash, referrer, user_agent, device_type, browser, os, country, city
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetClicksByLinkIDAndClickedAt :many
SELECT * FROM clicks
WHERE
	link_id = $1 AND clicked_at = $2;

-- name: GetClicksByLinkIDAndCountry :many
SELECT * FROM clicks
WHERE
	link_id= $1 AND country = $2;

-- name: SaveUser :one
INSERT INTO users (
	email, password_hash, oauth_provider, oauth_provider_id
) VALUES (
	$1, $2, $3, $4
)
RETURNING *;

-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (
	user_id, token_hash, expires_at, revoked
) VALUES (
	$1, $2, $3, $4
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE
	email = $1;

-- name: GetUserByOauthID :one
SELECT * FROM users
WHERE
	oauth_provider = $1 AND oauth_provider_id = $2;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens
WHERE
	token_hash = $1;
