#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
test_root="$(mktemp -d)"
cleanup() { rm -rf "$test_root"; }
trap cleanup EXIT

for tool in pgpac d1pac; do
  repository="$test_root/$tool"
  mkdir -p "$repository/.github/scripts" "$repository/products/$tool" "$repository/website" "$repository/bin"
  cp "$repo_root/.github/scripts/release.sh" "$repository/.github/scripts/release.sh"
  cp "$repo_root/website/scripts/version-docs.mjs" "$repository/website/version-docs.mjs"
  printf '{}\n' > "$repository/website/package.json"
  printf '# Changelog\n\n## [Unreleased]\n\n- Added the first release feature.\n- Fixed a multiline release path.\n' > "$repository/products/$tool/CHANGELOG.md"
  cat > "$repository/bin/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
  chmod +x "$repository/bin/npm" "$repository/.github/scripts/release.sh"
  git -C "$repository" init --quiet
  git -C "$repository" config user.name "Starpac release test"
  git -C "$repository" config user.email "release-test@starpac.invalid"
  git -C "$repository" remote add origin git@github.com:MagnusOpera/starpac.git
  git -C "$repository" add .
  git -C "$repository" commit --quiet -m "initial commit"
  (cd "$repository" && PATH="$repository/bin:$PATH" ./.github/scripts/release.sh "$tool" 9.8.7)
  grep -q '^## \[9.8.7\]$' "$repository/products/$tool/CHANGELOG.md"
  grep -q '^- Fixed a multiline release path\.$' "$repository/products/$tool/CHANGELOG.md"
  grep -q '^\*\*Full Changelog\*\*:' "$repository/products/$tool/CHANGELOG.md"
  git -C "$repository" rev-parse --verify --quiet "refs/tags/${tool}/v9.8.7" >/dev/null
  test "$(git -C "$repository" log -1 --format=%s)" = "chore(release): ${tool} 9.8.7"
done

test "$("$repo_root/.github/scripts/parse-release-tag.sh" pgpac/v1.2.3)" = "pgpac 1.2.3"
test "$("$repo_root/.github/scripts/parse-release-tag.sh" d1pac/v4.5.6)" = "d1pac 4.5.6"
if "$repo_root/.github/scripts/parse-release-tag.sh" 1.2.3 >/dev/null 2>&1; then
  echo "Unscoped release tag was unexpectedly accepted."
  exit 1
fi

mkdir -p "$test_root/homebrew-bin"
cat > "$test_root/homebrew-bin/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf 'darwin' > "${MOCK_TOOL}-${MOCK_VERSION}-darwin-arm64.zip"
printf 'linux-x64' > "${MOCK_TOOL}-${MOCK_VERSION}-linux-x64.zip"
printf 'linux-arm64' > "${MOCK_TOOL}-${MOCK_VERSION}-linux-arm64.zip"
EOF
chmod +x "$test_root/homebrew-bin/gh"
for tool in pgpac d1pac; do
  formula="$test_root/${tool}.rb"
  PATH="$test_root/homebrew-bin:$PATH" MOCK_TOOL="$tool" MOCK_VERSION="9.8.7" \
    "$repo_root/.github/scripts/generate-homebrew-tap" "$tool" 9.8.7 > "$formula"
  grep -q 'version "9.8.7"' "$formula"
  grep -q "MagnusOpera/starpac/releases/download/${tool}%2Fv9.8.7/${tool}-9.8.7-linux-arm64.zip" "$formula"
  grep -q "bin.install \"${tool}\"" "$formula"
done

echo "Both scoped release preparation paths passed."
