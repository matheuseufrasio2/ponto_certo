---
name: pre-commit-docs-check
description: >-
  Analyzes whether changes within the current task scope require updating README.md, AGENTS.md, Swagger docs, or .env.example before executing any git commit.
  Enforces a lean AGENTS.md (< 12KB) dedicated exclusively to non-negotiable guardrails and Go system invariants.
---

# Pre-Commit Documentation & Agent Guidelines Check

## Overview

This skill ensures that project documentation (`README.md`), OpenAPI/Swagger docs (`docs/`), environment variables (`.env.example`), and AI agent architectural guardrails (`AGENTS.md`) remain synchronized with codebase changes **within the scope of the current task** without inflating `AGENTS.md`.

---

## ⚠️ Core Rule: Strict Task-Scope Isolation & Lean AGENTS.md

> **CRITICAL**: The agent must **ONLY** analyze and document changes developed as part of the **current task / active user request**.
>
> - **DO NOT** document unrelated files or pre-existing uncommitted changes in the repository.
> - **DO NOT** update `README.md` or `AGENTS.md` for features, scripts, or refactors outside the task scope.
> - **LEAN AGENTS.MD RULE**: `AGENTS.md` is an **Always-On Rule** injected into the model's context on every single turn. It is strictly reserved for **inviolable system invariants and breaking guardrails** (e.g., consumer-defined interfaces, money stored as integer cents, context propagation, HTTP safety with `WriteJSON`/`ReadJSON`).
> - **DO NOT dump endpoint catalogs, complete SQL schemas, or long tutorials into `AGENTS.md`**.
> - Domain runbooks and detailed technical guides belong in **dedicated skills** under `.agents/skills/<name>/SKILL.md` (which use Progressive Disclosure and consume ~100 tokens when idle).
> - **HARD LIMIT**: `AGENTS.md` must **NEVER** exceed 12,000 bytes (~2,000 tokens) to prevent context budget bloat.

---

## When to Run

- **Mandatory Trigger**: Run this analysis before running any `git commit` command.
- **On Demand**: When the user requests a commit, asks to review docs before committing, or asks to finalize a feature/bugfix.

---

## Workflow Steps

### Step 1: Identify Task Scope & Filter Changes

1. Explicitly list the files created, modified, or deleted **specifically for the current task**.
2. Isolate your diff analysis to those files only:
   ```bash
   git diff -- <task-file-1> <task-file-2> ...
   ```
3. Ignore any other modified or untracked files present in `git status` that belong to other tasks.

---

### Step 2: Evaluate `AGENTS.md` vs Skills Criteria (Task-Scoped Only)

#### A. When to edit `AGENTS.md` (Strict Guardrails Only):
Did this task introduce or modify a **non-negotiable system invariant** that future agents could violate and break the system?
- A new core architectural rule (e.g., interface placement, transaction management in services).
- A mandatory database invariant (e.g., timezone policy, soft-deletion handling).
- A mandatory compiler/tooling invariant (e.g., linter rule, build flag).

👉 **Decision for `AGENTS.md`**:
- **YES**: Add a concise rule (1-3 lines max) in English under the relevant section in `AGENTS.md`. Verify file size remains `< 12,000 bytes`.
- **NO**: Leave `AGENTS.md` unchanged.

#### B. When to create or update an On-Demand Skill (`.agents/skills/`):
Did this task introduce a complex domain workflow, multi-step subsystem (e.g., payment gateway webhook, Google Calendar two-way sync)?
👉 **Decision for Skills**:
- **YES**: Document the guide in `.agents/skills/<name>/SKILL.md` using Progressive Disclosure.
- **NO**: Keep documentation in Go docstrings/comments.

---

### Step 3: Evaluate Swagger, `.env.example`, and `README.md` Criteria

Check if the changes made **in this task** introduced or modified:

1. **HTTP Endpoints or Payload Schemas**:
   - New or updated handlers, parameters, or response structures.
   - 👉 **Action**: Ensure godoc annotations are added to the handler and run `make docs` to regenerate `docs/`.
2. **Environment Variables**:
   - New configuration variables added to `internal/config/config.go`.
   - 👉 **Action**: Add the variable with an explanatory comment to `.env.example`.
3. **Database Schema**:
   - New migration files in `migrations/`.
   - 👉 **Action**: Document tables and relationships in `README.md` if user-facing.

---

### Step 4: Stage and Commit (Strictly Task-Scoped)

1. Stage **only** the files belonging to the current task and any corresponding doc updates (`README.md`, `docs/`, `.env.example`, `AGENTS.md`):
   ```bash
   git add <task-files> docs/ .env.example README.md
   ```
   > ⚠️ **NEVER** use blind `git add .` or `git add -A`.

2. Commit using Conventional Commits with standardized scopes.

