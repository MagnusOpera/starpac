---
title: Package Format
---

`.d1pkg` files are ZIP archives containing:

- `manifest.json`
- `model.json`
- `project.xml`
- `scripts/...`
- `checksums/files.sha256`

The manifest identifies the `cloudflare-d1` engine, package id and version,
SQLite compiler version, build time, project file, and source-file checksums.

`model.json` stores the normalized desired schema used by plan and apply. The
original project and SQL inputs make the package independently reviewable.
Cloudflare credentials and target database identifiers are never packaged.
