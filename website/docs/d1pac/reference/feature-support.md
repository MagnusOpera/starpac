---
title: Schema Feature Support
---

This page records d1pac's current desired-state coverage. SQLite and D1 support
only a limited set of direct alterations, so many changes use a table rebuild:
create a replacement table, copy same-named columns, replace the original, and
recreate its indexes and triggers.

## Tables and columns

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Create table | Native | `CREATE TABLE`; requires `AllowCreate`. |
| Drop table | Destructive | `DROP TABLE`; requires drop authorization. |
| Add nullable/defaulted column | Native | `ALTER TABLE ... ADD COLUMN`; physical placement is always trailing and requires `AllowAlter`. |
| Add column in strict mode before an existing column | Migration | Table rebuild to preserve declared order; new constraints and defaults are validated. |
| Drop column | Destructive | Table rebuild copying remaining same-named columns; requires drop authorization. |
| Change type, default, or nullability | Migration | Table rebuild copying same-named values into the desired definition. |
| Rename a column | Not inferred | Treated as an added column plus a removed column; the old values are not copied. |

## Integrity constraints

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Primary key | Migration | Table rebuild and validation. |
| Foreign key | Migration | Table rebuild and validation. |
| Unique constraint | Migration | Table rebuild and validation. |
| Check constraint | Migration | Table rebuild and validation. |
| Referenced-table rebuild | Blocked | Automatically blocked when another table references the target. |

## Other schema objects

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Index | Create/drop/replace | Definition changes use `DROP INDEX` followed by `CREATE INDEX`. |
| View | Create/drop/replace | Definition changes use `DROP VIEW` followed by `CREATE VIEW`. |
| Trigger | Create/drop/replace | Definition changes use `DROP TRIGGER` followed by `CREATE TRIGGER`. |
| Schemas, extensions, routines, enums, domains, sequences | Not applicable | These PostgreSQL object types are not part of SQLite/D1. |
| Object comments, owners, and privileges | Not modeled | No desired-state operations are planned for them. |

## Permissions and failure behavior

- `AllowCreate` controls new objects.
- `AllowAlter` controls direct alterations, replacements, and table rebuilds.
- `AllowDrop` or `--allow-drop` controls removed objects and columns.
- Column order is ignored by default. `--strict` requires exact declared order.
- Disabled or unsafe operations remain visible as commented `blocked-*` plan
  entries.
- `UseTransaction="true"` submits the change as one batch; existing values must
  satisfy the desired types and constraints.
- Large rebuilds remain subject to Cloudflare D1 request and execution limits.
- `StopOnDataLossRisk` is recorded but is not an additional enforcement gate.

See the [safety model](./safety-model.md) for rebuild and authorization details.
