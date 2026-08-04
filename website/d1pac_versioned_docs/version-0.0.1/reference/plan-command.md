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
  [--allow-drop]
```

`--account-id` defaults to `CLOUDFLARE_ACCOUNT_ID`. The API token defaults to
`CLOUDFLARE_API_TOKEN` and can be supplied with `--api-token` when necessary.

Plans contain summary metadata and ordered operations with a kind, object
type, object key, risk, and executable or blocked SQL. `--script` writes the
complete SQL preview without changing the database.
