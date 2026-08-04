#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
test_root="$(mktemp -d)"
repository="$test_root/repository"
cleanup() { rm -rf "$test_root"; }
trap cleanup EXIT

mkdir -p "$repository/.github/scripts" "$repository/website" "$repository/bin"
cp "$repo_root/.github/scripts/release.sh" "$repository/.github/scripts/release.sh"
cp "$repo_root/.github/scripts/detect-impact.py" "$repository/.github/scripts/detect-impact.py"
printf '{}\n' > "$repository/website/package.json"
printf '# Changelog\n\n## [Unreleased]\n\n- Established the global release baseline.\n' > "$repository/CHANGELOG.md"
cat > "$repository/bin/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
chmod +x "$repository/bin/npm" "$repository/.github/scripts/release.sh" "$repository/.github/scripts/detect-impact.py"

git -C "$repository" init --quiet
git -C "$repository" config user.name "Starpac release test"
git -C "$repository" config user.email "release-test@starpac.invalid"
git -C "$repository" remote add origin git@github.com:MagnusOpera/starpac.git
git -C "$repository" add .
git -C "$repository" commit --quiet -m "initial commit"

add_change() {
  local path="$1"
  local bullet="$2"
  mkdir -p "$repository/$(dirname "$path")"
  printf 'change\n' >> "$repository/$path"
  python3 - "$repository/CHANGELOG.md" "$bullet" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
bullet = sys.argv[2]
content = path.read_text()
content = content.replace("## [Unreleased]\n", f"## [Unreleased]\n\n- {bullet}\n", 1)
path.write_text(content)
PY
  git -C "$repository" add .
  git -C "$repository" commit --quiet -m "test change"
}

assert_artifacts() {
  local expected="$1"
  python3 - "$repository/.starpac/release.json" "$expected" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as source:
    actual = json.load(source)["artifacts"]
expected = [item for item in sys.argv[2].split(",") if item]
if actual != expected:
    raise SystemExit(f"artifacts {actual!r}, expected {expected!r}")
PY
}

(
  cd "$repository"
  PATH="$repository/bin:$PATH" ./.github/scripts/release.sh 0.6.0
)
assert_artifacts "pgpac,d1pac"
git -C "$repository" rev-parse --verify --quiet refs/tags/v0.6.0 >/dev/null
test "$(git -C "$repository" log -1 --format=%s)" = "chore(release): 0.6.0"

add_change "internal/d1/introspect/change.go" "Changed only d1pac."
(
  cd "$repository"
  PATH="$repository/bin:$PATH" ./.github/scripts/release.sh 0.6.1
)
assert_artifacts "d1pac"

add_change "internal/pac/render/change.go" "Changed shared runtime behavior."
(
  cd "$repository"
  PATH="$repository/bin:$PATH" ./.github/scripts/release.sh 0.6.2
)
assert_artifacts "pgpac,d1pac"

add_change "website/docs/change.md" "Changed only documentation."
(
  cd "$repository"
  PATH="$repository/bin:$PATH" ./.github/scripts/release.sh 0.6.3
)
assert_artifacts ""

grep -q '^## \[0.6.3\]$' "$repository/CHANGELOG.md"
grep -q '^\*\*Full Changelog\*\*:' "$repository/CHANGELOG.md"

test "$("$repo_root/.github/scripts/parse-release-tag.sh" v1.2.3)" = "1.2.3"
if "$repo_root/.github/scripts/parse-release-tag.sh" pgpac/v1.2.3 >/dev/null 2>&1; then
  echo "Scoped product tag was unexpectedly accepted."
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
  PATH="$test_root/homebrew-bin:$PATH" MOCK_TOOL="$tool" MOCK_VERSION="0.6.1" \
    "$repo_root/.github/scripts/generate-homebrew-tap" "$tool" 0.6.1 > "$formula"
  grep -q 'version "0.6.1"' "$formula"
  grep -q "MagnusOpera/starpac/releases/download/v0.6.1/${tool}-0.6.1-linux-arm64.zip" "$formula"
  grep -q "bin.install \"${tool}\"" "$formula"
done

echo "Global full, partial, shared, and documentation-only release paths passed."
