# Go Release

Release a new version of a Go module with semantic versioning.

## When to Apply

Use this skill when the user says:
- "release", "launch", "publish"
- "new version", "bump version"
- "tag version", "create release"

## Process

### 1. Check Current State

```bash
# Get latest tag
git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"

# Check for uncommitted changes
git status --porcelain

# Check if on main/master branch
git branch --show-current
```

### 2. Determine Version Bump

Ask the user which type of release:
- **patch** (v0.1.0 → v0.1.1): Bug fixes, small changes
- **minor** (v0.1.0 → v0.2.0): New features, backward compatible
- **major** (v0.1.0 → v1.0.0): Breaking changes

### 3. Create Release

IMPORTANT: Never delete or move existing tags. Always create a new tag and mark it as latest.

```bash
# Create annotated tag
git tag -a v{VERSION} -m "Release v{VERSION}"

# Push tag to remote
git push origin v{VERSION}

# Create GitHub Release and mark as latest
gh release create v{VERSION} --title "v{VERSION}" --generate-notes --latest
```

### 4. Verify Release

```bash
# Confirm tag was pushed
git ls-remote --tags origin | grep v{VERSION}

# Confirm release is marked as latest
gh release view v{VERSION} --json isLatest --jq '.isLatest'

# Show install command
echo "Install with: go install {MODULE_PATH}@v{VERSION}"
```

## Version Calculation

Given current version `vX.Y.Z`:
- **patch**: `vX.Y.(Z+1)`
- **minor**: `vX.(Y+1).0`
- **major**: `v(X+1).0.0`

## Pre-release Checklist

Before releasing, verify:
1. All changes are committed
2. Tests pass (if applicable)
3. Code builds successfully: `go build ./...`
4. On correct branch (main/master)

## Rules

- **Never delete existing tags** — tags are immutable references; create a new version instead
- **Never move tags** to point to a different commit
- **Always use `gh release create --latest`** to mark the new release as the latest
- If a previous release failed (e.g., CI broke), fix the issue, then create the next patch version

## Example Output

```
Current version: v0.1.0
Release type: minor
New version: v0.2.0

✓ Created tag v0.2.0
✓ Pushed to origin
✓ Marked as latest release

Install with:
  go install github.com/user/project@v0.2.0
  go install github.com/user/project@latest
```
