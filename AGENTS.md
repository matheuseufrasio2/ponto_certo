# 🤖 AGENTS.md — Development Guide & Architectural Guardrails

This document defines the **non-negotiable system invariants and architectural guardrails** for AI agents working on **ponto_certo**. Detailed domain runbooks and progressive workflows live under `.agents/skills/`.

---

## 🎯 Project Overview

**ponto_certo** is a high-performance RESTful API for barbershop scheduling and operations management, written in **Go 1.22+**. It is built on standard library idioms using **`go-chi/chi/v5`**, structured logging (**`log/slog`**), PostgreSQL (**`pgx/v5`** / **`sqlc`**), and interactive OpenAPI documentation (**Swagger**).

---

## 🛡️ Core System Invariants & Guardrails

### 1. Consumer-Defined Interfaces & Clean Boundaries (Go ISP)
- **"Accept interfaces, return structs"**: Packages must never export speculative interfaces in `domain/` or central packages.
- The `internal/service` layer defines the repository interfaces it requires to fulfill use cases:
  ```go
  type AppointmentRepository interface {
      Create(ctx context.Context, appt domain.Appointment) error
      HasConflict(ctx context.Context, barberID string, start, end time.Time) (bool, error)
  }
  ```
- Concrete implementations in `internal/repository` return concrete structs (`*PostgresAppointmentRepo`), satisfying interfaces implicitly without circular imports.
- `internal/domain` must contain **only** pure structs and domain errors (`ErrAppointmentConflict`). It must never import any other internal package, and **MUST NEVER** import `net/http` or HTTP status codes.

### 2. HTTP Request Safety, Response Helpers & Error Mapping
- Handlers **MUST** use the standard response helpers from [`internal/handler/response.go`](internal/handler/response.go):
  - `WriteJSON(w, http.StatusOK, data)`: sets `application/json` header and encodes payload (200 OK).
  - `WriteCreated(w, data)`: sends 201 Created for new resource creation.
  - `WriteNoContent(w)`: sends 204 No Content for successful updates/deletions with empty body.
  - `ReadJSON(w, r, &req)`: enforces 1MB body limit (`http.MaxBytesReader`), rejects unknown JSON fields, and returns explicit errors for malformed syntax or type mismatches.
  - `WriteError(w, status, "reason")`: returns uniform error payloads (`{"error": "reason"}`).
  - `HandleError(w, err)`: maps domain errors to standard HTTP status codes (`400`, `404`, `409`, `500`).
- **Domain-to-HTTP Error Mapping Invariant**:
  - `domain.ErrNotFound` / `Err*NotFound` $\rightarrow$ `404 Not Found`
  - `domain.ErrAppointmentConflict` $\rightarrow$ `409 Conflict`
  - `domain.ErrInvalidTimeSlot` / malformed payload $\rightarrow$ `400 Bad Request`
  - Unexpected/system errors $\rightarrow$ `500 Internal Server Error` (masked to `"internal server error"` and logged via `slog.Error`).
- **Swagger Documentation Invariant**: Every handler must declare `@Success` (200, 201, 204) and all applicable `@Failure` annotations (400, 404, 409, 500) matching the standardized taxonomy.
- Handlers must never swallow errors; every `if err != nil` must be handled with `HandleError(w, err)` or logged.

### 3. Context & Timeout Propagation
- Every service method, database query, and external HTTP call **MUST** accept `ctx context.Context` as its first parameter and pass it downstream.
- Database queries must respect query timeouts and client aborts (`r.Context()`).
- **NEVER** launch unmanaged background goroutines (`go func() { ... }()`) without a bounded context, recovery mechanism, or `sync.WaitGroup`.

### 4. Barbershop Domain Invariants
- **Monetary Values**: All monetary values (service prices, totals) **MUST** be stored as integer cents (`PriceInCents int64`), never as floating-point numbers (`float64`).
- **Timestamps**: All dates and appointment intervals must be stored and processed in **UTC** (`time.Time`).
- **Schedule Overlap Rule**: Two appointments for the same barber overlap if and only if:
  $$\text{newStart} < \text{existingEnd} \quad \text{and} \quad \text{newEnd} > \text{existingStart}$$
- **State Machine Transitions**: Appointments start in `SCHEDULED`. Valid transitions are only:
  - `SCHEDULED` $\rightarrow$ `COMPLETED`
  - `SCHEDULED` $\rightarrow$ `CANCELLED`

### 5. Structured Logging & Production Telemetry (`log/slog`)
- **NEVER** use `fmt.Println` or standard `log.Print` in production code.
- Always use `slog.Info`, `slog.Warn`, `slog.Error` with contextual structured attributes:
  ```go
  slog.Info("appointment created", "appointment_id", appt.ID, "barber_id", appt.BarberID)
  ```
- Sensitive client data (passwords, complete tokens) must never be logged.

---

## 📐 Coding Standards & Tooling

- **Language Policy**:
  - **Conversations & Plans**: All plans, discussions, and chat responses with the user must be conducted in **Portuguese**.
  - **Codebase Artifacts**: English for all code, functions, types, comments, documentation, git commits, and PR descriptions. Only the brand/app name (`ponto_certo`) remains in Portuguese.
- **Go Formatting & Tooling**:
  - Code must be formatted with standard `gofmt` (tabs for indentation, 4-space visual width).
  - All code must pass `golangci-lint` without warnings (`make dev` / `golangci-lint run`).
  - Unit and integration tests must run with race condition detection: `go test -v -race ./...`.
- **Release Please & Conventional Commits**:
  - Releases, SemVer tags (`v0.1.0`), and `CHANGELOG.md` are completely automated via **Google Release Please**.
  - Commits must strictly follow Conventional Commits (`feat:`, `fix:`, `perf:`, `chore:`, `docs:`, `test:`).
  - Commits must be atomic with single responsibilities. Commit titles **MUST NOT** contain conjunctions (`and`, `e`, `&`, `+`). Consult [.agents/skills/atomic-commits/SKILL.md](file:///.agents/skills/atomic-commits/SKILL.md).

---

## 🧰 Progressive Domain Skills (`.agents/skills/`)

For specialized tasks, consult the on-demand skills:
- [analysis-first-approval](file:///.agents/skills/analysis-first-approval/SKILL.md): Enforces an analysis-first protocol and explicit user approval before touching code.
- [atomic-commits](file:///.agents/skills/atomic-commits/SKILL.md): Enforces atomic, semantically focused commits partitioned by Go architectural layers.
- [pre-commit-docs-check](file:///.agents/skills/pre-commit-docs-check/SKILL.md): Ensures Swagger (`docs/`), `README.md`, and `.env.example` stay in sync before committing.
- [pull-request-template-check](file:///.agents/skills/pull-request-template-check/SKILL.md): Enforces PR template standards and reproducible manual testing guides.
