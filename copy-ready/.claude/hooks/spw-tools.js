#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function output(result, rawValue, raw) {
  if (raw) {
    process.stdout.write(String(rawValue ?? ''));
    return;
  }
  process.stdout.write(JSON.stringify(result, null, 2));
}

function fail(message, raw) {
  const result = { ok: false, error: message };
  if (raw) {
    process.stdout.write('');
    process.exit(1);
  }
  process.stderr.write(JSON.stringify(result, null, 2) + '\n');
  process.exit(1);
}

function isTruthy(v) {
  const n = String(v || '').toLowerCase();
  return ['true', '1', 'yes', 'on'].includes(n);
}

function readToml(filePath) {
  const out = {};
  if (!fs.existsSync(filePath)) return out;

  const lines = fs.readFileSync(filePath, 'utf8').split(/\r?\n/);
  let section = null;
  let collectingArray = null;

  for (const rawLine of lines) {
    const noInline = rawLine.replace(/\s+#.*$/, '');
    const line = noInline.trim();
    if (!line || line.startsWith('#')) continue;

    if (collectingArray) {
      const matches = [...line.matchAll(/"([^"]+)"|'([^']+)'/g)];
      collectingArray.values.push(...matches.map((m) => m[1] || m[2]));

      if (line.includes(']')) {
        if (!out[collectingArray.section]) out[collectingArray.section] = {};
        out[collectingArray.section][collectingArray.key] = collectingArray.values;
        collectingArray = null;
      }
      continue;
    }

    if (line.startsWith('[') && line.endsWith(']')) {
      section = line.slice(1, -1).trim();
      if (!out[section]) out[section] = {};
      continue;
    }

    const idx = line.indexOf('=');
    if (idx === -1 || !section) continue;
    const key = line.slice(0, idx).trim();
    const valueRaw = line.slice(idx + 1).trim();

    let value;
    if (valueRaw.startsWith('[') && valueRaw.endsWith(']')) {
      const matches = [...valueRaw.matchAll(/"([^"]+)"|'([^']+)'/g)];
      value = matches.map((m) => m[1] || m[2]);
    } else if (valueRaw.startsWith('[')) {
      const matches = [...valueRaw.matchAll(/"([^"]+)"|'([^']+)'/g)];
      collectingArray = {
        section,
        key,
        values: matches.map((m) => m[1] || m[2])
      };
      continue;
    } else if (
      (valueRaw.startsWith('"') && valueRaw.endsWith('"')) ||
      (valueRaw.startsWith("'") && valueRaw.endsWith("'"))
    ) {
      value = valueRaw.slice(1, -1);
    } else if (/^(true|false)$/i.test(valueRaw)) {
      value = /^true$/i.test(valueRaw);
    } else if (/^-?[0-9]+$/.test(valueRaw)) {
      value = parseInt(valueRaw, 10);
    } else {
      value = valueRaw;
    }

    out[section][key] = value;
  }

  return out;
}

function resolveConfigPath(cwd) {
  const canonical = path.join(cwd, '.spec-workflow', 'spw-config.toml');
  const fallback = path.join(cwd, '.spw', 'spw-config.toml');
  if (fs.existsSync(canonical)) return { path: canonical, source: 'canonical' };
  if (fs.existsSync(fallback)) return { path: fallback, source: 'fallback' };
  return { path: canonical, source: 'missing' };
}

function getConfigValue(config, sectionDotKey) {
  const parts = sectionDotKey.split('.');
  if (parts.length !== 2) return undefined;
  const [section, key] = parts;
  return config?.[section]?.[key];
}

function listDirs(absPath) {
  if (!fs.existsSync(absPath)) return [];
  return fs
    .readdirSync(absPath, { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => ({ name: d.name, full: path.join(absPath, d.name) }));
}

function readJson(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch {
    return null;
  }
}

function runDirInspection(runDir) {
  const issues = [];
  const subagents = [];
  const handoff = path.join(runDir, '_handoff.md');
  if (!fs.existsSync(handoff)) {
    issues.push('missing:_handoff.md');
  }

  const entries = listDirs(runDir).filter((d) => !d.name.startsWith('_') && !d.name.startsWith('.'));
  for (const entry of entries) {
    const required = ['brief.md', 'report.md', 'status.json'];
    const missing = required.filter((f) => !fs.existsSync(path.join(entry.full, f)));
    if (missing.length > 0) {
      issues.push(`missing:${entry.name}:${missing.join(',')}`);
    }

    const statusPath = path.join(entry.full, 'status.json');
    const status = readJson(statusPath);
    if (status && String(status.status || '').toLowerCase() === 'blocked') {
      issues.push(`blocked:${entry.name}`);
    }

    subagents.push({ name: entry.name, missing, blocked: !!(status && String(status.status || '').toLowerCase() === 'blocked') });
  }

  return {
    unfinished: issues.length > 0,
    issues,
    subagents
  };
}

function cmdConfigGet(cwd, args, raw) {
  const key = args[0];
  if (!key) fail('config get requires <section.key>', raw);

  let defaultValue = undefined;
  const defIdx = args.indexOf('--default');
  if (defIdx !== -1 && args[defIdx + 1] !== undefined) {
    defaultValue = args[defIdx + 1];
  }

  const resolved = resolveConfigPath(cwd);
  const config = readToml(resolved.path);
  let value = getConfigValue(config, key);

  if (value === undefined) {
    value = defaultValue;
  }

  const result = {
    ok: true,
    key,
    value,
    config_path: path.relative(cwd, resolved.path),
    config_source: resolved.source
  };

  let rawValue = '';
  if (Array.isArray(value)) rawValue = value.join(',');
  else if (value === undefined || value === null) rawValue = '';
  else rawValue = String(value);

  output(result, rawValue, raw);
}

function cmdSpecResolveDir(cwd, args, raw) {
  const spec = args[0];
  if (!spec) fail('spec resolve-dir requires <spec-name>', raw);

  const rel = path.join('.spec-workflow', 'specs', spec);
  const abs = path.join(cwd, rel);
  const found = fs.existsSync(abs) && fs.statSync(abs).isDirectory();

  const result = { ok: true, spec, found, directory: found ? rel : null };
  output(result, found ? rel : '', raw);
}

function cmdRunsLatestUnfinished(cwd, args, raw) {
  const phaseDirArg = args[0];
  if (!phaseDirArg) fail('runs latest-unfinished requires <phase-dir>', raw);

  const phaseDir = path.isAbsolute(phaseDirArg) ? phaseDirArg : path.join(cwd, phaseDirArg);
  if (!fs.existsSync(phaseDir) || !fs.statSync(phaseDir).isDirectory()) {
    const result = { ok: true, phase_dir: phaseDirArg, found: false, reason: 'phase_dir_missing', run: null };
    output(result, '', raw);
    return;
  }

  const runs = listDirs(phaseDir)
    .map((r) => ({
      name: r.name,
      full: r.full,
      mtime: fs.statSync(r.full).mtimeMs
    }))
    .sort((a, b) => b.mtime - a.mtime);

  for (const run of runs) {
    const inspection = runDirInspection(run.full);
    if (inspection.unfinished) {
      const rel = path.relative(cwd, run.full);
      const result = {
        ok: true,
        phase_dir: phaseDirArg,
        found: true,
        run: rel,
        issues: inspection.issues,
        subagents: inspection.subagents
      };
      output(result, rel, raw);
      return;
    }
  }

  const result = { ok: true, phase_dir: phaseDirArg, found: false, reason: 'no_unfinished_run', run: null };
  output(result, '', raw);
}

function cmdHandoffValidate(cwd, args, raw) {
  const runDirArg = args[0];
  if (!runDirArg) fail('handoff validate requires <run-dir>', raw);

  const runDir = path.isAbsolute(runDirArg) ? runDirArg : path.join(cwd, runDirArg);
  if (!fs.existsSync(runDir) || !fs.statSync(runDir).isDirectory()) {
    fail(`run directory not found: ${runDirArg}`, raw);
  }

  const inspection = runDirInspection(runDir);
  const result = {
    ok: true,
    run_dir: runDirArg,
    valid: !inspection.unfinished,
    issues: inspection.issues,
    subagents: inspection.subagents
  };
  output(result, inspection.unfinished ? 'invalid' : 'valid', raw);
}

function cmdWaveResolveCurrent(cwd, args, raw) {
  const spec = args[0];
  if (!spec) fail('wave resolve-current requires <spec-name>', raw);

  const wavesDir = path.join(cwd, '.spec-workflow', 'specs', spec, '_agent-comms', 'waves');
  if (!fs.existsSync(wavesDir)) {
    const result = { ok: true, spec, found: false, wave: null, directory: null };
    output(result, 'none', raw);
    return;
  }

  const waves = listDirs(wavesDir)
    .map((d) => {
      const m = d.name.match(/^wave-(\d+)$/);
      return m ? { name: d.name, num: parseInt(m[1], 10), full: d.full } : null;
    })
    .filter(Boolean)
    .sort((a, b) => b.num - a.num);

  if (waves.length === 0) {
    const result = { ok: true, spec, found: false, wave: null, directory: null };
    output(result, 'none', raw);
    return;
  }

  const current = waves[0];
  const rel = path.relative(cwd, current.full);
  const result = { ok: true, spec, found: true, wave: current.name, directory: rel };
  output(result, current.name, raw);
}

function cmdSkillsEffectiveSet(cwd, args, raw) {
  const stage = (args[0] || '').toLowerCase();
  if (!['design', 'implementation'].includes(stage)) {
    fail('skills effective-set requires <design|implementation>', raw);
  }

  const resolved = resolveConfigPath(cwd);
  const cfg = readToml(resolved.path);

  const required = Array.isArray(cfg?.[`skills.${stage}`]?.required)
    ? [...cfg[`skills.${stage}`].required]
    : [];
  const optional = Array.isArray(cfg?.[`skills.${stage}`]?.optional)
    ? [...cfg[`skills.${stage}`].optional]
    : [];

  let enforceRequired = cfg?.[`skills.${stage}`]?.enforce_required;
  if (typeof enforceRequired !== 'boolean') {
    enforceRequired = String(cfg?.skills?.enforcement || '').toLowerCase() === 'strict';
  }

  const tddDefault = !!cfg?.execution?.tdd_default;
  if (stage === 'implementation' && tddDefault && !required.includes('test-driven-development')) {
    required.push('test-driven-development');
  }

  const result = {
    ok: true,
    stage,
    required,
    optional,
    enforce_required: enforceRequired,
    tdd_default: tddDefault,
    config_path: path.relative(cwd, resolved.path),
    config_source: resolved.source
  };

  output(result, required.join(','), raw);
}

function targetDocPath(specName, docType) {
  const normalized = String(docType || '').toLowerCase();
  if (normalized === 'requirements') return path.join('.spec-workflow', 'specs', specName, 'requirements.md');
  if (normalized === 'design') return path.join('.spec-workflow', 'specs', specName, 'design.md');
  if (normalized === 'tasks') return path.join('.spec-workflow', 'specs', specName, 'tasks.md');
  return null;
}

function cmdApprovalLocalFallbackId(cwd, args, raw) {
  const specName = args[0];
  const docType = args[1];
  if (!specName || !docType) fail('approval local-fallback-id requires <spec-name> <doc-type>', raw);

  const targetRel = targetDocPath(specName, docType);
  if (!targetRel) fail('doc-type must be one of: requirements|design|tasks', raw);

  const approvalsDir = path.join(cwd, '.spec-workflow', 'approvals', specName);
  if (!fs.existsSync(approvalsDir)) {
    const result = { ok: true, spec: specName, doc_type: docType, approval_id: null, source: null };
    output(result, '', raw);
    return;
  }

  const files = fs
    .readdirSync(approvalsDir)
    .filter((f) => /^approval_.*\.json$/i.test(f))
    .map((f) => ({
      name: f,
      full: path.join(approvalsDir, f),
      mtime: fs.statSync(path.join(approvalsDir, f)).mtimeMs
    }))
    .sort((a, b) => b.mtime - a.mtime);

  const targetNorm = targetRel.replace(/\\/g, '/');
  for (const file of files) {
    const json = readJson(file.full);
    if (!json) continue;

    const filePath = String(json.filePath || json.path || '').replace(/\\/g, '/');
    const matches =
      filePath.endsWith(targetNorm) ||
      filePath.endsWith('/' + path.basename(targetNorm)) ||
      filePath === targetNorm;

    if (!matches) continue;

    const approvalId = json.approvalId || json.id || json?.approval?.id || null;
    if (!approvalId) continue;

    const result = {
      ok: true,
      spec: specName,
      doc_type: docType,
      approval_id: String(approvalId),
      source: path.relative(cwd, file.full)
    };
    output(result, String(approvalId), raw);
    return;
  }

  const result = { ok: true, spec: specName, doc_type: docType, approval_id: null, source: null };
  output(result, '', raw);
}

function main() {
  const argv = process.argv.slice(2);
  const rawIdx = argv.indexOf('--raw');
  const raw = rawIdx !== -1;
  if (raw) argv.splice(rawIdx, 1);

  const command = argv[0];
  const subcommand = argv[1];
  const args = argv.slice(2);
  const cwd = process.cwd();

  if (!command) {
    fail('usage: spw-tools <command> <subcommand?> [args] [--raw]', raw);
  }

  if (command === 'config' && subcommand === 'get') return cmdConfigGet(cwd, args, raw);
  if (command === 'spec' && subcommand === 'resolve-dir') return cmdSpecResolveDir(cwd, args, raw);
  if (command === 'runs' && subcommand === 'latest-unfinished') return cmdRunsLatestUnfinished(cwd, args, raw);
  if (command === 'handoff' && subcommand === 'validate') return cmdHandoffValidate(cwd, args, raw);
  if (command === 'wave' && subcommand === 'resolve-current') return cmdWaveResolveCurrent(cwd, args, raw);
  if (command === 'skills' && subcommand === 'effective-set') return cmdSkillsEffectiveSet(cwd, args, raw);
  if (command === 'approval' && subcommand === 'local-fallback-id') return cmdApprovalLocalFallbackId(cwd, args, raw);

  fail(`unknown command: ${command} ${subcommand || ''}`.trim(), raw);
}

main();
