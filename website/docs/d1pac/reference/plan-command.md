---
title: plan Command
---

`plan` compares a built package with a live Cloudflare D1 database.

```bash
d1pac plan \
  --package <file.d1pkg> \
  --account-id <cloudflare-account-id> \
  --database <name-or-uuid> \
  [--format text|json] \
  [--script <file>] \
  [--allow-drop] \
  [--strict]
```

`--account-id` defaults to `CLOUDFLARE_ACCOUNT_ID`. The API token defaults to
`CLOUDFLARE_API_TOKEN` and can be supplied with `--api-token` when necessary.

Plans contain summary metadata and ordered operations with a kind, object
type, object key, risk, an optional reason for blocked operations, and
executable or blocked SQL. Text output prints the reason directly beneath a
blocked operation. `--script` writes the
complete SQL preview without changing the database.

By default, table comparison is column-order independent. An addable column is
appended with SQLite's `ALTER TABLE ... ADD COLUMN` even when it appears before
an existing column in the desired SQL. Later plans treat the resulting physical
column order as equivalent. Pass `--strict` to require the live column order to
match the desired declaration order; mismatches then use the normal table
rebuild and safety rules.
