#!/usr/bin/env node

const {
  emitViolation,
  extractPrompt,
  firstSpwCommand,
  getHookConfig,
  getWorkspaceRoot,
  hasSpecArg,
  readStdinJson
} = require("./spw-hook-lib");

const payload = readStdinJson();
const workspaceRoot = getWorkspaceRoot(payload);
const config = getHookConfig(workspaceRoot);

if (!config.enabled || !config.guardPromptRequireSpec) {
  process.exit(0);
}

const prompt = extractPrompt(payload);
if (!prompt) {
  process.exit(0);
}

const parsed = firstSpwCommand(prompt);
if (!parsed) {
  process.exit(0);
}

const requiresSpec = new Set([
  "prd",
  "plan",
  "design-research",
  "design-draft",
  "tasks-plan",
  "tasks-check",
  "exec",
  "checkpoint"
]);

if (!requiresSpec.has(parsed.command)) {
  process.exit(0);
}

if (!hasSpecArg(parsed.argsLine)) {
  emitViolation(config, `Missing <spec-name> for /spw:${parsed.command}`, [
    `Expected usage: /spw:${parsed.command} <spec-name>`,
    "Tip: use /spw:status if you need help discovering the current stage."
  ]);
}

process.exit(0);

