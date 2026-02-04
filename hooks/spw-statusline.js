#!/usr/bin/env node
'use strict';

// SPW statusline (fail-open):
// - model
// - project directory
// - git branch (+ dirty marker)
// - active spec + phase + task progress (best effort)
// - context usage bar
//
// If anything fails, it falls back to a minimal line and exits successfully.

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');

const ANSI = {
  dim: '\x1b[2m',
  bold: '\x1b[1m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  orange: '\x1b[38;5;208m',
  redBlink: '\x1b[5;31m',
  reset: '\x1b[0m'
};

function safeExec(cmd, args, cwd) {
  try {
    return execFileSync(cmd, args, {
      cwd,
      stdio: ['ignore', 'pipe', 'ignore'],
      encoding: 'utf8',
      timeout: 800
    }).trim();
  } catch (_) {
    return null;
  }
}

function getGitRoot(dir) {
  return safeExec('git', ['-C', dir, 'rev-parse', '--show-toplevel'], dir);
}

function getGitBranch(dir) {
  return safeExec('git', ['-C', dir, 'rev-parse', '--abbrev-ref', 'HEAD'], dir);
}

function isDirty(dir) {
  const out = safeExec('git', ['-C', dir, 'status', '--porcelain'], dir);
  return !!(out && out.length > 0);
}

function readTaskProgress(tasksPath) {
  try {
    const text = fs.readFileSync(tasksPath, 'utf8');
    const lines = text.split(/\r?\n/);
    let total = 0;
    let done = 0;
    for (const line of lines) {
      const m = line.match(/^\s*-\s*\[( |x|X)\]\s+/);
      if (!m) continue;
      total += 1;
      if (m[1].toLowerCase() === 'x') done += 1;
    }
    return total > 0 ? { done, total } : null;
  } catch (_) {
    return null;
  }
}

function detectActiveSpec(gitRoot) {
  const specsRoot = path.join(gitRoot, '.spec-workflow', 'specs');
  if (!fs.existsSync(specsRoot)) return null;

  let entries;
  try {
    entries = fs.readdirSync(specsRoot, { withFileTypes: true });
  } catch (_) {
    return null;
  }

  let best = null;

  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const specDir = path.join(specsRoot, entry.name);
    const req = path.join(specDir, 'requirements.md');
    const design = path.join(specDir, 'design.md');
    const tasks = path.join(specDir, 'tasks.md');

    const files = [
      { p: req, key: 'req' },
      { p: design, key: 'design' },
      { p: tasks, key: 'tasks' }
    ];

    let latest = 0;
    const exists = { req: false, design: false, tasks: false };
    for (const f of files) {
      try {
        const st = fs.statSync(f.p);
        exists[f.key] = true;
        const ms = st.mtimeMs || 0;
        if (ms > latest) latest = ms;
      } catch (_) {
        // ignore missing file
      }
    }

    if (latest === 0) continue;

    let phase = 'requirements';
    if (exists.tasks) phase = 'tasks';
    else if (exists.design) phase = 'design';

    const progress = exists.tasks ? readTaskProgress(tasks) : null;

    if (!best || latest > best.latest) {
      best = {
        spec: entry.name,
        phase,
        progress,
        latest
      };
    }
  }

  return best;
}

function formatContext(remainingPct) {
  if (typeof remainingPct !== 'number' || Number.isNaN(remainingPct)) return '';

  const rem = Math.max(0, Math.min(100, Math.round(remainingPct)));
  const rawUsed = 100 - rem;
  const used = Math.max(0, Math.min(100, Math.round((rawUsed / 80) * 100)));

  const seg = 10;
  const filled = Math.max(0, Math.min(seg, Math.floor(used / 10)));
  const bar = `${'#'.repeat(filled)}${'.'.repeat(seg - filled)}`;

  if (used < 63) return ` ${ANSI.green}${bar} ${used}%${ANSI.reset}`;
  if (used < 81) return ` ${ANSI.yellow}${bar} ${used}%${ANSI.reset}`;
  if (used < 95) return ` ${ANSI.orange}${bar} ${used}%${ANSI.reset}`;
  return ` ${ANSI.redBlink}${bar} ${used}%${ANSI.reset}`;
}

function buildLine(data) {
  const dir = (data && data.workspace && data.workspace.current_dir) || process.cwd();
  const model =
    (data && data.model && (data.model.display_name || data.model.name)) ||
    'Claude';

  const ctx = formatContext(
    data && data.context_window ? data.context_window.remaining_percentage : null
  );

  const project = path.basename(dir);
  const gitRoot = getGitRoot(dir) || dir;
  const branch = getGitBranch(dir);
  const dirty = branch && isDirty(dir) ? '*' : '';
  const specInfo = detectActiveSpec(gitRoot);

  const parts = [];
  parts.push(`${ANSI.dim}${model}${ANSI.reset}`);
  parts.push(`${ANSI.bold}${project}${ANSI.reset}`);

  if (branch) {
    parts.push(`${ANSI.dim}${branch}${dirty}${ANSI.reset}`);
  }

  if (specInfo) {
    const specLabel = `spec:${specInfo.spec}`;
    const phaseLabel = `phase:${specInfo.phase}`;
    if (specInfo.progress) {
      parts.push(`${ANSI.dim}${specLabel}${ANSI.reset}`);
      parts.push(`${ANSI.dim}${phaseLabel}${ANSI.reset}`);
      parts.push(
        `${ANSI.dim}tasks:${specInfo.progress.done}/${specInfo.progress.total}${ANSI.reset}`
      );
    } else {
      parts.push(`${ANSI.dim}${specLabel}${ANSI.reset}`);
      parts.push(`${ANSI.dim}${phaseLabel}${ANSI.reset}`);
    }
  }

  return `${parts.join(' | ')}${ctx}`;
}

function outputFallback() {
  const project = path.basename(process.cwd());
  process.stdout.write(`SPW | ${project}`);
}

let raw = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => {
  raw += chunk;
});
process.stdin.on('end', () => {
  try {
    const parsed = raw.trim() ? JSON.parse(raw) : {};
    process.stdout.write(buildLine(parsed));
  } catch (_) {
    outputFallback();
  }
});

process.stdin.on('error', () => {
  outputFallback();
});

if (process.stdin.isTTY) {
  outputFallback();
}
