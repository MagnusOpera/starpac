# d1pac

`d1pac` is a Go-first desired-state schema packaging tool for Cloudflare D1, in
the spirit of `sqlpackage` and `pgpac`.
It compiles SQLite SQL into a portable package, compares that model with a live
D1 database, and applies the resulting plan through the Cloudflare API.

Documentation website: <https://magnusopera.github.io/starpac/d1pac/>

## Commands

```bash
d1pac build --project testdata/sample/sample.d1pac --output out/

export CLOUDFLARE_ACCOUNT_ID="..."
export CLOUDFLARE_API_TOKEN="..."

d1pac plan --package out/SampleProject.d1pkg --database my-database
d1pac apply --package out/SampleProject.d1pkg --database my-database
d1pac --version
```

The API token must have D1 read access for `plan` and D1 write access for
`apply`. `--account-id` and `--api-token` can override the environment values.
The database argument accepts either a D1 database name or UUID.

## Project file

Projects use the `.d1pac` extension and describe the desired schema state:

```xml
<D1Pac ProjectVersion="1">
  <PropertyGroup>
    <PackageId>SampleProject</PackageId>
    <Version>0.1.0</Version>
  </PropertyGroup>

  <ItemGroup>
    <Table Include="Tables/**/*.sql" />
    <Index Include="Indexes/**/*.sql" />
    <View Include="Views/**/*.sql" />
    <Trigger Include="Triggers/**/*.sql" />
  </ItemGroup>

  <Target>
    <Comparison>
      <Ignore Type="table" Name="application_metadata" />
    </Comparison>
    <Plan AllowCreate="true" AllowAlter="true" AllowDrop="false" />
    <Apply UseTransaction="true" StopOnDataLossRisk="true" />
  </Target>
</D1Pac>
```

## Behavior

- Desired SQL is compiled offline in an isolated SQLite database.
- Live tables, columns, foreign keys, indexes, views, and triggers are
  introspected through the Cloudflare D1 REST API.
- Additive columns are emitted as trailing `ALTER TABLE ... ADD COLUMN`
  operations regardless of declared position. Column order is ignored by
  default; `--strict` enables exact order comparison.
- Other compatible table changes use SQLite's create/copy/drop/rename rebuild
  pattern while preserving common columns and recreating indexes and triggers.
- Removed columns and objects are destructive and remain blocked unless drops
  are explicitly allowed.
- Transactional apply uses a D1 batch and defers foreign-key validation until
  the batch completes.
- D1 internal objects, SQLite internal objects, and `d1_migrations` are ignored
  automatically.

Schema state deliberately excludes application data. Backfills and seed data
remain explicit migrations outside the desired-state package.

## Development

```bash
make build
make test
make sample
make release-build tool=d1pac version=0.0.0-dev
make website-install
make website-typecheck
make website-build
```

Global releases are prepared with `make release-prepare version=X.Y.Z`; d1pac artifacts are included automatically when d1pac or shared code changed.

## License

This repository is licensed under the MIT License. See [LICENSE](LICENSE).
