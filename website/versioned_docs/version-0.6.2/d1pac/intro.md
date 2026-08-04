---
title: Introduction
slug: /d1pac/
---

`d1pac` is a Go-first desired-state schema packaging tool for Cloudflare D1.
It brings the package, compare, and apply workflow of DACPAC and pgpac to D1's
SQLite semantics.

The workflow is:

1. Define the desired schema in ordinary SQLite SQL files.
2. Build an immutable `.d1pkg` artifact.
3. Compare it with a live D1 database.
4. Review and apply the generated plan.

The tool models tables, columns, foreign keys, explicit indexes, views, and
triggers. Application data, seed records, and backfills remain explicit data
migrations rather than desired schema state.
