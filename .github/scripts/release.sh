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
  exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
tag="v${version}"
if [[ "$dryrun" != "true" && -n "$(git status --porcelain)" ]]; then
  echo "ERROR: Working tree is not clean."
  exit 1
fi
if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  echo "ERROR: Tag '${tag}' already exists."
  exit 1
fi

unreleased=$(awk '/^## \[Unreleased\]/{inside=1; next} /^## \[/{if(inside){exit}} inside{print}' CHANGELOG.md)
if [[ -z "${unreleased//[[:space:]]/}" ]] || ! grep -qE '^[[:space:]]*- ' <<<"$unreleased"; then
  echo "ERROR: CHANGELOG.md Unreleased section must contain at least one bullet."
  exit 1
fi

previous_tag=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' | sort -V | tail -n 1 || true)
manifest_file="$(mktemp)"
body_file="$(mktemp)"
updated_changelog="$(mktemp)"
cleanup() { rm -f "$manifest_file" "$body_file" "$updated_changelog"; }
trap cleanup EXIT

if [[ -n "$previous_tag" ]]; then
  ./.github/scripts/detect-impact.py "$version" "$previous_tag" > "$manifest_file"
  compare_ref="$previous_tag"
else
  ./.github/scripts/detect-impact.py "$version" > "$manifest_file"
  compare_ref="f6c183dd832bfadb0e2bf932e9c5b24f583ebdcd"
fi
artifacts=$(python3 -c 'import json,sys; print(", ".join(json.load(sys.stdin)["artifacts"]) or "none")' < "$manifest_file")
compare_link="**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/${compare_ref}...${tag}"

if [[ "$dryrun" == "true" ]]; then
  echo "[DRY RUN] Would prepare Starpac ${version} as ${tag}."
  echo "[DRY RUN] Previous release: ${previous_tag:-none (full baseline)}"
  echo "[DRY RUN] Binary artifacts: ${artifacts}"
  python3 -m json.tool "$manifest_file"
  echo "[DRY RUN] ${compare_link}"
  exit 0
fi

printf '%s\n' "$unreleased" > "$body_file"
awk -v version="$version" -v body_file="$body_file" -v link="$compare_link" '
  $0 == "## [Unreleased]" {
    print
    print ""
    print "## [" version "]"
    print ""
    while ((getline body_line < body_file) > 0) print body_line
    close(body_file)
    print ""
    print link
    skip=1
    next
  }
  skip && /^## \[/ {skip=0}
  !skip {print}
' CHANGELOG.md > "$updated_changelog"
mv "$updated_changelog" CHANGELOG.md

mkdir -p .starpac
cp "$manifest_file" .starpac/release.json
(
  cd website
  rm -rf node_modules/.cache
  npm ci
  npm run version-docs -- "$version"
  STARPAC_DOCS_LAST_VERSION="$version" npm run build
)
git add CHANGELOG.md .starpac/release.json website
git commit -m "chore(release): ${version}"
git tag -a "$tag" -m "Release ${version}"
echo "Release ${version} prepared locally with binary artifacts: ${artifacts}."
echo "Push main and ${tag} only after review."
