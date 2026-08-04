# AGENTS

This file defines contributor expectations for building, testing, regression
safety, documentation, and release-note hygiene.

## Build, Test, and Non-Regression

Run these commands before opening or updating a PR:

- Build: `make build`
- Full test suite: `make test`
- Release build smoke: `make release-build version=0.0.0-dev`
- Website install: `make website-install`
- Website typecheck: `make website-typecheck`
- Website build: `make website-build`

Every feature and bug fix requires automated coverage. Add tests in the suite
matching the change surface. Never require live Cloudflare credentials in the
default test suite; use an HTTP test server for D1 API behavior.

## Code and Safety

- Keep TypeScript and Go checks passing without weakening compiler settings.
- Keep project parsing, packaging, diffing, and target access in separate
  packages.
- Treat Cloudflare API responses and remote schema data as untrusted input.
- Never log or package API tokens, account credentials, or environment values.
- Destructive schema changes must remain visible and explicitly authorized.
- Preserve D1 and SQLite internal objects during comparisons.
- Update reference documentation whenever observable behavior changes.

## Release Notes

- `CHANGELOG.md` must keep a top `## [Unreleased]` section.
- Every regular commit targeting `main` adds a short, user-facing bullet there.
- Release sections end with a `**Full Changelog**` compare link.
- Release commits named `chore(release): X.Y.Z` may leave `Unreleased` empty.
- Run `make verify-changelog` locally.

## Release Process

1. Run `make release-prepare version=X.Y.Z`.
2. Push the commit and tag with `git push origin main --follow-tags`.
3. Let tag CI create the draft release and artifacts.
4. Review and publish that draft release.
5. Published-release CI signs macOS and updates the Homebrew tap.
