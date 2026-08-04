#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <version>"
  exit 2
fi

version="$1"
section_body=$(awk -v header="## [${version}]" '
  $0 == header {inside=1; next}
  /^## / && inside {exit}
  inside {print}
' CHANGELOG.md)

if [[ -z "${section_body//[[:space:]]/}" ]]; then
  echo "ERROR: Missing or empty changelog section '## [${version}]'."
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
