---
name: atomic-commits
description: >-
  Enforces atomic, semantically focused commits and forbids bulk/monolithic commits.
  Provides the step-by-step workflow for partitioning multi-file Go tasks into single-responsibility commits with co-located tests.
---

# Atomic Commits & Semantic Granularity in Go

## 🎯 Purpose

This skill ensures AI agents produce **strictly atomic, semantically meaningful commits** instead of bundling multiple architectural layers into a single mega-commit. In `ponto_certo`, commit history directly drives automated changelogs and semver releases via **`release-please`**.

---

## 🛡️ Core Rules & Invariants

### 1. The Single-Responsibility Rule (No Conjunctions)
- Each commit **MUST** address exactly one concern, component, or logical unit.
- **NEVER** use conjunctions (`and`, `e`, `&`, `+`) in the commit title/header.
  - ❌ `feat(appointment): add booking service, handler and swagger docs`
  - ✅ `feat(domain): define appointment status transitions and domain errors`
  - ✅ `feat(service): implement booking conflict validation`
  - ✅ `feat(handler): add POST /api/v1/appointments endpoint`
  - ✅ `docs(swagger): document appointment booking request and response`
- If you find yourself writing "and" / "e" in a commit message, **stop and split the change into separate commits**.

### 2. Co-Located Tests (Green Commits)
- Always include the relevant unit or HTTP tests (`*_test.go`) in the **same commit** as the Go implementation.
- Every commit must represent a working, self-contained state that passes linting and tests (`go test -v -race ./...`).
- ❌ Do not create a dangling "add tests" commit at the end of the PR if intermediate commits break tests.
- ❌ Do not commit code that fails existing tests.

### 3. Progressive Commit Staging (Never `git add .`)
- In our Go Clean/Layered architecture, partition multi-file tasks into logical groups and commit incrementally:
  1. **Domain (`internal/domain`)**: Data structs, enums, domain errors.
  2. **Core Services (`internal/service`)**: Business rules + co-located unit tests (`*_test.go`).
  3. **Repository (`internal/repository`)**: SQL queries, Postgres access + co-located integration tests.
  4. **HTTP Handlers (`internal/handler`)**: Chi routes, JSON encode/decode + `httptest` tests.
  5. **Migrations & Config**: `migrations/`, `.env.example`, `internal/config/`.
  6. **Documentation & Tooling**: `README.md`, Swagger `docs/`, `Makefile`, CI workflows.

---

## 🏷️ Standardized Scopes for Barbershop API

Commit scopes must strictly match one of the allowed scopes:

| Scope | Domain / Responsibility in Go |
|---|---|
| `domain` | Pure entities (`Barber`, `Appointment`, `Service`) and domain errors |
| `service` | Business rules (scheduling conflicts, commissions, business hours) |
| `repo` | Database queries, pgx pool, sqlc repositories |
| `handler` | HTTP handlers, Chi routes, JSON validation, HTTP responses |
| `config` | Application configuration, environment variables, timeouts |
| `migrations` | SQL DDL migration files (`migrations/*.sql`) |
| `api` | Server initialization, Chi router, middleware chain, Graceful Shutdown |
| `docs` | Swagger/OpenAPI docs, architecture guides, README |
| `ci` | GitHub Actions, `.vscode`, Dockerfile, Makefile |
| `deps` | Go module updates (`go.mod`, `go.sum`) |
| `test` | Test helpers, test fixtures, integration test infrastructure |

---

## 📝 Staging & Commit Workflow for Agents

Instead of staging everything at once with `git add .`:

```bash
# 1. Stage and commit domain entities & errors
git add internal/domain/appointment.go
git commit -m "feat(domain): add appointment status and conflict errors"

# 2. Stage and commit service + co-located test
git add internal/service/appointment.go internal/service/appointment_test.go
git commit -m "feat(service): implement appointment scheduling conflict check"

# 3. Stage and commit repository implementation
git add internal/repository/appointment.go internal/repository/appointment_test.go
git commit -m "feat(repo): add postgres appointment persistence methods"

# 4. Stage and commit HTTP handler + Swagger update
git add internal/handler/appointment.go internal/handler/appointment_test.go docs/
git commit -m "feat(handler): add POST /api/v1/appointments endpoint"

# 5. Stage and commit documentation or config
git add README.md .env.example
git commit -m "docs(api): document appointment booking endpoint in README"
```

