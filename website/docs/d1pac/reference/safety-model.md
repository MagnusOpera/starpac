---
title: Safety Model
---

d1pac classifies every operation:

- `safe`: creates, supported additive columns, and replaceable secondary
  objects
- `migration`: table rebuilds that preserve every existing column
- `destructive`: removed objects or columns

Disabled operations remain visible as commented, `blocked-*` operations. A
destructive apply requires the project to allow drops or the invocation to
pass `--allow-drop` or `--force`.

## Additive and replacement operations

Creating a table, index, view, or trigger requires `AllowCreate="true"`.
Replacing an index, view, or trigger requires `AllowAlter="true"` and is
classified as safe.

A column is added with `ALTER TABLE ... ADD COLUMN` when it is not a primary-key
or hidden column and either is nullable or has a default. SQLite physically
appends the column, and default comparisons ignore column order so a later plan
does not report drift when the desired declaration placed it earlier. Pass
`--strict` to require declaration order; a non-trailing addition then uses the
rebuild path. Additive and migration operations are emitted as blocked SQL when
the corresponding project permission is disabled.

## Destructive operations

Dropping a table, index, view, or trigger requires drop authorization. Removing
a column also requires drop authorization and uses the table-rebuild path
below. The rebuild copies every remaining same-named column, but a rename is
interpreted as one removed column and one new column; it does not preserve the
renamed column's data.

## SQLite table rebuilds

Changes that SQLite cannot express as a direct alter use this sequence:

1. Create a temporary table with the desired definition.
2. Copy columns common to the old and new tables.
3. Drop the old table.
4. Rename the temporary table.
5. Recreate its explicit indexes and triggers.

Removing a column makes the rebuild destructive. Type and constraint changes
are migrations because SQLite conversion and new constraints can still reject
existing data. Always test these plans against representative data.

An automatic rebuild is blocked when another table references the target.
D1 keeps foreign keys enabled, and deferred checking does not suppress
`ON DELETE CASCADE`; dropping a referenced table could therefore delete child
rows. Handle such changes with an explicit, application-aware migration.

The referenced-table protection applies to rebuilds. A direct table drop is
still sent to D1 and can fail because of foreign-key enforcement. `--force`
does not override a blocked rebuild or make invalid data satisfy a new
constraint.

`StopOnDataLossRisk` records project intent but is not an additional
enforcement gate. With `UseTransaction="true"`, d1pac submits the schema change
as one batch. Without it, an error can leave earlier statements applied.

## Operational boundary

d1pac manages schema only. It cannot infer data backfills, value conversions,
or seed records. Keep those as reviewed, explicit migrations. Large D1 data
changes may also need application-level batching to stay within platform query
limits.
