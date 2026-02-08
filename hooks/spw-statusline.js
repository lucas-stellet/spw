#!/usr/bin/env node
// SPW Statusline - GSD-style (fail-open)
// Shows: model | current task | directory | spec | context usage
//
// Spec detection strategy (fast -> slow):
// 1) Cached spec (TTL, default 10s)
// 2) Git diff vs base branch (detect spec touched in current branch)
// 3) Fallback to most-recently-modified spec artifacts (mtime)

const fs = require('fs');
const path = require('path');
const os = require('os');
const { execFileSync } = require('child_process');

const DEFAULT_CACHE_TTL_SECONDS = 10;
const DEFAULT_BASE_BRANCHES = ['main', 'master', 'staging', 'develop'];

function runGit(args, cwd) {
  try {
    return execFileSync('git', args, {
      cwd,
      stdio: ['ignore', 'pipe', 'ignore']
    })
      .toString()
      .trim();
  } catch (_) {
    return null;
  }
}

function getRepoRoot(dir) {
  const out = runGit(['rev-parse', '--show-toplevel'], dir);
  return out && out.length ? out : null;
}

function resolveRuntimeConfigPath(repoRoot) {
  const preferred = path.join(repoRoot, '.spw', 'spw-config.toml');
  const legacy = path.join(repoRoot, '.spec-workflow', 'spw-config.toml');
  if (fs.existsSync(preferred)) return preferred;
  if (fs.existsSync(legacy)) return legacy;
  return preferred;
}

function readStatuslineConfig(repoRoot) {
  const configPath = resolveRuntimeConfigPath(repoRoot);
  const config = {
    baseBranches: DEFAULT_BASE_BRANCHES.slice(),
    cacheTtlSeconds: DEFAULT_CACHE_TTL_SECONDS,
    stickySpec: false
  };

  if (!fs.existsSync(configPath)) return config;

  try {
    const lines = fs.readFileSync(configPath, 'utf8').split(/\r?\n/);
    let inStatusline = false;

    for (const raw of lines) {
      const line = raw.trim();
      if (!line || line.startsWith('#')) continue;

      if (line.startsWith('[')) {
        inStatusline = line === '[statusline]';
        continue;
      }

      if (!inStatusline) continue;

      if (line.startsWith('base_branches')) {
        const matches = [...line.matchAll(/"([^"]+)"/g)].map((m) => m[1]);
        if (matches.length > 0) config.baseBranches = matches;
      } else if (line.startsWith('cache_ttl_seconds')) {
        const match = line.match(/cache_ttl_seconds\s*=\s*([0-9]+)/);
        if (match) config.cacheTtlSeconds = parseInt(match[1], 10);
      } else if (line.startsWith('sticky_spec')) {
        const match = line.match(/sticky_spec\s*=\s*(true|false)/i);
        if (match) config.stickySpec = match[1].toLowerCase() === 'true';
      }
    }
  } catch (_) {
    return config;
  }

  return config;
}

function getCachePaths(repoRoot) {
  const cacheDir = path.join(repoRoot, '.spec-workflow', '.spw-cache');
  return {
    cacheDir,
    cacheFile: path.join(cacheDir, 'statusline.json')
  };
}

function readCache(cacheFile, ttlSeconds, ignoreTtl) {
  try {
    if (!fs.existsSync(cacheFile)) return '';
    const data = JSON.parse(fs.readFileSync(cacheFile, 'utf8'));
    if (!data || !data.spec || !data.ts) return '';

    if (ignoreTtl) return String(data.spec);

    const ageMs = Date.now() - Number(data.ts);
    if (ageMs <= ttlSeconds * 1000) return String(data.spec);
  } catch (_) {
    return '';
  }
  return '';
}

function writeCache(cacheDir, cacheFile, spec, meta = {}) {
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
  } catch (_) {
    // fail-open
  }
}

function detectBaseRef(repoRoot, baseBranches) {
  const upstream = runGit(
    ['rev-parse', '--abbrev-ref', '--symbolic-full-name', '@{u}'],
    repoRoot
  );
  if (upstream) return upstream;

  const branches = baseBranches && baseBranches.length ? baseBranches : DEFAULT_BASE_BRANCHES;
  for (const base of branches) {
    const refs = [base, `origin/${base}`, `upstream/${base}`];
    for (const ref of refs) {
      const exists = runGit(['rev-parse', '--verify', ref], repoRoot);
      if (exists) return ref;
    }
  }

  return null;
}

function detectSpecFromGit(repoRoot, baseBranches) {
  const baseRef = detectBaseRef(repoRoot, baseBranches);
  if (!baseRef) return '';

  const diff = runGit(['diff', '--name-only', `${baseRef}...HEAD`], repoRoot);
  if (!diff) return '';

  const lines = diff.split(/\r?\n/).filter(Boolean);
  const candidates = new Map();

  lines.forEach((line, idx) => {
    const match = line.match(/\.spec-workflow\/specs\/([^/]+)\//);
    if (!match) return;

    const name = match[1];
    let score = 1;

    if (/\/(requirements|design|tasks)\.md$/.test(line)) {
      score = 3;
    } else if (/\/(DESIGN-RESEARCH|TASKS-CHECK|PRD)\.md$/.test(line)) {
      score = 2;
    }

    const prev = candidates.get(name);
    if (!prev || score > prev.score || (score === prev.score && idx < prev.idx)) {
      candidates.set(name, { score, idx });
    }
  });

  if (candidates.size === 0) return '';

  let best = null;
  for (const [name, info] of candidates.entries()) {
    if (!best || info.score > best.score || (info.score === best.score && info.idx < best.idx)) {
      best = { name, ...info };
    }
  }

  return best ? best.name : '';
}

function detectSpecByMtime(specsRoot) {
  try {
    if (!fs.existsSync(specsRoot)) return '';

    const entries = fs
      .readdirSync(specsRoot, { withFileTypes: true })
      .filter((e) => e.isDirectory());

    let latest = null;
    for (const entry of entries) {
      const specDir = path.join(specsRoot, entry.name);
      const files = ['requirements.md', 'design.md', 'tasks.md']
        .map((n) => path.join(specDir, n))
        .filter((p) => fs.existsSync(p));
      if (files.length === 0) continue;

      const mtime = Math.max(...files.map((p) => fs.statSync(p).mtimeMs || 0));
      if (!latest || mtime > latest.mtime) {
        latest = { name: entry.name, mtime };
      }
    }

    return latest ? latest.name : '';
  } catch (_) {
    return '';
  }
}

function detectActiveSpec(dir) {
  const repoRoot = getRepoRoot(dir);
  const specsRoot = repoRoot
    ? path.join(repoRoot, '.spec-workflow', 'specs')
    : path.join(dir, '.spec-workflow', 'specs');

  if (!fs.existsSync(specsRoot)) return '';

  if (!repoRoot) return detectSpecByMtime(specsRoot);

  const config = readStatuslineConfig(repoRoot);
  const { cacheDir, cacheFile } = getCachePaths(repoRoot);
  if (config.stickySpec) {
    const cachedSticky = readCache(cacheFile, config.cacheTtlSeconds, true);
    if (cachedSticky) return cachedSticky;
  } else {
    const cached = readCache(cacheFile, config.cacheTtlSeconds, false);
    if (cached) return cached;
  }

  const specFromGit = detectSpecFromGit(repoRoot, config.baseBranches);
  if (specFromGit) {
    writeCache(cacheDir, cacheFile, specFromGit, { source: 'git', sticky: config.stickySpec });
    return specFromGit;
  }

  const specByMtime = detectSpecByMtime(specsRoot);
  if (specByMtime) {
    writeCache(cacheDir, cacheFile, specByMtime, { source: 'mtime', sticky: config.stickySpec });
  }
  return specByMtime || '';
}

let input = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => (input += chunk));
process.stdin.on('end', () => {
  try {
    const data = JSON.parse(input || '{}');
    const model = data.model?.display_name || data.model?.name || 'Claude';
    const dir = data.workspace?.current_dir || process.cwd();
    const session = data.session_id || '';
    const remaining = data.context_window?.remaining_percentage;

    // Context window display (scaled to 80% real usage)
    let ctx = '';
    if (remaining != null) {
      const rem = Math.round(remaining);
      const rawUsed = Math.max(0, Math.min(100, 100 - rem));
      const used = Math.min(100, Math.round((rawUsed / 80) * 100));

      const filled = Math.floor(used / 10);
      const bar = '█'.repeat(filled) + '░'.repeat(10 - filled);

      if (used < 63) {
        ctx = ` \x1b[32m${bar} ${used}%\x1b[0m`;
      } else if (used < 81) {
        ctx = ` \x1b[33m${bar} ${used}%\x1b[0m`;
      } else if (used < 95) {
        ctx = ` \x1b[38;5;208m${bar} ${used}%\x1b[0m`;
      } else {
        ctx = ` \x1b[5;31m💀 ${bar} ${used}%\x1b[0m`;
      }
    }

    // Current task from session todos (same pattern as GSD)
    let task = '';
    const homeDir = os.homedir();
    const todosDir = path.join(homeDir, '.claude', 'todos');
    if (session && fs.existsSync(todosDir)) {
      try {
        const files = fs
          .readdirSync(todosDir)
          .filter((f) => f.startsWith(session) && f.includes('-agent-') && f.endsWith('.json'))
          .map((f) => ({ name: f, mtime: fs.statSync(path.join(todosDir, f)).mtime }))
          .sort((a, b) => b.mtime - a.mtime);

        if (files.length > 0) {
          try {
            const todos = JSON.parse(fs.readFileSync(path.join(todosDir, files[0].name), 'utf8'));
            const inProgress = todos.find((t) => t.status === 'in_progress');
            if (inProgress) task = inProgress.activeForm || inProgress.content || '';
          } catch (_) {
            // ignore
          }
        }
      } catch (_) {
        // ignore
      }
    }

    const dirname = path.basename(dir);
    const spec = detectActiveSpec(dir);
    const specLabel = spec ? ` │ \x1b[2mspec:${spec}\x1b[0m` : '';

    if (task) {
      process.stdout.write(
        `\x1b[2m${model}\x1b[0m │ \x1b[1m${task}\x1b[0m │ \x1b[2m${dirname}\x1b[0m${specLabel}${ctx}`
      );
    } else {
      process.stdout.write(`\x1b[2m${model}\x1b[0m │ \x1b[2m${dirname}\x1b[0m${specLabel}${ctx}`);
    }
  } catch (_) {
    // Silent fail: never break status line / startup
    process.stdout.write(`SPW │ ${path.basename(process.cwd())}`);
  }
});
