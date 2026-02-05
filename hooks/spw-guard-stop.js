#!/usr/bin/env node

const path = require("path");
const {
  checkRunCompleteness,
  collectRunDirs,
  emitInfo,
  emitViolation,
  getHookConfig,
  getWorkspaceRoot,
  isRecent,
  listSpecDirs,
  normalizeSlashes,
  readStdinJson
} = require("./spw-hook-lib");

const payload = readStdinJson();
const workspaceRoot = getWorkspaceRoot(payload);
const config = getHookConfig(workspaceRoot);

if (!config.enabled || !config.guardStopHandoff) {
  process.exit(0);
}

const nowMs = Date.now();
const windowMs = Math.max(1, config.recentRunWindowMinutes) * 60 * 1000;
const violations = [];

for (const specDir of listSpecDirs(workspaceRoot)) {
  const runDirs = collectRunDirs(specDir);
  for (const runDir of runDirs) {
    if (!isRecent(runDir, nowMs, windowMs)) continue;
    const issues = checkRunCompleteness(runDir);
    if (issues.length === 0) continue;

    const relativeRunDir = normalizeSlashes(path.relative(workspaceRoot, runDir));
    violations.push(`${relativeRunDir} -> ${issues.join("; ")}`);
  }
}

if (violations.length > 0) {
  emitViolation(config, "Recent run folders are missing required handoff files", [
    `Window: last ${config.recentRunWindowMinutes} minute(s)`,
    ...violations.slice(0, 20)
  ]);
}

emitInfo(config, "Stop guard passed.");
process.exit(0);

