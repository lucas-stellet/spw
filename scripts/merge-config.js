#!/usr/bin/env node
"use strict";

// merge-config.js <template> <user-config> <output>
//
// Merges a user's spw-config.toml with a new template version.
// Template structure (sections, comments, key order) is preserved.
// User values override template defaults for matching section.key pairs.
// New template keys appear with their defaults; deprecated user keys are dropped.

const fs = require("fs");

const [templatePath, userPath, outputPath] = process.argv.slice(2);

if (!templatePath || !userPath || !outputPath) {
  process.stderr.write(
    "Usage: node merge-config.js <template> <user-config> <output>\n"
  );
  process.exit(1);
}

// If user config doesn't exist, output = template verbatim
if (!fs.existsSync(userPath)) {
  fs.copyFileSync(templatePath, outputPath);
  process.exit(0);
}

// ---------------------------------------------------------------------------
// Detect whether a raw value after '=' starts a multi-line array
// (contains '[' but no closing ']' on the same logical segment)
// ---------------------------------------------------------------------------
function isMultiLineArrayStart(rawValue) {
  const stripped = rawValue.replace(/#.*$/, "").trim();
  return stripped.includes("[") && !stripped.includes("]");
}

// ---------------------------------------------------------------------------
// Parse config into Map<"section.key", rawValue>
// rawValue is everything after '=' (single or multi-line), preserving format.
// ---------------------------------------------------------------------------
function parseConfig(text) {
  const map = new Map();
  const lines = text.split("\n");
  let currentSection = "";
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    const sectionMatch = line.match(/^\s*\[([^\]]+)\]\s*$/);
    if (sectionMatch) {
      currentSection = sectionMatch[1];
      i++;
      continue;
    }

    if (/^\s*#/.test(line) || /^\s*$/.test(line)) {
      i++;
      continue;
    }

    const kvMatch = line.match(/^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=(.*)/);
    if (kvMatch) {
      const key = kvMatch[1];
      let rawValue = kvMatch[2];

      if (isMultiLineArrayStart(rawValue)) {
        i++;
        while (i < lines.length) {
          rawValue += "\n" + lines[i];
          if (lines[i].includes("]")) break;
          i++;
        }
      }

      const fullKey = currentSection ? `${currentSection}.${key}` : key;
      map.set(fullKey, rawValue);
      i++;
      continue;
    }

    i++;
  }

  return map;
}

// ---------------------------------------------------------------------------
// Walk template line-by-line, substituting user values for matching keys.
// Template structure (sections, comments, blank lines, key order) is preserved.
// ---------------------------------------------------------------------------
function mergeWithTemplate(templateText, userMap) {
  const templateLines = templateText.split("\n");
  const output = [];
  const templateKeys = new Set();
  let currentSection = "";
  let i = 0;

  while (i < templateLines.length) {
    const line = templateLines[i];

    // Section header — pass through
    const sectionMatch = line.match(/^\s*\[([^\]]+)\]\s*$/);
    if (sectionMatch) {
      currentSection = sectionMatch[1];
      output.push(line);
      i++;
      continue;
    }

    // Comments and blank lines — pass through from template
    if (/^\s*#/.test(line) || /^\s*$/.test(line)) {
      output.push(line);
      i++;
      continue;
    }

    // Key = value
    const kvMatch = line.match(/^(\s*([A-Za-z_][A-Za-z0-9_]*)\s*=)(.*)/);
    if (kvMatch) {
      const prefix = kvMatch[1]; // e.g. "key ="
      const key = kvMatch[2];
      const templateRawValue = kvMatch[3];
      const fullKey = currentSection ? `${currentSection}.${key}` : key;
      templateKeys.add(fullKey);

      // Collect all template lines for this value (handles multi-line arrays)
      const templateValueLines = [line];
      if (isMultiLineArrayStart(templateRawValue)) {
        i++;
        while (i < templateLines.length) {
          templateValueLines.push(templateLines[i]);
          if (templateLines[i].includes("]")) {
            i++;
            break;
          }
          i++;
        }
      } else {
        i++;
      }

      if (userMap.has(fullKey)) {
        // User value replaces template value (user's raw string after '=')
        const userRaw = userMap.get(fullKey);
        output.push(`${prefix}${userRaw}`);
      } else {
        // Keep template default lines verbatim
        for (const tl of templateValueLines) {
          output.push(tl);
        }
      }

      continue;
    }

    // Anything else — pass through
    output.push(line);
    i++;
  }

  // Report deprecated keys (in user, not in template)
  for (const [userKey] of userMap) {
    if (!templateKeys.has(userKey)) {
      process.stderr.write(
        `[merge-config] Deprecated key dropped: ${userKey}\n`
      );
    }
  }

  return output.join("\n");
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------
const templateText = fs.readFileSync(templatePath, "utf8");
const userText = fs.readFileSync(userPath, "utf8");

const userMap = parseConfig(userText);
const merged = mergeWithTemplate(templateText, userMap);

fs.writeFileSync(outputPath, merged);
