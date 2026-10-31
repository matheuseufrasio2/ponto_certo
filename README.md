# Ponto Certo API 💈

High-performance RESTful API for barbershop scheduling and operations management, built in Go following modern, clean, and idiomatic patterns.

---

## 🏛️ Architecture Overview

The project uses a pragmatic layered architecture (Clean/Layered Architecture) tailored for Go:

```text
ponto_certo/
├── cmd/
│   └── api/
│       └── main.go           # Application entrypoint, dependency injection, and Graceful Shutdown
├── internal/                 # Private code (compiler prevents external imports)
│   ├── config/               # Strongly-typed configuration (.env / environment variables)
│   ├── domain/               # Pure domain entities and business errors (zero dependencies)
│   ├── handler/              # HTTP layer (request routing, validation, JSON responses)
│   ├── service/              # Use cases and business logic (consumer-defined interfaces)
│   └── repository/           # Persistence layer (PostgreSQL)
├── migrations/               # Versioned SQL migration scripts
├── docs/                     # Generated OpenAPI/Swagger specifications
├── .air.toml                 # Hot-reload configuration for local development
├── .env.example              # Environment variables template
├── Makefile                  # Automation commands (dev, docs, build, test)
├── go.mod
└── go.sum
```

### Unidirectional Data Flow:
$$\text{HTTP Request} \longrightarrow \text{internal/handler} \longrightarrow \text{internal/service} \longrightarrow \text{internal/repository} \longrightarrow \text{PostgreSQL}$$

> **Core Go Principle:** "Accept interfaces, return structs". Persistence interfaces are defined inside `internal/service` where they are consumed, eliminating premature coupling and circular imports (`import cycle not allowed`).

---

## 🚀 Getting Started

### Prerequisites
* **Go** 1.22+ installed
* (Optional) **Air** for hot-reload during development:
  ```bash
  make air-install
  ```

### 1. Configure Environment Variables
Copy the example environment configuration:
```bash
cp .env.example .env
```

### 2. Start the Server
```bash
# Starts development server (uses Air if installed, falls back to go run)
make dev
```

The server starts at `http://localhost:8080`.

### 3. Interactive Swagger Documentation
With the server running, access the interactive OpenAPI documentation:
👉 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 🛠️ Makefile Commands

| Command | Description |
| :--- | :--- |
| `make dev` | Generates Swagger docs and starts the server with hot-reload |
| `make docs` | Updates Swagger documentation using `swag` |
| `make build` | Compiles optimized production binary into `bin/api` |
| `make test` | Runs the full test suite |
| `make db-up` | Starts PostgreSQL 16 container and waits for healthcheck |
| `make db-down` | Stops PostgreSQL container preserving volume data |
| `make db-logs` | Streams PostgreSQL container logs in real time |
| `make db-shell` | Opens interactive `psql` shell in database container |
| `make db-reset` | Recreates database from scratch (drops volume) |
| `make migrate-up` | Applies versioned SQL database migrations |
| `make migrate-down` | Reverts the most recent database migration |
| `make migrate-status` | Checks current database migration version |
| `make air-install` | Installs the `air` hot-reload tool |
| `make clean` | Removes build binaries and temporary artifacts |

---

## ✨ Features & Patterns

1. **Strongly-Typed Config (`internal/config`)**:
   - Environment variable loader with native `.env` support and zero third-party dependencies.
2. **Structured Logging (`log/slog`)**:
   - Human-friendly colorized text for development.
   - High-throughput JSON output for production environments.
3. **Graceful Shutdown**:
   - Captures `SIGINT` (Ctrl+C) and `SIGTERM` (Docker/K8s) allowing in-flight requests up to 10s to complete.
4. **Safe HTTP Helpers (`internal/handler/response.go`)**:
   - `WriteJSON`: sets headers and marshals payload.
   - `ReadJSON`: DoS protection via 1MB body limit (`http.MaxBytesReader`), unknown field rejection, and explicit error diagnosis.
   - `WriteError`: uniform error JSON responses (`{"error": "..."}`).
5. **Hot-Reload with Air**:
   - Configured through `.air.toml` and wired into `make dev`.
6. **Automated Releases with Release Please**:
   - Automated `CHANGELOG.md` generation and SemVer git tags driven by Conventional Commits.
