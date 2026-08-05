#!/usr/bin/env python3
import json
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple


TOOLS = ("pgpac", "d1pac")


def changed_files(base: Optional[str]) -> List[str]:
    if base is None:
        return []

    result = subprocess.run(
        ["git", "diff", "--name-only", "--find-renames", f"{base}...HEAD"],
        check=True,
        capture_output=True,
        text=True,
    )
    return [line for line in result.stdout.splitlines() if line]


def classify(path: str) -> Tuple[str, ...]:
    documentation_only = (
        path == "CHANGELOG.md"
        or path in {"README.md", "AGENTS.md", "LICENSE"}
        or path.startswith("website/")
        or path in {
            "products/pgpac/CHANGELOG.md",
            "products/pgpac/README.md",
            "products/d1pac/CHANGELOG.md",
            "products/d1pac/README.md",
            ".starpac/release.json",
            ".starpac/release-index.json",
        }
    )
    if documentation_only:
        return ()

    if path.startswith(("cmd/pgpac/", "internal/postgres/", "products/pgpac/testdata/")):
        return ("pgpac",)
    if path.startswith(("cmd/d1pac/", "internal/d1/", "products/d1pac/testdata/")):
        return ("d1pac",)
    if path.startswith("internal/pac/") or path in {"go.mod", "go.sum"}:
        return TOOLS

    # Repository and release infrastructure is intentionally conservative.
    # Publishing both is cheaper than silently omitting an affected binary.
    return TOOLS


def manifest(version: str, base: Optional[str]) -> Dict[str, object]:
    commit = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()

    reasons: Dict[str, List[str]] = {tool: [] for tool in TOOLS}
    if base is None:
        for tool in TOOLS:
            reasons[tool].append("initial global Starpac release")
    else:
        for path in changed_files(base):
            for tool in classify(path):
                reasons[tool].append(path)

    artifacts = [tool for tool in TOOLS if reasons[tool]]
    return {
        "version": version,
        "previous": base,
        "commit": commit,
        "artifacts": artifacts,
        "reasons": {tool: reasons[tool] for tool in artifacts},
    }


def main() -> int:
    if len(sys.argv) not in {2, 3}:
        print(f"Usage: {Path(sys.argv[0]).name} <version> [previous-tag]", file=sys.stderr)
        return 2

    version = sys.argv[1]
    base = sys.argv[2] if len(sys.argv) == 3 else None
    json.dump(manifest(version, base), sys.stdout, indent=2)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
