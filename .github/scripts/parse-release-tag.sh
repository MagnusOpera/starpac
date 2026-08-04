#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <pgpac/vX.Y.Z|d1pac/vX.Y.Z>" >&2
  exit 2
fi

tag="$1"
if [[ ! "$tag" =~ ^(pgpac|d1pac)/v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  echo "ERROR: Invalid scoped release tag '$tag'." >&2
  exit 2
fi

printf '%s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}"
