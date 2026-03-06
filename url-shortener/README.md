# URL Shortener

## Database Layer

- pgxpool - performant postgresql driver with connection pooling, automatically uses cached prepared statements
- sqlc - type-safe Go code generated from SQL
- migrate - handling database migrations

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
