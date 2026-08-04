# Changelog

All notable Starpac changes are documented here. Versions are global repository versions; a release contains binary artifacts only for tools affected since the previous release.

The standalone release histories remain in `products/pgpac/CHANGELOG.md` and `products/d1pac/CHANGELOG.md`.

## [Unreleased]

### Starpac

- Unified pgpac and d1pac under global Starpac versions with automatic partial artifact publishing.
- Consolidated the documentation into one version history and one website version selector.

### pgpac

- Moved pgpac into the Starpac monorepo while preserving its CLI, project, package, and release contracts.
- Shared stable package, plan, rendering, project-resolution, safety, and CLI infrastructure with d1pac.

### d1pac

- Moved d1pac into the Starpac monorepo while preserving its CLI, project, package, and release contracts.
- Shared stable package, plan, rendering, project-resolution, safety, and CLI infrastructure with pgpac.
