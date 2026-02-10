#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

function normalizeSlashes(value) {
  return String(value || "").replace(/\\/g, "/");
}

function readStdinJson() {
  try {
    const raw = fs.readFileSync(0, "utf8");
    if (!raw || !raw.trim()) return {};
    return JSON.parse(raw);
  } catch (_error) {
    return {};
  }
}

function getWorkspaceRoot(payload) {
  const candidates = [
    payload.cwd,
    payload?.workspace?.current_dir,
    process.env.CLAUDE_PROJECT_DIR,
    process.cwd()
  ];

  for (const candidate of candidates) {
    if (!candidate) continue;
    const abs = path.resolve(String(candidate));
    if (fs.existsSync(abs) && fs.statSync(abs).isDirectory()) {
      return abs;
    }
  }
  return process.cwd();
}

function readTomlValue(filePath, section, key, defaultValue) {
  if (!fs.existsSync(filePath)) return defaultValue;

  const lines = fs.readFileSync(filePath, "utf8").split(/\r?\n/);
  let inSection = false;

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    if (trimmed.startsWith("[") && trimmed.endsWith("]")) {
      inSection = trimmed === `[${section}]`;
      continue;
    }
    if (!inSection) continue;

    const match = trimmed.match(new RegExp(`^${key}\\s*=\\s*(.+)$`));
    if (!match) continue;

    let value = match[1].replace(/\s+#.*$/, "").trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    return value;
  }

  return defaultValue;
}

function toBool(value, fallback) {
  if (value == null || value === "") return fallback;
  const normalized = String(value).toLowerCase();
  if (["true", "1", "yes", "on"].includes(normalized)) return true;
  if (["false", "0", "no", "off"].includes(normalized)) return false;
  return fallback;
}

function toInt(value, fallback) {
  const parsed = parseInt(String(value), 10);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function resolveRuntimeConfigPath(workspaceRoot) {
  const canonical = path.join(workspaceRoot, ".spec-workflow", "spw-config.toml");
  const fallback = path.join(workspaceRoot, ".spw", "spw-config.toml");
  if (fs.existsSync(canonical)) return canonical;
  if (fs.existsSync(fallback)) return fallback;
  return canonical;
}

function getHookConfig(workspaceRoot) {
  const configPath = resolveRuntimeConfigPath(workspaceRoot);
  const defaults = {
    enabled: true,
    enforcementMode: "warn",
    verbose: true,
    recentRunWindowMinutes: 30,
    guardPromptRequireSpec: true,
    guardPaths: true,
    guardWaveLayout: true,
    guardStopHandoff: true
  };

  const enabled = toBool(readTomlValue(configPath, "hooks", "enabled", defaults.enabled), defaults.enabled);
  const enforcementMode = String(
    readTomlValue(configPath, "hooks", "enforcement_mode", defaults.enforcementMode)
  ).toLowerCase();
  const verbose = toBool(readTomlValue(configPath, "hooks", "verbose", defaults.verbose), defaults.verbose);
  const recentRunWindowMinutes = toInt(
    readTomlValue(configPath, "hooks", "recent_run_window_minutes", defaults.recentRunWindowMinutes),
    defaults.recentRunWindowMinutes
  );
  const guardPromptRequireSpec = toBool(
    readTomlValue(configPath, "hooks", "guard_prompt_require_spec", defaults.guardPromptRequireSpec),
    defaults.guardPromptRequireSpec
  );
  const guardPaths = toBool(
    readTomlValue(configPath, "hooks", "guard_paths", defaults.guardPaths),
    defaults.guardPaths
  );
  const guardWaveLayout = toBool(
    readTomlValue(configPath, "hooks", "guard_wave_layout", defaults.guardWaveLayout),
    defaults.guardWaveLayout
  );
  const guardStopHandoff = toBool(
    readTomlValue(configPath, "hooks", "guard_stop_handoff", defaults.guardStopHandoff),
    defaults.guardStopHandoff
  );

  return {
    enabled,
    enforcementMode: enforcementMode === "block" ? "block" : "warn",
    verbose,
    recentRunWindowMinutes,
    guardPromptRequireSpec,
    guardPaths,
    guardWaveLayout,
    guardStopHandoff,
    configPath
  };
}

function emitViolation(config, title, details) {
  const lines = [`[spw-hook] ${title}`];
  if (Array.isArray(details)) {
    for (const item of details) lines.push(`[spw-hook] - ${item}`);
  }

  for (const line of lines) {
    console.error(line);
  }

  if (config.enforcementMode === "block") {
    process.exit(2);
  }
  process.exit(0);
}

function emitInfo(config, message) {
  if (config.verbose) {
    console.error(`[spw-hook] ${message}`);
  }
}

function getToolInput(payload) {
  return payload.tool_input || payload.toolInput || payload.input || {};
}

function getToolName(payload) {
  return (
    payload.tool_name ||
    payload.toolName ||
    payload?.tool?.name ||
    payload?.input?.tool_name ||
    ""
  );
}

function resolveTargetPath(payload, workspaceRoot) {
  const input = getToolInput(payload);
  const filePath = input.file_path || input.path || input.target_path || input.filename;
  if (!filePath || typeof filePath !== "string") return null;

  const cwd = payload.cwd || payload?.workspace?.current_dir || workspaceRoot;
  const absPath = path.isAbsolute(filePath) ? filePath : path.resolve(String(cwd), filePath);
  const relPath = normalizeSlashes(path.relative(workspaceRoot, absPath));

  return {
    raw: filePath,
    absPath,
    relPath
  };
}

function extractPrompt(payload) {
  const candidates = [
    payload.prompt,
    payload.userPrompt,
    payload.user_prompt,
    payload.message,
    payload?.input?.prompt,
    payload?.input?.message
  ];
  return candidates.find((value) => typeof value === "string" && value.trim()) || "";
}

function firstSpwCommand(prompt) {
  const lines = String(prompt || "").split(/\r?\n/);
  for (const line of lines) {
    const trimmed = line.trim();
    const match = trimmed.match(/^\/spw:([a-z-]+)(?:\s+(.*))?$/i);
    if (match) {
      return {
        command: match[1].toLowerCase(),
        argsLine: (match[2] || "").trim()
      };
    }
  }
  return null;
}

function tokenizeArgs(argsLine) {
  if (!argsLine) return [];
  return argsLine.match(/"[^"]*"|'[^']*'|\S+/g) || [];
}

function extractSpecArg(argsLine) {
  const tokens = tokenizeArgs(argsLine)
    .map((token) => token.replace(/^["']|["']$/g, ""))
    .filter(Boolean);
  if (tokens.length === 0) return "";

  // Spec must be the first positional argument. If command starts with a flag,
  // treat as missing spec instead of accidentally using option values.
  const first = tokens[0];
  if (first.startsWith("--")) return "";
  return first;
}

function hasSpecArg(argsLine) {
  return extractSpecArg(argsLine) !== "";
}

function writeStatuslineCache(workspaceRoot, spec, meta = {}) {
  if (!spec) return false;
  const cacheDir = path.join(workspaceRoot, ".spec-workflow", ".spw-cache");
  const cacheFile = path.join(cacheDir, "statusline.json");

  try {
    fs.mkdirSync(cacheDir, { recursive: true });
    fs.writeFileSync(
      cacheFile,
      JSON.stringify(
        {
          ts: Date.now(),
          spec,
          ...meta
        },
        null,
        2
      )
    );
    return true;
  } catch (_error) {
    return false;
  }
}

function listSpecDirs(workspaceRoot) {
  const specsRoot = path.join(workspaceRoot, ".spec-workflow", "specs");
  if (!fs.existsSync(specsRoot)) return [];
  const entries = fs.readdirSync(specsRoot, { withFileTypes: true });
  return entries
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(specsRoot, entry.name));
}

function listDirSafe(dirPath) {
  try {
    return fs.readdirSync(dirPath, { withFileTypes: true });
  } catch (_error) {
    return [];
  }
}

function collectRunDirs(specDir) {
  const commsRoot = path.join(specDir, "_agent-comms");
  if (!fs.existsSync(commsRoot)) return [];
  const runs = [];

  const simpleCommands = ["design-research", "prd", "tasks-plan", "tasks-check", "checkpoint"];
  for (const command of simpleCommands) {
    const commandRoot = path.join(commsRoot, command);
    for (const entry of listDirSafe(commandRoot)) {
      if (!entry.isDirectory()) continue;
      runs.push(path.join(commandRoot, entry.name));
    }
  }

  const wavesRoot = path.join(commsRoot, "waves");
  for (const waveEntry of listDirSafe(wavesRoot)) {
    if (!waveEntry.isDirectory()) continue;
    const waveDir = path.join(wavesRoot, waveEntry.name);
    for (const stage of ["execution", "checkpoint", "post-check"]) {
      const stageDir = path.join(waveDir, stage);
      for (const runEntry of listDirSafe(stageDir)) {
        if (!runEntry.isDirectory()) continue;
        runs.push(path.join(stageDir, runEntry.name));
      }
    }
  }

  return runs;
}

function isRecent(runDir, nowMs, windowMs) {
  try {
    const stats = fs.statSync(runDir);
    return nowMs - stats.mtimeMs <= windowMs;
  } catch (_error) {
    return false;
  }
}

function checkRunCompleteness(runDir) {
  const issues = [];
  const handoffPath = path.join(runDir, "_handoff.md");
  if (!fs.existsSync(handoffPath)) {
    issues.push("missing _handoff.md");
  }

  const subagentDirs = listDirSafe(runDir).filter((entry) => entry.isDirectory());
  for (const entry of subagentDirs) {
    const subagentDir = path.join(runDir, entry.name);
    const missing = [];
    for (const fileName of ["brief.md", "report.md", "status.json"]) {
      if (!fs.existsSync(path.join(subagentDir, fileName))) {
        missing.push(fileName);
      }
    }
    if (missing.length > 0) {
      issues.push(`${entry.name}: missing ${missing.join(", ")}`);
    }
  }

  return issues;
}

module.exports = {
  checkRunCompleteness,
  collectRunDirs,
  emitInfo,
  emitViolation,
  extractSpecArg,
  extractPrompt,
  firstSpwCommand,
  getHookConfig,
  getToolInput,
  getToolName,
  getWorkspaceRoot,
  hasSpecArg,
  isRecent,
  listSpecDirs,
  normalizeSlashes,
  readStdinJson,
  resolveTargetPath,
  writeStatuslineCache
};
