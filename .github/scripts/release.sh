#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "Usage: $0 <pgpac|d1pac> <version> [dryrun]"
  exit 2
fi

tool="$1"
version="$2"
dryrun="${3:-false}"
case "$tool" in pgpac|d1pac) ;; *) echo "ERROR: Unknown tool '$tool'."; exit 2;; esac
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "ERROR: Expected version X.Y.Z."
  exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
changelog="products/${tool}/CHANGELOG.md"
tag="${tool}/v${version}"
if [[ "$dryrun" != "true" && -n "$(git status --porcelain)" ]]; then
  echo "ERROR: Working tree is not clean."
  exit 1
fi
if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  echo "ERROR: Tag '${tag}' already exists."
  exit 1
fi

unreleased=$(awk '/^## \[Unreleased\]/{inside=1; next} /^## \[/{if(inside){exit}} inside{print}' "$changelog")
if [[ -z "${unreleased//[[:space:]]/}" ]] || ! grep -qE '^[[:space:]]*- ' <<<"$unreleased"; then
  echo "ERROR: ${changelog} Unreleased section must contain at least one bullet."
  exit 1
fi

previous_tag=$(git tag --list "${tool}/v*" | sed -n "s#^${tool}/v##p" | sort -V | tail -n 1 || true)
if [[ -n "$previous_tag" ]]; then
  previous_ref="${tool}/v${previous_tag}"
else
  case "$tool" in
    pgpac) previous_ref="f6c183dd832bfadb0e2bf932e9c5b24f583ebdcd" ;;
    d1pac) previous_ref="ddda54b8c9191be4efe67a45623043f6488c2690" ;;
  esac
fi
compare_link="**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/${previous_ref}...${tag}"

if [[ "$dryrun" == "true" ]]; then
  echo "[DRY RUN] Would release ${tool} ${version} as ${tag}."
  echo "[DRY RUN] Previous ref: ${previous_ref}"
  echo "[DRY RUN] ${compare_link}"
  exit 0
fi

body_file="$(mktemp)"
updated_changelog="$(mktemp)"
cleanup() { rm -f "$body_file" "$updated_changelog"; }
trap cleanup EXIT
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
' "$changelog" > "$updated_changelog"
mv "$updated_changelog" "$changelog"

case "$tool" in
  pgpac) docs_version_variable="PGPAC_DOCS_LAST_VERSION" ;;
  d1pac) docs_version_variable="D1PAC_DOCS_LAST_VERSION" ;;
esac
(cd website && npm ci && npm run version-docs -- "$tool" "$version" && env "${docs_version_variable}=$version" npm run build)
git add "$changelog" website
git commit -m "chore(release): ${tool} ${version}"
git tag -a "$tag" -m "Release ${tool} ${version}"
echo "Release prepared locally. Push main and ${tag} only after review."
