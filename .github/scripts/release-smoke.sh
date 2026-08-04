#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"
smoke_dir="$(mktemp -d)"
cleanup() { rm -rf "$smoke_dir"; }
trap cleanup EXIT

native_target="$(go env GOOS) $(go env GOARCH)"
for tool in pgpac d1pac; do
  targets=("darwin arm64" "linux amd64" "linux arm64")
  if [[ "$tool" == "pgpac" ]]; then
    # pg_query uses CGo. CI builds pgpac on native macOS and Linux runners;
    # a local cross-build cannot link it without a platform C toolchain.
    targets=("$native_target")
  fi
  for target in "${targets[@]}"; do
    set -- $target
    GOOS="$1" GOARCH="$2" ./.github/scripts/build-release-archive.sh "$tool" 0.0.0-smoke "$smoke_dir"
  done
done

native_arch="$(go env GOARCH)"
if [[ "$native_arch" == "amd64" ]]; then native_arch="x64"; fi
for expected in \
  "pgpac-0.0.0-smoke-$(go env GOOS)-${native_arch}.zip" \
  d1pac-0.0.0-smoke-darwin-arm64.zip d1pac-0.0.0-smoke-linux-x64.zip d1pac-0.0.0-smoke-linux-arm64.zip; do
  test -s "$smoke_dir/$expected"
done

echo "d1pac passed cross-platform archive smoke builds; pgpac passed its native CGo archive build."
