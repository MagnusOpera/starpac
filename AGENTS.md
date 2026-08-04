# Contributor guide

## Build and regression gates

Before committing, run:

- `make build`
- `make test`
- `make release-smoke` when distribution code changes
- `make website-install website-typecheck website-build` when the site or docs change
- `make verify-changelog`

Features and fixes require automated coverage. Keep live PostgreSQL and Cloudflare credentials out of the default suite; use opt-in integration tests or local test servers.

## Product boundaries

- Preserve both public CLI and archive contracts.
- Keep parsing, schema models, introspection, diffing, application, and transport in the product engine.
- Put code in `internal/pac` only when its contract is genuinely database-independent and stable for both products.
- Do not introduce a generic database-engine interface.
- Treat database and API responses as untrusted input and keep destructive operations explicitly gated.

## Changelogs

- Every change updates the root `CHANGELOG.md`.
- Keep a top-level `## [Unreleased]` section with short, user-facing bullets under the relevant Starpac, pgpac, or d1pac heading.
- Product changelogs under `products/` are standalone-history archives and are not updated for new releases.
- Run `make verify-changelog`; CI enforces the same rule.

## Releases

Use `make release-prepare version=X.Y.Z`. Preview with `dryrun=true`. The command detects affected binaries, records them in `.starpac/release.json`, prepares the changelog and global docs version, then creates an annotated `vX.Y.Z` tag. Tag CI trusts that manifest and publishes only its selected artifacts.
