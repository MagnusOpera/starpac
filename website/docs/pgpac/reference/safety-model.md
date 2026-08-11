---
title: Safety Model
---

pgpac classifies every planned operation as `safe`, `migration`, or
`destructive`. Operations that require drop authorization remain visible as
commented `blocked-*` operations when authorization is absent.

## Additive and replacement operations

The following operations are classified as safe and do not require drop
authorization:

- creating a missing schema object
- adding a column when all existing columns and the primary key are unchanged
- replacing a view or routine with `CREATE OR REPLACE`
- setting or clearing a supported comment

Adding a column uses `ALTER TABLE ... ADD COLUMN` and preserves existing rows.
PostgreSQL validates any new default and `NOT NULL` requirement while applying
the statement.

Changing or removing a column default and removing `NOT NULL` are also safe
native alterations because they do not rewrite existing values.

## Row-preserving migrations

The following changes use native `ALTER TABLE` operations and preserve rows:

- changing a column type with an explicit `USING column::new_type` conversion
- adding `NOT NULL`
- adding, removing, renaming, or changing the columns of a primary key
- adding, removing, renaming, or changing `FOREIGN KEY`, `UNIQUE`, and `CHECK`
  constraints

These operations are classified as migrations. They do not require drop
authorization, but PostgreSQL can reject them when existing values cannot be
cast, contain nulls, are not unique, fail a check, lack a referenced key, or
violate another dependency. pgpac does not infer a custom data-conversion
expression.

Named constraints are matched by name and semantic definition. An unnamed
constraint in the desired SQL matches an equivalent target constraint without
depending on PostgreSQL's generated name. Constraint replacement drops the old
named constraint before adding and validating the desired definition; it does
not recreate the table.

## Destructive operations

The following operations require drop authorization:

- dropping a schema, extension, table, index, view, routine, enum, domain, or
  sequence that is absent from the desired model
- dropping a column that is absent from the desired table
- recreating an object that cannot be altered or replaced in place

A column drop uses `ALTER TABLE ... DROP COLUMN` without `CASCADE`. It preserves
the table and its remaining rows. If another object depends on the column,
PostgreSQL rejects the operation instead of allowing pgpac to remove an
unplanned dependency.

## Current table-change limitations

Column additions, removals, type changes, defaults, nullability, primary keys,
foreign keys, unique constraints, and check constraints are incremental. A
column rename is not inferred: it is represented as an addition and a
destructive removal, so the old column's values are not copied.

PostgreSQL exclusion constraints are not yet modeled independently. Use an
explicit, reviewed migration for those changes.

If pgpac cannot parse a table shape into supported native alterations, it can
still fall back to dropping and recreating the complete table. That fallback
removes all table data and uses `CASCADE`; always inspect destructive SQL before
authorizing it.

`StopOnDataLossRisk` is currently recorded from the project but is not an
additional enforcement gate.

## Authorization

Without authorization, destructive SQL is emitted as comments and the plan is
reported as blocked. Authorize drops either persistently with
`Target/Plan AllowDrop="true"` or for one invocation with `--allow-drop`.

## Apply

`apply` refuses to execute a destructive plan unless one of these is true:

- the project target allows drops
- `--allow-drop` is passed
- `--force` is passed

`--force` bypasses the authorization check; it does not change dependency
handling or make an otherwise invalid PostgreSQL statement succeed.

## Timeouts and transactions

Project files can define:

- `LockTimeout`
- `StatementTimeout`
- `UseTransaction`

These are applied before executing the plan so release automation and manual
invocations behave consistently. With `UseTransaction="true"`, a failed
operation rolls the complete apply back. Without it, earlier operations remain
committed.
