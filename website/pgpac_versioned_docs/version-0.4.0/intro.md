---
id: intro
title: pgpac
slug: /
---

`pgpac` is a Go-first PostgreSQL schema packaging tool in the spirit of `sqlpackage`, built around a standalone CLI and an XML project file.

The workflow is straightforward:

1. Define the desired schema state in a `.pgpac` project.
2. Build that project into a `.pgpkg` artifact.
3. Compare the package against a live PostgreSQL database.
4. Apply the generated plan with destructive-operation safeguards.

The tool currently supports:

- XML project parsing and validation
- offline desired-state parsing from SQL files
- `.pgpkg` package creation with manifest and checksums
- PostgreSQL 17+ introspection
- typed plan generation
- guarded apply execution

Use the docs in this site for installation, quickstart, command reference, and release notes.
