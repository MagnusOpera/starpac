---
title: Safety Model
---

d1pac classifies every operation:

- `safe`: creates, additive columns, and replaceable secondary objects.
- `migration`: table rebuilds that preserve every existing column.
- `destructive`: removed objects or columns.

Disabled operations remain visible as commented, `blocked-*` operations. A
destructive apply requires the project to allow drops or the invocation to
pass `--allow-drop` or `--force`.

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

## Operational boundary

d1pac manages schema only. It cannot infer data backfills, value conversions,
or seed records. Keep those as reviewed, explicit migrations. Large D1 data
changes may also need application-level batching to stay within platform query
limits.
