# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Execution Plan: SPW Refactor (REFACTOR-PLAN.md)

## Context

The SPW codebase has detailed reference docs and a refactor plan (`docs/REFACTOR-PLAN.md`) but **zero implementation** has been done. The refactor standardizes dispatch patterns, directory structure, and PR review across all 11 commands.

Current state: all 13 workflows use inline `<workflow>`, `<file_handoff_protocol>`, `<resume_policy>`, `<skills_policy>` sections with old path patterns (`_generated/`...

