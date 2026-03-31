-- name: SaveShortLink :one
INSERT INTO links (
	user_id, short_code, original_url, ownership_token, expires_at
) VALUES (
	$1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetShortLinkByCode :one
SELECT id, user_id, short_code, original_url, ownership_token FROM links
WHERE
	short_code = $1;

-- name: GetExpiredShortLinkByUserID :many
SELECT id, user_id, short_code, original_url, ownership_token FROM links
WHERE
	user_id=$1 AND expires_at IS NOT NULL AND expires_at < NOW()
ORDER BY created_at
LIMIT 50;

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
	link_id = $1 AND clicked_at = $2
ORDER BY clicked_at
LIMIT 50;

-- name: GetClicksByLinkIDAndCountry :many
SELECT * FROM clicks
WHERE
	link_id= $1 AND country = $2
ORDER BY clicked_at
LIMIT 50;

-- name: SaveUser :one
INSERT INTO users (
	email, password, oauth_provider, oauth_provider_id
) VALUES (
	$1, $2, $3, $4
)
RETURNING *;

-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (
	user_id, token, expires_at, revoked, created_at
) VALUES (
	$1, $2, $3, $4, $5
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE
	email = $1;

-- name: GetUserByOauthID :one
SELECT * FROM users
WHERE
	oauth_provider = $1 AND oauth_provider_id = $2;

-- name: GetUserByID :one
SELECT * FROM users
WHERE
	id = $1;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens
WHERE
	token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
	revoked = TRUE
WHERE
	token = $1;

-- name: GetUserByPublicID :one
SELECT * FROM users
WHERE
	public_id = $1;

-- name: CreateOTP :exec
INSERT INTO otp (
	code, expires_at, purpose, revoked, user_public_id
) VALUES (
	$1, $2, $3, $4, $5
);

-- name: GetOTPByCodeAndUserID :one
SELECT code, user_public_id, revoked, expires_at FROM otp
WHERE
	user_public_id=$1 AND code=$2 AND purpose=$3
LIMIT 1;

-- name: RevokeAllOTPsForUser :exec
UPDATE otp SET revoked = TRUE
WHERE user_public_id=$1 AND purpose=$2 AND revoked=FALSE;
