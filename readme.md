# Megome (Go)

A modular backend API written in **Go**, designed to serve as the backend for a personal portfolio, resume builder, and other related projects.  
Built with scalability, modularity, and simplicity in mind.

<!-- --- -->

<!-- ## Features

- **Health Check**
  - `GET /health` – Verify the API is running.
- **Profile Management**
  - `GET /profile` – Fetch profile information.
  - `PUT /profile` – Update profile information.
- **Projects Management**
  - `GET /projects` – List all projects.
  - `POST /projects` – Add a new project.
  - `GET /projects/:id` – Get project by ID.
- **Resume Builder**
  - `GET /resume/sections` – List resume sections.
  - `POST /resume/sections` – Add a resume section.
  - `DELETE /resume/sections/:id` – Delete a section.
- **Authentication (JWT)**
  - Admin routes are protected with JSON Web Tokens.
- **Persistence**
  - SQLite / Postgres integration (configurable).
- **Logging & Error Handling**
  - Structured logs for requests and errors. -->

---

## Project Structure
    /cmd/api # Main entry point
    /internal/handlers # HTTP handlers
    /internal/services # Business logic
    /internal/data # Database layer
    /pkg # Reusable packages

---

<!-- ## Getting Started

### Prerequisites

- Go 1.25+ installed
- SQLite / Postgres (optional)
- `make` (optional, for scripts) -->

<!-- ### Run Locally

```bash
git clone https://github.com/Gadu25/megome.git
cd portfolio-backend
go mod tidy
go run ./cmd/api -->

# Database Migrations

This project uses **golang-migrate** with **MySQL** and **file-based SQL migrations** to manage database schema changes.

Migrations allow you to version, apply, and revert database changes in a safe and repeatable way.

---

## 📁 Migration Structure

All migration files live in:
cmd/migrate/migrations/


Each migration consists of **two files**:
<timestamp><name>.up.sql
<timestamp><name>.down.sql


### Example
20251208080354_add-projects-table.up.sql
20251208080354_add-projects-table.down.sql


- `.up.sql` → applies the schema change
- `.down.sql` → reverts the schema change
- Both files must exist
- The timestamp/version must match

---

## 🔧 Prerequisites

### 1. Go
Ensure Go is installed:

      go version

### 2. Install golang-migrate CLI (MySQL support)
     go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
      export PATH=$PATH:$HOME/go/bin

      migrate -version

If migrate is not found, ensure $HOME/go/bin is in your PATH:
export PATH=$PATH:$HOME/go/bin

Reload your shell if needed.

### 3. Creating a Migration

Use the Makefile command:
make migration <migration-name>

Example
make migration add-projects-table


### 4. Running Migrations

Apply all pending migrations
make migrate-up

This runs:
go run cmd/migrate/main.go up

Revert all migrations
go run cmd/migrate/main.go down 

⚠️ Warning: This will revert all migrations.
Use only in development environments.
