---
title: Schema Feature Support
---

This page records pgpac's current desired-state coverage. `Native` operations
preserve the table and its remaining rows. `Migration` operations also preserve
rows, but PostgreSQL must validate or convert existing data. `Destructive`
operations require `AllowDrop="true"`, `--allow-drop`, or `--force`.

## Tables and columns

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Create table | Native | `CREATE TABLE`; requires `AllowCreate`. |
| Drop table | Destructive | `DROP TABLE ... CASCADE`; requires drop authorization. |
| Add column | Native | `ALTER TABLE ... ADD COLUMN`; requires `AllowAlter`. |
| Drop column | Destructive | `ALTER TABLE ... DROP COLUMN` without `CASCADE`; preserves remaining data. |
| Change column type | Migration | Explicit `USING column::new_type` conversion; incompatible values fail. |
| Set or remove a default | Native | `ALTER COLUMN ... SET/DROP DEFAULT`. |
| Set `NOT NULL` | Migration | Existing rows are validated. |
| Remove `NOT NULL` | Native | `ALTER COLUMN ... DROP NOT NULL`. |
| Rename a column | Not inferred | Appears as an added column plus a destructive removed column; values are not copied. |

## Integrity constraints

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Primary key | Migration | Native add, drop, rename, or replacement. |
| Foreign key | Migration | Native add, drop, rename, or replacement with target validation. |
| Unique constraint | Migration | Native add, drop, rename, or replacement with existing-row validation. |
| Check constraint | Migration | Native add, drop, rename, or replacement with existing-row validation. |
| Unnamed constraints | Supported | Matched semantically without depending on PostgreSQL-generated names. |
| Referenced-key alteration | Migration | Dependent foreign keys are dropped first and recreated after the referenced change. |
| Exclusion constraint | Unsupported | Use an explicit reviewed migration. |

## Other schema objects

| Feature | Status | Planned behavior |
| --- | --- | --- |
| Schema | Create/drop | Drops are destructive and use `CASCADE`. |
| Extension | Create/drop/recreate | Version changes can require destructive recreation. |
| Index | Create/drop/recreate | Definition changes currently use destructive recreation. |
| View | Create/replace/drop | Changes use `CREATE OR REPLACE VIEW`. |
| Function or procedure | Create/replace/drop | Changes use `CREATE OR REPLACE`. |
| Enum, domain, or sequence | Create/drop/recreate | Unsupported in-place changes fall back to destructive recreation. |
| Table and column comments | Set/clear | Native comment operations; requires `AllowAlter`. |
| Owners and privileges | Not modeled | Project comparison flags are reserved; no owner or privilege operations are currently planned. |

## Permissions and failure behavior

- `AllowCreate` controls new objects.
- `AllowAlter` controls native changes, migrations, and replacements.
- `AllowDrop` or `--allow-drop` controls destructive removals and recreations.
- Disabled operations remain visible as commented `blocked-*` plan entries.
- `UseTransaction="true"` rolls the complete apply back when conversion,
  dependency, uniqueness, foreign-key, or check validation fails.
- `StopOnDataLossRisk` is recorded but is not an additional enforcement gate.

See the [safety model](./safety-model.md) for authorization details.
