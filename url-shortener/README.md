# URL Shortener

## Database Layer

sqlx package<br>

### Prepared Statements

Prepared statements offer security, avoid sql injections and offer opportunities to gain performance advantages. Patterns to avoid:

- relying solely on database/sql's implicit caching of the statements
- preparing statements over and over everytime the query runs

Better pattern is caching the prepared statements at the repository layer. This will require a database connection i.e `sql.DB` and a statement cache i.e `map[string]*sqlx.NamedStmt`.

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
