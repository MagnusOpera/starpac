---
title: Installation
---

## Homebrew

```bash
brew tap magnusopera/tap
brew install d1pac
```

## Source

```bash
git clone https://github.com/MagnusOpera/d1pac.git
cd d1pac
make build
```

The resulting binary is `./d1pac`.

## Cloudflare credentials

`plan` requires D1 read access. `apply` requires D1 write access.

```bash
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_API_TOKEN="your-scoped-api-token"
```

Credentials are read at execution time and are never written into packages.
