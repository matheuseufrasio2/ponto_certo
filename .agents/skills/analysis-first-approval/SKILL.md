---
name: analysis-first-approval
description: >-
  Enforces an "Analysis First" protocol whenever the user asks for analysis, investigation, debugging, or opinions.
  Prohibits executing code modifications, running mutating commands, committing, or opening PRs until the user explicitly approves.
  Mandates ending analysis responses with an explicit confirmation question before execution.
---

# Analysis First & Explicit Approval Protocol

## 🎯 Purpose

This skill ensures that whenever the user requests an analysis, investigation, bug triage, architecture review, or opinion, the AI agent **must never jump straight into modifying code, committing, or opening PRs without explicit confirmation**.

The user must always receive a comprehensive analysis first, understand the root cause and alternatives, and be given the opportunity to ask questions or approve before any code is touched.

---

## ⚡ Trigger Conditions

Activate this protocol whenever the user's request involves:
- Investigating errors, bugs, or unexpected behavior (e.g. "analyze this", "inspect logs", "why is this failing?").
- Reviewing code, architecture, or performance issues.
- Brainstorming or asking for an opinion ("what do you think?", "which approach is best?").
- Any prompt where the user asks diagnostic questions or sends log files/screenshots without an explicit instruction like "fix directly", "implement right now", or "apply fix immediately".

---

## 🚫 Strictly Prohibited Actions (During Analysis Phase)

When triggered, the agent **MUST NOT**:
1. Edit, create, or delete any source files.
2. Run mutating CLI commands that alter codebase state (e.g., git commits, merges, pushes, PR creations).
3. Assume automatic permission to execute a fix without the user's explicit approval.

---

## 📋 Required Structure of the Analysis Output

Every analysis response must follow this clear structure:

1. **Context & Root Cause**:
   - What specifically broke or failed?
   - Why did it happen (trace the mechanism down to the Go code/database/system level)?
2. **Impact & Evidence**:
   - How does this explain the user's symptoms, logs, or test errors?
3. **Alternatives & Trade-offs**:
   - Present the viable approaches to solve the issue (e.g., Option A vs Option B) and why one is recommended over another.
4. **Proposed Action Plan**:
   - Exact files to be touched.
   - Specific Go logic changes planned.
   - Tests and verification strategy (`go test -v -race`).
5. **Mandatory Closing Gate**:
   - The response **MUST ALWAYS** end with an explicit confirmation gate (e.g., in Portuguese if chatting in Portuguese: *"Essa é a análise e minha opinião, tem perguntas? Ou posso executar?"*, or in English: *"This is the analysis and recommendation. Do you have questions, or should I proceed?"*).

---

## 🚦 Execution Gate (Post-Approval Only)

Only proceed with implementing code changes, tests, commits, and PRs **after** the user responds with explicit confirmation (e.g. "proceed", "execute", "pode executar", or resolves questions and approves).
