# AGENTS

This file defines contributor expectations for building, testing, regression safety, documentation, and release-note hygiene.

## Build, Test, and Non-Regression

Use these commands before opening or updating a PR:

- Build: `make build`
- Full test suite: `make test`
- Release build smoke: `make release-build version=0.0.0-dev`
- Website install: `make website-install`
- Website typecheck: `make website-typecheck`
- Website build: `make website-build`

Equivalent direct commands:

- `go build ./cmd/pgpac`
- `go test ./...`

## Test Quality Policy

- Every new feature must include automated test coverage.
- Every bug fix must include a regression test reproducing the prior failure mode.
- Add tests in the suite matching the change surface:
  - CLI and wiring -> `cmd/pgpac`
  - Package/model/project parsing -> matching `internal/...` package tests
  - SQL diff/apply behavior -> `internal/diff` and `internal/apply`
- If release/build/distribution behavior changes, run `make release-build version=0.0.0-dev`.
- If website docs or release notes change, run `make website-typecheck` and `make website-build`.

## Release Notes (Unreleased)

- `CHANGELOG.md` must keep a top `## [Unreleased]` section.
- Each new feature/fix entry must be a short, single-line bullet.
- Write entries in user-facing terms, not implementation detail.
- At release time, move unreleased entries to the versioned section and reset `Unreleased`.
- Each released version section should end with a compare link:
  - `**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/<previous-tag>...<new-tag>`
- When publishing the GitHub release, include that same compare link in the release notes body.

## Commit Gate (Hard Requirement)

- Every commit that targets `main` must update `CHANGELOG.md`.
- Required format for regular commits:
  - add at least one short, single-line bullet under `## [Unreleased]`.
- Scope is strict, including docs/process/build/release work.
- Exception: release commits (`chore(release): X.Y.Z`) may leave `## [Unreleased]` empty.
- Local preflight command:
  - `make verify-changelog`
- CI enforces this on both PRs and direct pushes to `main`.

## Release Process (Tags and GitHub Draft)

Follow this exact sequence for every release:

1. Run `make release-prepare version=X.Y.Z`.
   - Optional preview mode: `make release-prepare version=X.Y.Z dryrun=true`
2. Push commit and tag together: `git push origin main --follow-tags`.
3. Wait for CI to create the GitHub Release as draft from the tag workflow.
4. Confirm the draft notes are sourced from `CHANGELOG.md` `## [X.Y.Z]` including compare link.
5. Publish that existing draft release.

Rules:

- Tag-triggered CI is the source of truth for release artifacts and draft release creation.
- Do not bypass the draft step.
- Tag workflow must fail if `CHANGELOG.md` has no non-empty `## [X.Y.Z]` section with bullets and compare link.
- Release notes must match the `CHANGELOG.md` version section and keep the compare link.
- `make release-prepare` supports `X.Y.Z` only.

## Documentation Maintenance

- Any behavioral change in project parsing, packaging, planning, apply semantics, or safety rules must update the corresponding docs.
- Any release/distribution/workflow change must update `README.md`, website docs, and this file when relevant.
- If docs are reorganized or files are moved, all internal links must be updated in the same PR.

## PR Checklist

- Build passes.
- Relevant tests pass.
- New behavior is test covered.
- Release build still succeeds when relevant.
- Website typecheck/build passes when docs or website code change.
- `CHANGELOG.md` `## [Unreleased]` has concise one-line entries for the change.
- Relevant documentation has been updated.

## Direct To Main Policy

- Committing directly to `main` follows the same quality bar as a PR.
- All checklist items above still apply.
- Documentation and release notes must be updated in the same change set.
- Direct-to-main commits are blocked by the changelog gate if `CHANGELOG.md` is not updated.
