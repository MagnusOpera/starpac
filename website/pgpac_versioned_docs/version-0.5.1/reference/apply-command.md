---
title: apply Command
---

`apply` executes a generated plan directly against a PostgreSQL target.

```bash
pgpac apply \
  --package <file.pgpkg> \
  --connection <postgres-uri> \
  [--allow-drop] \
  [--force]
```

## Required flags

- `--package`: path to a `.pgpkg` file
- `--connection`: PostgreSQL connection string

## Optional flags

- `--allow-drop`: explicitly allow destructive operations
- `--force`: bypass destructive-operation protection

## Apply behavior

- Applies lock and statement timeouts from the project file when present.
- Runs in a transaction when `UseTransaction="true"`.
- Skips blocked or comment-only operations.
- Fails on the first database error and returns the operation context.
