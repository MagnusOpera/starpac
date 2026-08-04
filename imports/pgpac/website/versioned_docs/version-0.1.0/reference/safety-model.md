---
title: Safety Model
---

`pgpac` is opinionated about destructive operations.

## Planning

When the target contains objects that are absent from the desired model:

- if drops are allowed, the plan emits executable destructive SQL
- otherwise, the plan emits blocked operations with commented SQL

This makes the risk visible without silently applying destructive changes.

## Apply

`apply` refuses to execute a destructive plan unless one of these is true:

- the project target allows drops
- `--allow-drop` is passed
- `--force` is passed

## Timeouts and transactions

Project files can define:

- `LockTimeout`
- `StatementTimeout`
- `UseTransaction`

These are applied before executing the plan so release automation and manual invocations behave consistently.
