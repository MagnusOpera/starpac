#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <pgpac|d1pac> <version>"
  exit 2
fi

tool="$1"
version="$2"
case "$tool" in pgpac|d1pac) ;; *) echo "ERROR: Unknown tool '$tool'."; exit 2;; esac
changelog="products/${tool}/CHANGELOG.md"
section_body=$(awk -v header="## [${version}]" '
  $0 == header {inside=1; next}
  /^## \[/ && inside {exit}
  inside {print}
' "$changelog")

if [[ -z "${section_body//[[:space:]]/}" ]]; then
  echo "ERROR: Missing or empty changelog section '## [${version}]' in ${changelog}."
  exit 1
fi
if ! grep -qE '^[[:space:]]*- ' <<<"$section_body"; then
  echo "ERROR: Release section must include at least one bullet."
  exit 1
fi
if ! grep -q '\*\*Full Changelog\*\*:' <<<"$section_body"; then
  echo "ERROR: Release section must include a Full Changelog link."
  exit 1
fi
printf '%s\n' "$section_body"
