---
title: build Command
---

`build` compiles a `.d1pac` project into a `.d1pkg` archive without contacting
Cloudflare.

```bash
d1pac build --project <file.d1pac> --output <dir-or-file>
```

Desired SQL is executed against an isolated in-memory SQLite database. This
validates the DDL and produces a normalized schema model using the same catalog
and PRAGMA surfaces used for live comparisons.

When `--output` names a directory, the result is
`<PackageId>.d1pkg`. The resolved path is printed to stdout.
