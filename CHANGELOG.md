# Changelog

All notable Starpac changes are documented here. Starpac versions are global repository versions; a release contains binary artifacts only for tools affected since the previous release.

Global Starpac versioning begins at `0.6.0`. Earlier entries retain their product provenance: `pgpac` versions through `0.5.1` originated in the standalone pgpac repository, while `d1pac` versions through `0.0.3` originated in the standalone d1pac repository. Identical pre-Starpac version numbers belong to different products and are not global Starpac versions.

## [Unreleased]

## [0.7.1]


### d1pac

- Fixed no-op JSON plans to emit an empty `operations` array instead of
  `null`, preserving the documented machine-readable plan contract.

### Starpac

- Fixed signed macOS release uploads under GitHub Actions' Node 24 runtime by
  enabling the runner's system certificate authorities.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/v0.7.0...v0.7.1
## [0.7.0]


### Starpac

- Added per-product schema feature-support references covering native,
  migration, destructive, blocked, and unsupported behavior.

### pgpac

- Added row-preserving, explicitly gated `ALTER TABLE ... DROP COLUMN` plans and
  made blocked PostgreSQL plans report their unsupported status.
- Added native row-preserving plans for column type, default, and nullability
  changes and for primary-key additions, removals, and replacements.
- Added semantic diffing and native row-preserving alterations for named and
  unnamed foreign-key, unique, and check constraints.
- Enforced the pgpac project's `AllowCreate` and `AllowAlter` plan permissions
  with visible blocked operations, matching its existing drop authorization.
- Added proactive foreign-key dependency refresh around referenced key and
  column-type alterations instead of relying on PostgreSQL to reject them.
- Documented additive and destructive schema behavior, authorization, and the
  remaining table and constraint diffing limitations.

### d1pac

- Documented additive, table-rebuild, destructive, and blocked schema behavior
  with the applicable project and command-line safety gates.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/v0.6.3...v0.7.0
## [0.6.3]


### Starpac

- Published a cumulative product release index so consumers can discover each tool's latest artifact-bearing release.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/v0.6.2...v0.6.3
## [0.6.2]


### Starpac

- Made the `*pac` family naming explicit on the homepage and kept the latest released documentation free of archive warnings.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/v0.6.1...v0.6.2
## [0.6.1]


### Starpac

- Fixed Homebrew formula template resolution and serialized tap updates for multi-tool releases.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/v0.6.0...v0.6.1
## [0.6.0]


### Starpac

- Unified pgpac and d1pac under global Starpac versions with automatic partial artifact publishing.
- Consolidated the documentation into one version history and one website version selector.
- Simplified the homepage and aligned the pgpac and d1pac primary actions.
- Made release preparation print the exact command for pushing the release commit and annotated tag.

### pgpac

- Moved pgpac into the Starpac monorepo while preserving its CLI, project, package, and release contracts.
- Shared stable package, plan, rendering, project-resolution, safety, and CLI infrastructure with d1pac.

### d1pac

- Moved d1pac into the Starpac monorepo while preserving its CLI, project, package, and release contracts.
- Shared stable package, plan, rendering, project-resolution, safety, and CLI infrastructure with pgpac.

**Full Changelog**: https://github.com/MagnusOpera/starpac/compare/f6c183dd832bfadb0e2bf932e9c5b24f583ebdcd...v0.6.0
## Standalone pgpac release history

The following releases originated in the standalone `MagnusOpera/pgpac` repository. Their versions apply only to pgpac.

### pgpac [0.5.1]

- Added linked Magnus Opera branding to the documentation footer.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.5.0...0.5.1

### pgpac [0.5.0]

- Redesigned the website home page to present pgpac as desired-state schema management for PostgreSQL.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.4.0...0.5.0

### pgpac [0.4.0]

- Added the MIT license for repository distribution and reuse.
- Fixed macOS release signing to use the explicit arm64 code-signing identifier.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.3.0...0.4.0

### pgpac [0.3.0]

- Renamed the CLI, release artifacts, and install surfaces from `pgpackage` to `pgpac`.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.2.1...0.3.0

### pgpac [0.2.1]

- Pinned the macOS notarization release job to macOS 15 so published releases keep using the expected Xcode toolchain.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.2.0...0.2.1

### pgpac [0.2.0]

- Fixed release preparation to clear website build cache before reinstalling docs dependencies.
- Improved the website home page responsive layout on mobile screens.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.1.0...0.2.0

### pgpac [0.1.0]

- Added Linux x64 release archives alongside the existing macOS arm64 and Linux arm64 binaries.
- Reworked the website home page to better explain pgpac as desired-state schema management for PostgreSQL.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.0.2...0.1.0

### pgpac [0.0.2]

- Limited release artifacts to Linux arm64 and macOS arm64 only, removing Windows and macOS x64 targets.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.0.1...0.0.2

### pgpac [0.0.1]

- Added Docusaurus website scaffolding, docs versioning, and GitHub Pages release deployment.
- Added changelog-gated CI, draft GitHub release automation, and Homebrew tap update workflows.
- Added build-time CLI version reporting and cross-platform release packaging targets.
- Added macOS Developer ID signing and notarization to the release workflow for GitHub Release and Homebrew distribution.
- Fixed tag-release artifact uploads from the hidden `.out/` directory and removed the broken Windows arm64 release leg.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/b015550b6a7bbd28f781fb3662935e7752cd1532...0.0.1

## Standalone d1pac release history

The following releases originated in the standalone `MagnusOpera/d1pac` repository. Their versions apply only to d1pac.

### d1pac [0.0.3]

- Fixed remote D1 introspection using the restricted `sqlite_version()` function.

**Full Changelog**: https://github.com/MagnusOpera/d1pac/compare/0.0.2...0.0.3

### d1pac [0.0.2]

- Fixed explicit Cloudflare target flags being discarded by `plan` and `apply`.

**Full Changelog**: https://github.com/MagnusOpera/d1pac/compare/0.0.1...0.0.2

### d1pac [0.0.1]

- Added desired-state build, plan, and apply workflows for Cloudflare D1 schemas.
- Added transactional D1 API deployment with destructive-operation safeguards.
- Added package artifacts, documentation website, release automation, and Homebrew publishing.
- Fixed the homepage deployment example contrast in light theme.
- Fixed changelog validation in a newly initialized repository.
- Fixed release preparation on macOS when release notes span multiple lines.

**Full Changelog**: https://github.com/MagnusOpera/d1pac/compare/f281c6ba62fbeeac848dba8f1eb0f76991e431c2...0.0.1
