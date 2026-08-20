---
title: apply Command
---

`apply` recomputes the plan from live state and executes it through the
Cloudflare D1 API.

```bash
d1pac apply \
  --package <file.d1pkg> \
  --account-id <cloudflare-account-id> \
  --database <name-or-uuid> \
  [--allow-drop] \
  [--force] \
  [--strict]
```

When `UseTransaction="true"`, executable operations are sent as one D1 batch.
D1 rolls the batch back if a statement fails. d1pac prepends
`PRAGMA defer_foreign_keys = ON` so temporary violations created by table
rebuilds are checked at the end of the deployment.

- `--allow-drop` explicitly authorizes destructive operations.
- `--force` bypasses the destructive-operation guard.
- `--strict` requires live table column order to match the desired declaration
  order. Without it, column order is ignored and addable columns are appended.
- Blocked and comment-only operations are never executed.

Apply talks directly to the administrative Cloudflare API. It does not modify
Wrangler's `d1_migrations` ledger.
