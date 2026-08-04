# pgpac

`pgpac` is a Go-first PostgreSQL schema packaging tool in the spirit of `sqlpackage`, but designed around a standalone CLI and an XML project file.

Documentation website: <https://magnusopera.github.io/pgpac/>

## Commands

```bash
pgpac build --project testdata/sample/sample.pgpac --output out/
pgpac plan --package out/SampleProject.pgpkg --connection "postgres://..."
pgpac apply --package out/SampleProject.pgpkg --connection "postgres://..."
pgpac --version
```

## Project file

Projects use the `.pgpac` extension and describe the desired schema state.

```xml
<PgPac ProjectVersion="1">
  <PropertyGroup>
    <PackageId>SampleProject</PackageId>
    <Version>0.1.0</Version>
    <PostgresVersion>18</PostgresVersion>
    <DefaultSchema>app</DefaultSchema>
  </PropertyGroup>

  <ItemGroup>
    <Schema Include="Schemas/**/*.sql" />
    <Table Include="Tables/**/*.sql" />
    <View Include="Views/**/*.sql" />
    <Function Include="Functions/**/*.sql" />
    <Type Include="Types/**/*.sql" />
    <Extension Include="Extensions/**/*.sql" />
    <Security Include="Security/**/*.sql" />
  </ItemGroup>

  <Target>
    <OwnedSchemas>
      <Schema Name="app" />
    </OwnedSchemas>
    <Extensions>
      <Extension Name="pgcrypto" Version="1.3" />
    </Extensions>
    <Comparison MatchPrivileges="false" MatchOwners="false" MatchComments="true" />
    <Plan AllowCreate="true" AllowAlter="true" AllowDrop="false" />
    <Apply UseTransaction="true" LockTimeout="5s" StatementTimeout="10m" StopOnDataLossRisk="true" />
  </Target>
</PgPac>
```

## Status

This repository implements the v1 architecture and core end-to-end flow:

- XML project parsing and validation
- offline desired-state parsing using `libpg_query`
- `.pgpkg` package creation
- target PostgreSQL 17+ introspection
- typed plan generation
- apply execution with destructive-op safeguards

Integration tests that hit a live database are gated behind `PGPAC_TEST_DSN`.

Current compatibility policy:

- project `PostgresVersion` must be `17` or newer
- target databases must be PostgreSQL `17` or newer
- the tool does not require an exact project-version/target-version match

## Website and release workflow

- Docs live under `website/` and are built with Docusaurus.
- Releases are prepared with `make release-prepare version=X.Y.Z`.
- Tagged builds create draft GitHub releases with platform archives.
- Published releases update `magnusopera/homebrew-tap`; the website is deployed independently from `main` with the Publish Website workflow.

## License

This repository is licensed under the MIT License. See [LICENSE](/LICENSE).
