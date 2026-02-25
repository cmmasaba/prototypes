# URL Shortener

## Database Layer
sqlx package<br>
### Prepared Statements
Prepared statements offer security, avoid sql injections and offer opportunities to gain performance advantages. Patterns to avoid:
- relying solely on database/sql's implicit caching of the statements
- preparing statements over and over everytime the query runs

Better pattern is caching the prepared statements at the repository layer. This will require a database connection i.e `sql.DB` and a statement cache i.e `map[string]*sqlx.NamedStmt`.
