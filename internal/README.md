This directory follows the Go `internal` package pattern. It is organized into:
- `domain/` – business logic and data models grouped by domain
- `api/` – HTTP router and handler definitions
- `pkg/` – reusable utilities (auth, HTTP helpers, mailer, storage, validation)
- `middleware/` – HTTP middleware (auth, rate limiting, logging)
- `config/` – environment-based configuration loading
- `db/` – database connection and seeders
