#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
base_ref="${GITHUB_BASE_REF:-}"

if [[ -n "$base_ref" ]]; then
  git fetch --no-tags --depth=1 origin "$base_ref" >/dev/null 2>&1 || true
  changed_files="$(git diff --name-only "origin/${base_ref}...HEAD")"
elif [[ -n "$(git status --porcelain)" ]]; then
  changed_files="$({ git diff --name-only HEAD; git ls-files --others --exclude-standard; } | sort -u)"
elif git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
  changed_files="$(git diff --name-only HEAD~1...HEAD)"
else
  echo "No comparable changes; skipping changelog check."
  exit 0
fi

if [[ -z "$changed_files" ]]; then
  echo "No changed files detected; skipping changelog check."
  exit 0
fi
if ! grep -qx 'CHANGELOG.md' <<<"$changed_files"; then
  echo "ERROR: CHANGELOG.md must be updated."
  exit 1
fi

unreleased=$(awk '/^## \[Unreleased\]/{inside=1; next} /^## \[/{if(inside){exit}} inside{print}' CHANGELOG.md)
head_subject="$(git log -1 --pretty=%s 2>/dev/null || true)"
if [[ -z "${unreleased//[[:space:]]/}" ]]; then
  if [[ "$head_subject" =~ ^chore\(release\):\ [0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Changelog gate passed for release commit."
    exit 0
  fi
  echo "ERROR: CHANGELOG.md has an empty Unreleased section."
  exit 1
fi
if ! grep -qE '^[[:space:]]*- ' <<<"$unreleased"; then
  echo "ERROR: CHANGELOG.md Unreleased section must contain a bullet."
  exit 1
fi

echo "Changelog gate passed."
