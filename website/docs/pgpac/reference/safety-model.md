---
title: Safety Model
---

pgpac classifies every planned operation as `safe` or `destructive`. Operations
that require drop authorization remain visible as commented `blocked-*`
operations when authorization is absent.

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

Only column additions and removals are incremental. Other table differences,
including a changed column type, default, nullability, or primary key, currently
fall back to dropping and recreating the complete table. That fallback removes
all table data and uses `CASCADE`, so always inspect the generated SQL before
authorizing it.

Table-level `FOREIGN KEY`, `UNIQUE`, and `CHECK` constraints are not yet modeled
independently. Some constraint-only changes may therefore be missed rather than
planned. Use an explicit, reviewed migration until first-class constraint
diffing is available.

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
