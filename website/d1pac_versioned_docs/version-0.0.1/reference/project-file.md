---
title: Project File
---

Projects use the `.d1pac` extension:

```xml
<D1Pac ProjectVersion="1">
  <PropertyGroup>
    <PackageId>ApplicationDatabase</PackageId>
    <Version>1.0.0</Version>
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

## Properties

- `PackageId`: artifact name when the output is a directory.
- `Version`: application-controlled schema package version.
- `ProjectVersion`: d1pac project format version.

## Items

- `Table`, `Index`, `View`, and `Trigger` accept doublestar glob patterns.
- Files must have a `.sql` extension.
- Tables are compiled before indexes, views, and triggers, independently of
  XML declaration order.

Each file can contain multiple statements. Desired inputs should contain DDL,
not data changes.

## Target

- `Comparison/Ignore` excludes a named `table`, `index`, `view`, `trigger`, or
  all types with `Type="*"`.
- `AllowCreate`, `AllowAlter`, and `AllowDrop` control which plan operations
  are executable.
- `UseTransaction` applies the plan as one D1 query batch when enabled.
- `StopOnDataLossRisk` records the project's safety intent; destructive
  operations always require explicit drop authorization.

SQLite names beginning with `sqlite_` or `_cf_`, plus `d1_migrations`, are
always ignored.
