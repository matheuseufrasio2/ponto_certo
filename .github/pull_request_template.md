## 📝 Summary & Motivation

<!-- Provide a clear, concise overview of what this PR does and why it is needed. -->
<!-- Explain the problem solved, feature added, or technical debt addressed in the Barbershop API. -->

### 🔑 Key Changes

- <!-- Bullet point summarizing domain or business rule changes -->
- <!-- Bullet point summarizing HTTP endpoints or request/response updates -->
- <!-- Bullet point detailing database migrations, queries, or config changes -->

<!-- References to issues, if any (e.g. Closes #123, Ref #456) -->

---

## 📌 Type of Change

Select all that apply:
- [ ] `feat`: New feature (bumps Minor version via Release Please)
- [ ] `fix`: Bug fix (bumps Patch version via Release Please)
- [ ] `perf`: Performance improvement (bumps Patch version)
- [ ] `refactor`: Code refactoring without behavioral changes
- [ ] `chore`: Maintenance, dependencies, or tooling updates
- [ ] `docs`: Documentation updates only (Swagger / README)
- [ ] `test`: Adding or updating unit/integration tests

> ⚠️ **Breaking Change?** If this PR introduces breaking changes, use `feat!:` in the commit message or include `BREAKING CHANGE:` in the commit footer.

---

## 🔍 Architecture & Implementation Details

<!-- Highlight any key design decisions, repository changes, or side effects: -->
<!-- - Consumer-defined repository interfaces in internal/service -->
<!-- - SQL migrations and indexes added in migrations/ -->
<!-- - Context propagation and timeout handling -->
<!-- - Memory, goroutine, or concurrency safety considerations -->

---

## ✅ Quality Checklist

- [ ] Commit messages strictly follow the **Conventional Commits** specification (`feat:`, `fix:`, `chore:`, etc.).
- [ ] Automated tests pass with race condition detector (`go test -v -race ./...`).
- [ ] Swagger documentation is up to date (`make docs` generated `docs/`).
- [ ] No secrets, sensitive credentials, or `.env` files are committed.
- [ ] Documentation (`README.md`, `.env.example` and/or `AGENTS.md`) is updated if applicable.

---

## 🧪 Manual Testing Guide

<!-- Mandatory step-by-step instructions for manual validation of the API -->

### ⚙️ Prerequisites

1. API running locally (`make dev`) or database running (`docker-compose up -d`).
2. Valid environment variables loaded (`.env`).

### 🪜 Step-by-Step Verification

1. Send HTTP request using `curl`, Swagger UI (`http://localhost:8080/swagger/index.html`), or HTTP client:
   ```bash
   curl -X POST http://localhost:8080/api/v1/... \
     -H "Content-Type: application/json" \
     -d '{"example": "value"}'
   ```
2. Inspect the HTTP status code and response payload.
3. Test edge cases (e.g., overlapping appointment schedule, invalid payload, missing fields).

### 🎯 Expected Results

- **Success Criteria**: Describe the expected status code (e.g. `200 OK`, `201 Created`) and payload structure.
- **Error Handling**: Describe how validation errors (e.g. `400 Bad Request`, `409 Conflict`) are returned.

### 📸 Evidence (Optional)

<!-- Attach terminal output, curl response, or Swagger screenshots demonstrating validation -->

