#!/usr/bin/env node
// SPW Statusline - GSD-style (fail-open)
// Shows: model | current task | directory | spec | context usage

const fs = require('fs');
const path = require('path');
const os = require('os');

function detectActiveSpec(dir) {
  try {
    const specsRoot = path.join(dir, '.spec-workflow', 'specs');
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
