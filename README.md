# Starpac

Starpac is the shared home of two independent desired-state database delivery tools:

- **pgpac** packages SQL into `.pgpkg` artifacts, compares them with PostgreSQL, and safely applies the resulting plan.
- **d1pac** packages SQL into `.d1pkg` artifacts and provides the same workflow for Cloudflare D1 with SQLite-native behavior.

The tools share stable infrastructure but keep separate CLIs, project files, package formats, and engines. Starpac uses one repository version and publishes binary artifacts only for tools affected by a release. There is deliberately no generic database-engine abstraction.

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

Release preparation is intentionally automatic:

```bash
make release-prepare version=X.Y.Z dryrun=true
make release-prepare version=X.Y.Z
make release-build tool=pgpac version=X.Y.Z
make release-build tool=d1pac version=X.Y.Z
```

`release-prepare` compares the repository with the previous global `vX.Y.Z` tag. Product-specific changes select that product, shared or unknown runtime changes select both, and website-only changes select no binaries. It records the decision in `.starpac/release.json`, versions the combined documentation, commits, and creates the annotated global tag.

The first global release, `v0.6.0`, is always a full pgpac and d1pac baseline. Later releases may contain only one tool. Version gaps in a tool's artifacts are therefore expected.

## License

Starpac is licensed under the MIT License. See [LICENSE](LICENSE).
