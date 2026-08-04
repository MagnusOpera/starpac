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

- pgpac changes update `products/pgpac/CHANGELOG.md`.
- d1pac changes update `products/d1pac/CHANGELOG.md`.
- shared code, repository infrastructure, or combined website changes update both changelogs.
- Keep a top-level `## [Unreleased]` section with short, user-facing bullets.
- Run `make verify-changelog`; CI enforces the same product-aware rules.

## Releases

Use `make release-prepare tool=<pgpac|d1pac> version=X.Y.Z`. Preview with `dryrun=true`. The command prepares only the selected changelog and docs, then creates an annotated `<tool>/vX.Y.Z` tag. Tag CI builds only that product; publishing its draft release signs it and updates only its Homebrew formula.
