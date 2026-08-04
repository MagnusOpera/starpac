#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <vX.Y.Z>" >&2
  exit 2
fi

tag="$1"
if [[ ! "$tag" =~ ^v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  echo "ERROR: Invalid Starpac release tag '$tag'." >&2
  exit 2
fi

printf '%s\n' "${BASH_REMATCH[1]}"
