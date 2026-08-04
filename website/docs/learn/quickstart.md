---
title: Quickstart
---

## 1. Build a package

```bash
d1pac build \
  --project testdata/sample/sample.d1pac \
  --output out/
```

This produces `out/SampleProject.d1pkg`.

## 2. Plan against D1

```bash
d1pac plan \
  --package out/SampleProject.d1pkg \
  --database my-database \
  --script deployment.sql
```

Use `--format json` when CI needs structured plan output. A database UUID can
be supplied instead of its name.

## 3. Apply

```bash
d1pac apply \
  --package out/SampleProject.d1pkg \
  --database my-database
```

If the plan would remove schema or data, inspect it and explicitly pass
`--allow-drop`. `--force` is reserved for controlled recovery and automation
where all safety checks have already happened elsewhere.
