#!/usr/bin/env node

const path = require("path");
const {
  emitViolation,
  getHookConfig,
  getWorkspaceRoot,
  normalizeSlashes,
  readStdinJson,
  resolveTargetPath
} = require("./spw-hook-lib");

function isManagedArtifactFile(baseName) {
  return (
    baseName === "DESIGN-RESEARCH.md" ||
    baseName === "TASKS-CHECK.md" ||
    baseName === "CHECKPOINT-REPORT.md" ||
    baseName === "STATUS-SUMMARY.md" ||
    /^SKILLS-[A-Z0-9-]+\.md$/i.test(baseName) ||
    /^PRD(?:-[A-Z0-9-]+)?\.md$/i.test(baseName) ||
    /^PRD-SOURCE-NOTES\.md$/i.test(baseName) ||
    /^PRD-STRUCTURE\.md$/i.test(baseName) ||
    /^PRD-REVISION-(PLAN|QUESTIONS|NOTES)\.md$/i.test(baseName)
  );
}

const payload = readStdinJson();
const workspaceRoot = getWorkspaceRoot(payload);
const config = getHookConfig(workspaceRoot);

if (!config.enabled || (!config.guardPaths && !config.guardWaveLayout)) {
  process.exit(0);
}

const resolved = resolveTargetPath(payload, workspaceRoot);
if (!resolved) {
  process.exit(0);
}

const relPath = normalizeSlashes(resolved.relPath);
const baseName = path.basename(resolved.absPath);

if (config.guardPaths) {
  const isSpecLocal = relPath.includes(".spec-workflow/specs/");
  if (isManagedArtifactFile(baseName) && !isSpecLocal) {
    emitViolation(config, "SPW artifact path violation", [
      `File: ${relPath}`,
      "Managed SPW artifacts must stay under .spec-workflow/specs/<spec-name>/"
    ]);
  }
}

if (config.guardWaveLayout) {
  if (relPath.includes("agent-comms/checkpoint/")) {
    emitViolation(config, "Legacy checkpoint folder layout is not allowed", [
      `File: ${relPath}`,
      "Use: .spec-workflow/specs/<spec-name>/agent-comms/waves/wave-<NN>/checkpoint/<run-id>/"
    ]);
  }

  if (relPath.includes("agent-comms/waves/")) {
    const waveMatch = relPath.match(/agent-comms\/waves\/([^/]+)/);
    if (waveMatch) {
      const waveId = waveMatch[1];
      if (!/^wave-\d{2}$/.test(waveId)) {
        emitViolation(config, "Wave folder must use zero-padded format", [
          `Found wave folder: ${waveId}`,
          "Expected format: wave-01, wave-02, ..."
        ]);
      }
    }

    const stageMatch = relPath.match(/agent-comms\/waves\/wave-\d{2}\/([^/]+)/);
    if (stageMatch) {
      const stage = stageMatch[1];
      const allowedStages = new Set(["execution", "checkpoint", "post-check", "_wave-summary.md", "_latest.json"]);
      if (!allowedStages.has(stage)) {
        emitViolation(config, "Invalid wave stage folder", [
          `File: ${relPath}`,
          "Allowed wave entries: execution, checkpoint, post-check, _wave-summary.md, _latest.json"
        ]);
      }
    }
  }
}

process.exit(0);

