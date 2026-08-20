---
title: plan Command
---

`plan` compares a built package to a live PostgreSQL database and emits the resulting operations.

```bash
pgpac plan \
  --package <file.pgpkg> \
  --connection <postgres-uri> \
  [--format text|json] \
  [--script <file>] \
  [--allow-drop]
```

## Required flags

- `--package`: path to a `.pgpkg` file
- `--connection`: PostgreSQL connection string

## Optional flags

- `--format`: `text` or `json`, defaults to `text`
- `--script`: writes rendered SQL to a file
- `--allow-drop`: allows destructive operations to be rendered as executable SQL instead of blocked comments

## Plan model

Plans include:

- summary metadata
- destructive-operation detection
- ordered operations with kind, object type, object key, risk, and SQL
