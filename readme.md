# Megome

A modular backend API written in **Go**, built to power a personal portfolio, resume builder, and related projects. It exposes a private admin API for content management and a public API (secured with personal access tokens) for consuming that content.

## Features

- **Authentication & Account Management**
  - Email/password registration and login
  - Google OAuth login
  - JWT access tokens + rotating refresh tokens
  - Session management (list and revoke active sessions)
  - Password change, password reset via email, email/username change, account deletion
- **Content Management**
  - Profile, experiences, skills, education, certifications, and projects
  - Project images (uploaded to Cloudflare R2) and technology tags
- **Public API**
  - `/public/v1` routes authenticated with personal access tokens (PATs)
  - Per-request API usage logging
- **AI Assist**
  - Google Gemini–powered content generation for bios, taglines, project descriptions, experiences, and education
  - Automatic quota cooldown tracking with graceful `429` handling
- **Extras**
  - Dashboard overview with API usage stats and PAT counts
  - Full JSON data export of all user content
  - Per-IP rate limiting, structured request logging, CORS
- **Persistence**
  - MySQL with versioned, file-based SQL migrations (`golang-migrate`)
  - Seeding for starter data

## Tech Stack

| Layer      | Technology |
|------------|------------|
| Language   | Go 1.25 |
| HTTP       | `gorilla/mux` |
| Database   | MySQL (`go-sql-driver/mysql`) |
| Migrations | `golang-migrate` |
| Auth       | `golang-jwt/jwt/v4`, Google OAuth (`golang.org/x/oauth2`) |
| Storage    | Cloudflare R2 (`minio-go`) |
| Email      | SMTP (`gopkg.in/gomail.v2`) with HTML templates |
| AI         | Google Gemini REST API |
| Validation | `go-playground/validator/v10` |
| Config     | `.env` via `joho/godotenv` |

## Project Structure

```
cmd/
  api/                  # Main API entry point
  migrate/              # Database migration runner
  seed/                 # Database seeder
internal/
  api/                  # HTTP router and route registration
    handler/            # Private (JWT) handlers
    handler/public/     # Public (PAT) handlers
  config/               # Environment-based configuration
  data/                 # Data seeders
  db/                   # Database connection and seeders
  domain/               # Models and repositories, grouped by domain
  middleware/           # Auth (JWT + PAT), rate limiting, request logging
  pkg/                  # Reusable packages: auth, ai, mailer, storage, httputil, validator
```

## Getting Started

### Prerequisites

- Go 1.25+
- MySQL
- `make` (optional, for the commands below)
- `golang-migrate` CLI (only for creating new migrations with `make migration`):
  ```bash
  go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  export PATH=$PATH:$HOME/go/bin
  ```

### Setup

```bash
git clone <your-repo-url> && cd megome
go mod tidy
```

Create a `.env` file based on the settings in `internal/config/config.go`:

```
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
PORT="8080"

DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="your_db_user"
DB_PASSWORD="your_db_password"
DB_NAME="megome"

JWT_SECRET="a-long-random-secret"
JWT_EXP="1800"
```

Optional integrations (leave unset to disable):

```
GOOGLE_OAUTH_CLIENT_ID="..."
GOOGLE_OAUTH_SECRET="..."

SMTP_HOST="smtp.gmail.com"
SMTP_PORT="587"
SMTP_USERNAME="..."
SMTP_PASSWORD="..."
SMTP_FROM="..."

GEMINI_API_KEY="..."
GEMINI_MODEL="gemini-2.0-flash"
GEMINI_QUOTA_COOLDOWN="1800"

R2_ACCOUNT_ID="..."
R2_ACCESS_KEY_ID="..."
R2_SECRET_ACCESS_KEY="..."
R2_BUCKET="megome"
R2_ENDPOINT="..."
R2_PUBLIC_URL="..."
```

### Database Migrations

This project uses **golang-migrate** with **MySQL** and file-based SQL migrations to manage schema changes. Each migration consists of a matching pair of files:

```
cmd/migrate/migrations/<timestamp>_<name>.up.sql
cmd/migrate/migrations/<timestamp>_<name>.down.sql
```

- `.up.sql` applies the schema change
- `.down.sql` reverts it
- Both files must exist and share the same timestamp

Create a new migration:

```bash
make migration add-example-table
```

Apply all pending migrations:

```bash
make migrate-up
```

Revert all migrations (development only):

```bash
make migrate-down
```

### Run

```bash
make run          # builds and runs the API server
# or
go run ./cmd/api
```

Seed the database with starter data:

```bash
make seed
```

## Makefile Commands

| Command                | Description                        |
|------------------------|------------------------------------|
| `make build`           | Build the binary to `bin/megome`   |
| `make run`             | Build and run the API server       |
| `make test`            | Run the full test suite            |
| `make migration`       | Create a new migration pair        |
| `make migrate-up`      | Apply all pending migrations       |
| `make migrate-down`    | Revert all migrations              |
| `make seed`            | Seed the database                  |

## API Overview

Routes are grouped under two base paths:

- `/api/v1` — private admin API, protected by JWT
- `/public/v1` — public content API, protected by personal access tokens (PATs)

### Auth (`/api/v1`)

- `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`
- `GET /auth/verify`
- `GET /auth/google`, `GET /auth/google/callback`
- `POST /auth/forgot-pass`, `POST /auth/change-forgot-pass`
- `POST /auth/change-password`, `POST /auth/change-email`, `POST /auth/change-username`
- `GET /auth/sessions`, `POST /auth/sessions/{id}/revoke`
- `DELETE /auth/account`

### Content Management (`/api/v1`)

- Profile, experience, skill, education, certification, project, and technology CRUD
- `GET|POST /project/{id}/images`, `PUT /project/{id}/cover` (uploads to R2) and `DELETE /project-images/{id}`
- Project/experience technology tagging via `/projectTech` and `/experienceTech` (including batch updates)

### Tools (`/api/v1`)

- `GET /pat`, `POST /pat`, `POST /pat/{id}/revoke`, `DELETE /pat/{id}`, `GET /pat/count`
- `POST /ai/assist`, `GET /ai/status`
- `GET /dashboard/overview`
- `GET /data/export`

### Public (`/public/v1`)

- `GET /profile`, `GET /skills`, `GET /education`, `GET /projects`, `GET /experiences`, `GET /certifications`

Requests to the public API are authenticated with a personal access token and logged for usage analytics.

## License

No license specified.
