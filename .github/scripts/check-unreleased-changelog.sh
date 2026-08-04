#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "$0")/../.." && pwd)
cd "$root_dir"

base_ref="${GITHUB_BASE_REF:-}"
require_always="${REQUIRE_CHANGELOG_ALWAYS:-false}"
enforce_bullet="${ENFORCE_UNRELEASED_BULLET:-false}"

if [[ -n "$base_ref" ]]; then
  git fetch --no-tags --depth=1 origin "$base_ref" >/dev/null 2>&1 || true
  mapfile -t changed_files < <(git diff --name-only "origin/${base_ref}...HEAD")
elif git rev-parse --verify HEAD >/dev/null 2>&1; then
  mapfile -t changed_files < <(git status --porcelain | awk '{print $2}')
else
  changed_files=()
  while IFS= read -r -d '' entry; do
    changed_files+=("${entry:3}")
  done < <(git status --porcelain=v1 -z --untracked-files=all)
fi

if [[ ${#changed_files[@]} -eq 0 ]]; then
  echo "No changed files detected; skipping changelog check."
  exit 0
fi

if ! printf '%s\n' "${changed_files[@]}" | grep -qx 'CHANGELOG.md' && [[ "$require_always" == "true" ]]; then
  echo "ERROR: CHANGELOG.md must be updated in every commit."
  exit 1
fi

unreleased_block=$(awk '
  /^## \[Unreleased\]/{inside=1; next}
  /^## \[/{if(inside){exit}}
  inside{print}
' CHANGELOG.md)

if [[ -z "${unreleased_block//$'\n'/}" ]]; then
  head_subject="$(git log -1 --pretty=%s 2>/dev/null || true)"
  if [[ "$head_subject" =~ ^chore\(release\):\ [0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Changelog gate passed."
    exit 0
  fi
  echo "ERROR: ## [Unreleased] section is empty."
  exit 1
fi

if [[ "$enforce_bullet" == "true" ]] && ! grep -qE '^- ' <<<"$unreleased_block"; then
  echo "ERROR: ## [Unreleased] must contain a bullet."
  exit 1
fi

echo "Changelog gate passed."
