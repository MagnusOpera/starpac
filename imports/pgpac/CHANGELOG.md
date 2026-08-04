# Changelog

All notable changes to pgpac are documented in this file.

## [Unreleased]

- Fixed the mobile website navigation so its menu remains fully visible.

## [0.5.1]


- Added linked Magnus Opera branding to the documentation footer.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.5.0...0.5.1

## [0.5.0]


- Redesigned the website home page to present pgpac as desired-state schema management for PostgreSQL.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.4.0...0.5.0

## [0.4.0]


- Added the MIT license for repository distribution and reuse.
- Fixed macOS release signing to use the explicit arm64 code-signing identifier.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.3.0...0.4.0

## [0.3.0]


- Renamed the CLI, release artifacts, and install surfaces from `pgpackage` to `pgpac`.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.2.1...0.3.0

## [0.2.1]


- Pinned the macOS notarization release job to macOS 15 so published releases keep using the expected Xcode toolchain.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.2.0...0.2.1

## [0.2.0]


- Fixed release preparation to clear website build cache before reinstalling docs dependencies.
- Improved the website home page responsive layout on mobile screens.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.1.0...0.2.0

## [0.1.0]


- Added Linux x64 release archives alongside the existing macOS arm64 and Linux arm64 binaries.
- Reworked the website home page to better explain pgpac as desired-state schema management for PostgreSQL.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.0.2...0.1.0

## [0.0.2]


- Limited release artifacts to Linux arm64 and macOS arm64 only, removing Windows and macOS x64 targets.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/0.0.1...0.0.2

## [0.0.1]


- Added Docusaurus website scaffolding, docs versioning, and GitHub Pages release deployment.
- Added changelog-gated CI, draft GitHub release automation, and Homebrew tap update workflows.
- Added build-time CLI version reporting and cross-platform release packaging targets.
- Added macOS Developer ID signing and notarization to the release workflow for GitHub Release and Homebrew distribution.
- Fixed tag-release artifact uploads from the hidden `.out/` directory and removed the broken Windows arm64 release leg.

**Full Changelog**: https://github.com/MagnusOpera/pgpac/compare/b015550b6a7bbd28f781fb3662935e7752cd1532...0.0.1
