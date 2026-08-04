# Starpac

Starpac is the shared home of two independent desired-state database delivery tools:

- **pgpac** packages SQL into `.pgpkg` artifacts, compares them with PostgreSQL, and safely applies the resulting plan.
- **d1pac** packages SQL into `.d1pkg` artifacts and provides the same workflow for Cloudflare D1 with SQLite-native behavior.

The tools share stable infrastructure but keep separate CLIs, project files, package formats, engines, documentation, changelogs, versions, and releases. There is deliberately no generic database-engine abstraction.

Documentation: <https://magnusopera.github.io/starpac/>

## Build and test

```bash
make build
make test
make pgpac-build
make d1pac-build
make sample
```

The binaries remain `pgpac` and `d1pac`; their existing commands, flags, environment variables, XML roots, project extensions, and package extensions are unchanged.

## Repository layout

```text
cmd/                         public CLI entry points
internal/pac/                stable shared infrastructure
internal/postgres/           pgpac engine
internal/d1/                 d1pac engine
products/{pgpac,d1pac}/      changelogs, product notes, and test data
website/                     combined Starpac site and versioned docs
```

## Releases

Releases are independent and use scoped tags:

```bash
make release-prepare tool=pgpac version=X.Y.Z dryrun=true
make release-prepare tool=d1pac version=X.Y.Z dryrun=true
make release-build tool=pgpac version=X.Y.Z
make release-build tool=d1pac version=X.Y.Z
```

Published tags are `pgpac/vX.Y.Z` and `d1pac/vX.Y.Z`. Artifact names and embedded versions remain unscoped, such as `pgpac-X.Y.Z-linux-arm64.zip`.

## License

Starpac is licensed under the MIT License. See [LICENSE](LICENSE).
