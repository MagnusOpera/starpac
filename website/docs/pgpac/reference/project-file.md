---
title: Project File
---

Projects use the `.pgpac` extension and declare the desired schema state.

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

## PropertyGroup

- `PackageId`: package identifier used for the output file name and manifest.
- `Version`: package version embedded in the generated `.pgpkg`.
- `PostgresVersion`: minimum supported PostgreSQL major version for the project and target.
- `DefaultSchema`: default schema used by the parser when schema names are omitted.

## ItemGroup

Each item type points to SQL files using glob patterns relative to the project file directory.

- `Schema`
- `Table`
- `View`
- `Function`
- `Type`
- `Extension`
- `Security`

## Target

- `OwnedSchemas`: schemas treated as owned by the package.
- `Extensions`: expected extensions and versions.
- `Comparison`: toggles for privileges, owners, and comments.
- `Plan`: `AllowDrop` authorizes destructive operations from the project.
  `AllowCreate` and `AllowAlter` are reserved project settings and are not
  currently enforced by pgpac.
- `Apply`: transaction and timeout behavior for execution.

See the [safety model](./safety-model.md) for the exact additive and destructive
operation behavior and current table-change limitations.
