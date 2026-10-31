---
name: pull-request-template-check
description: >-
  Ensures all Pull Requests created via CLI or tools strictly adhere to the repository's PR template (.github/pull_request_template.md).
  Use this skill whenever creating or updating Pull Requests in this repository.
---

# Pull Request Template Enforcement

## 🎯 Purpose

This skill ensures that every Pull Request created or edited in this repository strictly adheres to the standard format defined in [`.github/pull_request_template.md`](../../.github/pull_request_template.md).

---

## 📋 Required Structure

Whenever opening or updating a PR (e.g. using `gh pr create` or `gh pr edit`), the PR description (`--body`) **MUST** include all sections from `.github/pull_request_template.md`:

1. **`## 📝 Summary & Motivation`**:
   - High-level context of why the change is needed and what problem it solves in the Barbershop API.
   - **`### 🔑 Key Changes`**: Bulleted list of domain, service, handler, or database changes.
   - Issue references if applicable (`Closes #`, `Ref #`).
2. **`## 📌 Type of Change`**:
   - Check the relevant checkbox corresponding to the Conventional Commit type (e.g., `- [x] \`feat\`: ...` or `- [x] \`fix\`: ...`).
3. **`## 🔍 Architecture & Implementation Details`**:
   - Summary of architectural impacts, consumer-defined repository interfaces, migrations, concurrency safety, or context propagation.
4. **`## ✅ Quality Checklist`**:
   - Check all applicable quality gates:
     - `- [x] Commit messages follow the **Conventional Commits** specification (\`feat:\`, \`fix:\`, \`chore:\`, etc.).`
     - `- [x] Automated tests pass with race condition detector (\`go test -v -race ./...\`).`
     - `- [x] Swagger documentation is updated (\`make docs\`).`
     - `- [x] No sensitive credentials or \`.env\` files were accidentally committed.`
     - `- [x] Documentation (\`README.md\`, \`.env.example\` and/or \`AGENTS.md\`) has been updated if applicable.`
5. **`## 🧪 Manual Testing Guide`** *(Mandatory at the end of the PR)*:
   - **`### ⚙️ Prerequisites`**: Environment requirements (e.g. `make dev`, PostgreSQL running).
   - **`### 🪜 Step-by-Step Verification`**: Numbered, reproducible steps with `curl` or Swagger UI commands.
   - **`### 🎯 Expected Results`**: Clear success criteria and expected HTTP status codes/payloads.
   - **`### 📸 Evidence (Optional)`**: Terminal logs or curl responses demonstrating verification.

---

## 🤖 Guidelines for AI Agents

- **🌐 Strict English-Only Rule (Mandatory)**:
  - **CRITICAL INVARIANT**: PR title, description, section content, bullet points, testing guides, and all PR metadata **MUST BE WRITTEN 100% IN ENGLISH**.
  - All repository git artifacts, commits, and Pull Requests belong to the public engineering record.
- **Never omit sections**: Do not replace the template with a custom ad-hoc format.
- **Always provide concrete manual testing steps**: Do not leave placeholder text in `## 🧪 Manual Testing Guide`. Write specific `curl` commands, JSON payloads, and expected responses tailored to the PR.
- **Mark completed items**: Use `[x]` for items that have been verified and applied.

