#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <version> [dryrun]"
  exit 2
fi

version="$1"
dryrun="${2:-false}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "ERROR: Expected version X.Y.Z."
  exit 1
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
if [[ "$dryrun" != "true" && -n "$(git status --porcelain)" ]]; then
  echo "ERROR: Working tree is not clean."
  exit 1
fi
if git rev-parse -q --verify "refs/tags/${version}" >/dev/null; then
  echo "ERROR: Tag '${version}' already exists."
  exit 1
fi

unreleased=$(awk '
  /^## \[Unreleased\]/{inside=1; next}
  /^## \[/{if(inside){exit}}
  inside{print}
' CHANGELOG.md)
if [[ -z "${unreleased//[[:space:]]/}" ]] || ! grep -qE '^[[:space:]]*- ' <<<"$unreleased"; then
  echo "ERROR: Unreleased must contain at least one bullet."
  exit 1
fi

previous_tag=$(git tag --list | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true)
if [[ -z "$previous_tag" ]]; then
  previous_tag=$(git rev-list --max-parents=0 HEAD | tail -n 1)
fi
compare_link="**Full Changelog**: https://github.com/MagnusOpera/d1pac/compare/${previous_tag}...${version}"

if [[ "$dryrun" == "true" ]]; then
  echo "[DRY RUN] Would release ${version}."
  echo "[DRY RUN] ${compare_link}"
  exit 0
fi

temporary=$(mktemp)
awk -v version="$version" -v body="$unreleased" -v link="$compare_link" '
  $0 == "## [Unreleased]" {
    print
    print ""
    print "## [" version "]"
    print ""
    print body
    print ""
    print link
    skip=1
    next
  }
  skip && /^## \[/ {skip=0}
  !skip {print}
' CHANGELOG.md > "$temporary"
mv "$temporary" CHANGELOG.md

(cd website && npm ci && npm run version-docs -- "$version" && D1PAC_DOCS_LAST_VERSION="$version" npm run build)
git add CHANGELOG.md website
git commit -m "chore(release): ${version}"
git tag -a "$version" -m "Release ${version}"
echo "Release prepared. Push with: git push origin main --follow-tags"
