---
title: Quickstart
---

This walkthrough uses the sample project in the repository.

## 1. Build a package

```bash
pgpac build --project testdata/sample/sample.pgpac --output out/
```

That produces `out/SampleProject.pgpkg`.

## 2. Generate a plan against a target database

```bash
pgpac plan \
  --package out/SampleProject.pgpkg \
  --connection "postgres://user:pass@localhost:5432/appdb"
```

Use `--format json` if you want machine-readable output, or `--script plan.sql` to write the rendered SQL preview to disk.

## 3. Apply the plan

```bash
pgpac apply \
  --package out/SampleProject.pgpkg \
  --connection "postgres://user:pass@localhost:5432/appdb"
```

If the plan contains destructive operations, `apply` stops unless you explicitly opt in with `--allow-drop` or `--force`.
