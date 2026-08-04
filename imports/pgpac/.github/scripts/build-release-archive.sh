#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <version> [output-dir]"
  exit 2
fi

version="$1"
out_dir="${2:-.out}"

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

goos="$(go env GOOS)"
goarch="$(go env GOARCH)"

case "$goarch" in
  amd64) arch_label="x64" ;;
  arm64) arch_label="arm64" ;;
  *)
    echo "ERROR: Unsupported GOARCH '$goarch'."
    exit 1
    ;;
esac

binary_name="pgpac"
if [[ "$goos" == "windows" ]]; then
  binary_name="pgpac.exe"
fi

commit="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
build_date="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
ldflags="-X main.version=${version} -X main.commit=${commit} -X main.buildDate=${build_date}"

platform_dir="${out_dir}/${goos}/${arch_label}"
archive_path="${out_dir}/pgpac-${version}-${goos}-${arch_label}.zip"

rm -rf "$platform_dir"
mkdir -p "$platform_dir"

go build -ldflags "$ldflags" -o "${platform_dir}/${binary_name}" ./cmd/pgpac

python3 - <<'PY' "$platform_dir" "$binary_name" "$archive_path"
import sys
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

platform_dir = Path(sys.argv[1])
binary_name = sys.argv[2]
archive_path = Path(sys.argv[3])
archive_path.parent.mkdir(parents=True, exist_ok=True)

with ZipFile(archive_path, "w", compression=ZIP_DEFLATED) as zf:
    zf.write(platform_dir / binary_name, arcname=binary_name)
PY

echo "Created ${archive_path}"
