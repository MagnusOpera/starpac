---
title: build Command
---

`build` compiles a `.pgpac` project into a `.pgpkg` archive.

```bash
pgpac build --project <file.pgpac> --output <dir-or-file>
```

## Required flags

- `--project`: path to the `.pgpac` project file
- `--output`: output directory or a direct `.pgpkg` file path

## Output

The command prints the resolved package path to stdout.

When the output points to a directory, the file name is `<PackageId>.pgpkg`.
