# URL Shortener

## Database Layer

- [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5@v5.8.0/pgxpool) - performant postgresql driver with connection pooling, automatically uses cached prepared statements
- [sqlc](https://docs.sqlc.dev/en/latest/index.html) - type-safe Go code generated from SQL
- [migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4@v4.19.1) - handling database migrations

#### BIGSERIAL vs UUID for Primary Keys

Performance benefits:

- Bigserial takes up only 8 bytes compared to 16 for UUID, this makes the index size smaller, allowing more rows to fit per page and faster joins.
- Bigserial produces sequential values which are B-tree friendly, new values append to the end of the index, avoiding fragmentation and page splits.
- UUIDs (v4) are random, causing scattered inserts and more time I/O.

Usability:

- Human-readable i.e using in queries, referencing in logs or when debugging.
- Natural ordering: using ORDER BY id gives insertion order for free.

Scenarios where UUID is better choice:

- Distributed systems where multiple nodes generate IDs independently.
- IDs are exposed publicly (sequential IDs leak count/ordering info)
- Merging data from multiple databases.

## Authentication

### Email/Password

Considerations:

- return the same error message (invalid credentials) for all authentication failures to combat `user enumeration` attacks.
- for pseudo-random RNGs use libraries that are cryptographically secure that do not produce predictable sequences. Libraries that are not very secure produce predictable sequences i.e given the same seed they always produce the same output and an attacker who can determine/influence the seed can "predict" all random values.
- to combat CSRF, validate the `Origin` header on all non-GET requests to ensure the request origin in legitimate.
- Reduce open redirect vulnerabilities by doing proper validation of redirect destinations.
- implement rate limits on critical endpoints either on IP-level or account-level.

Password Authentication:

- password validation should happen both on client-side and server-side
- set sensible minimum password requirements i.e minimum and maximum charactler length, what characters are acceptable, enable paste, avoid mandatory password resets, never silently modify user input.
- to verify password security run 2 checks: entropy-based strength estimation and [breach database checking](https://api.pwnedpasswords.com/range/%7Bfirst5chars%7D)
- MFA to combat brute force attacks
- Store passwords in hashed format using algorithms like bcrypt.

Email Verification:

- keep input validation rules simple
- Send a validation challenge/code to the provided email to confirm it exists, can be contacted and prove ownership

Delegate authentication to OAuth providers
Use MFA to boost security i.e password, TOTP
For JWT use ED25519 algorithm instead of HS256

References

- Generating JWT secrets: [online](https://jwtsecretkeygenerator.com/)
- Do you need OAuth and OIDC? [article](https://hackernoon.com/you-probably-dont-need-oauth2openid-connect-heres-why)
- Considerations for building Auth systems [article 1](https://medium.com/@loggd/building-secure-authentication-a-complete-guide-to-jwts-passwords-mfa-and-oauth-fdad8d243b91) [article 2](https://docs.cloud.google.com/solutions/modern-password-security-for-system-designers.pdf)

## Telemetry
