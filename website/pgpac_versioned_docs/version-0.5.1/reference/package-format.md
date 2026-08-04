---
title: Package Format
---

`.pgpkg` files are zip archives.

Each package currently contains:

- `manifest.json`
- `model.json`
- `project.xml`
- `scripts/...`
- `checksums/files.sha256`

## Manifest

The manifest records:

- package id
- package version
- PostgreSQL version
- build timestamp
- project file name
- packaged file list and checksums

## Model

`model.json` contains the normalized desired schema model used for planning and apply.

## Scripts

The original SQL inputs are stored under `scripts/` and listed in the checksum file.
