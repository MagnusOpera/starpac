#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
test_root=$(mktemp -d)
cleanup() {
  rm -rf "$test_root"
}
trap cleanup EXIT

mkdir -p "$test_root/bin" "$test_root/repository/website"
cp "$repo_root/.github/scripts/release.sh" "$test_root/release.sh"

cat > "$test_root/repository/CHANGELOG.md" <<'EOF'
# Changelog

## [Unreleased]

- Added the first release feature.
- Fixed a release bug across multiple lines.
EOF

cat > "$test_root/repository/website/package.json" <<'EOF'
{}
EOF

cat > "$test_root/bin/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
chmod +x "$test_root/bin/npm"

git -C "$test_root/repository" init --quiet
git -C "$test_root/repository" config user.name "d1pac release test"
git -C "$test_root/repository" config user.email "release-test@d1pac.invalid"
git -C "$test_root/repository" add CHANGELOG.md website/package.json
git -C "$test_root/repository" commit --quiet -m "initial commit"

(
  cd "$test_root/repository"
  PATH="$test_root/bin:$PATH" "$test_root/release.sh" 0.0.1
)

grep -q '^## \[0.0.1\]$' "$test_root/repository/CHANGELOG.md"
grep -q '^- Added the first release feature\.$' "$test_root/repository/CHANGELOG.md"
grep -q '^- Fixed a release bug across multiple lines\.$' "$test_root/repository/CHANGELOG.md"
grep -q '^\*\*Full Changelog\*\*:' "$test_root/repository/CHANGELOG.md"
git -C "$test_root/repository" rev-parse --verify --quiet 'refs/tags/0.0.1' >/dev/null
test "$(git -C "$test_root/repository" log -1 --format=%s)" = 'chore(release): 0.0.1'
