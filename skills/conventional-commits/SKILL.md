---
name: conventional-commits
description: Use when creating git commit messages, finalizing task-level commits, or reviewing commit history quality. Enforces Conventional Commits with concise, imperative subjects.
---

# Conventional Commits

Use this skill whenever you create or suggest commit messages.

## Required format

`<type>(<scope>): <subject>`

- `type`: one of `feat`, `fix`, `refactor`, `test`, `docs`, `chore`
- `scope`: short module/spec scope (for Oráculo, prefer spec name)
- `subject`: imperative mood, lower-case start, no trailing period, <= 72 chars

## Type selection

- `feat`: new behavior/capability
- `fix`: bug/regression fix
- `refactor`: internal structure change without behavior change
- `test`: tests only
- `docs`: docs only
- `chore`: maintenance/tooling/config

## Oráculo task commits

For Oráculo task-level commits, include task id in subject:

`<type>(<spec-name>): task <task-id> - <short-title>`

Examples:
- `feat(spike-signicat-language): task 2.1 - update onboarding URL builder`
- `fix(migrate-historical-fields): task 6.2 - stabilize migration report test`
- `test(migrate-historical-fields): task 6.1 - add full report integration coverage`

## Body (optional)

Add a body only when needed:
- why decision was made
- important constraints/tradeoffs
- follow-up notes

Keep body concise and factual.
