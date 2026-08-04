#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
base_ref="${GITHUB_BASE_REF:-}"
require_always="${REQUIRE_CHANGELOG_ALWAYS:-false}"
enforce_bullet="${ENFORCE_UNRELEASED_BULLET:-false}"

if [[ -n "$base_ref" ]]; then
  git fetch --no-tags --depth=1 origin "$base_ref" >/dev/null 2>&1 || true
  changed_files=()
  while IFS= read -r file; do changed_files+=("$file"); done < <(git diff --name-only "origin/${base_ref}...HEAD")
else
  changed_files=()
  while IFS= read -r file; do changed_files+=("$file"); done < <({ git diff --name-only HEAD; git ls-files --others --exclude-standard; } | sort -u)
  if [[ ${#changed_files[@]} -eq 0 ]] && git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
    while IFS= read -r file; do changed_files+=("$file"); done < <(git diff --name-only HEAD~1...HEAD)
  fi
fi

if [[ ${#changed_files[@]} -eq 0 ]]; then
  echo "No changed files detected; skipping changelog check."
  exit 0
fi

needs_pgpac=false
needs_d1pac=false
for file in "${changed_files[@]}"; do
  case "$file" in
    cmd/pgpac/*|internal/postgres/*|products/pgpac/*) needs_pgpac=true ;;
    cmd/d1pac/*|internal/d1/*|products/d1pac/*) needs_d1pac=true ;;
    internal/pac/*|website/*|.github/*|Makefile|go.mod|go.sum|README.md|AGENTS.md)
      needs_pgpac=true
      needs_d1pac=true
      ;;
  esac
done

if [[ "$require_always" == "true" ]]; then
  needs_pgpac=true
  needs_d1pac=true
fi

head_subject="$(git log -1 --pretty=%s 2>/dev/null || true)"
for tool in pgpac d1pac; do
  required_var="needs_${tool}"
  if [[ "${!required_var}" != "true" ]]; then
    continue
  fi
  changelog="products/${tool}/CHANGELOG.md"
  if ! printf '%s\n' "${changed_files[@]}" | grep -qx "$changelog"; then
    echo "ERROR: ${changelog} must be updated for these changes."
    exit 1
  fi
  unreleased=$(awk '/^## \[Unreleased\]/{inside=1; next} /^## \[/{if(inside){exit}} inside{print}' "$changelog")
  if [[ -z "${unreleased//[[:space:]]/}" ]]; then
    if [[ "$head_subject" =~ ^chore\(release\):\ (pgpac|d1pac)\ [0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      continue
    fi
    echo "ERROR: ${changelog} has an empty Unreleased section."
    exit 1
  fi
  if [[ "$enforce_bullet" == "true" ]] && ! grep -qE '^[[:space:]]*- ' <<<"$unreleased"; then
    echo "ERROR: ${changelog} Unreleased section must contain a bullet."
    exit 1
  fi
done

echo "Product-aware changelog gate passed."
