#!/usr/bin/env python3
from __future__ import annotations

import re
import subprocess
from pathlib import Path


def get_commit_hash(repo_root: Path) -> str:
    """Fetch and validate the current Git commit hash."""
    output = subprocess.check_output(
        ["git", "rev-parse", "HEAD"],
        cwd=repo_root,
        text=True,
    )
    commit = output.strip()

    # Support both standard SHA-1 (40) and newer SHA-256 (64) git hashes
    if not re.fullmatch(r"[0-9a-f]{40,64}", commit):
        raise ValueError(f"Unexpected commit hash format: {commit!r}")

    return commit


def replace_commit_line(file_path: Path, commit: str) -> bool:
    """Update the COMMIT variable in the specified file."""
    content = file_path.read_text(encoding="utf-8")

    # Match 'var COMMIT =' (preserving indentation) and replace the rest of the line
    pattern = re.compile(r"^(\s*var\s+COMMIT\s*=).*$", re.MULTILINE)
    new_content, count = pattern.subn(rf'\1 "{commit}"', content)

    if count == 0:
        return False

    file_path.write_text(new_content, encoding="utf-8")
    return True


def main() -> int | str:
    repo_root = Path(__file__).resolve().parent.parent
    target = repo_root / "revoltgo.go"

    if not target.exists():
        return f"ERROR: File not found: {target}"

    try:
        commit = get_commit_hash(repo_root)
    except (subprocess.CalledProcessError, ValueError) as exc:
        return f"ERROR: Failed to get commit hash: {exc}"

    if not replace_commit_line(target, commit):
        return "ERROR: No line starting with 'var COMMIT =' found in revoltgo.go"

    print(f"Updated {target.name} with COMMIT={commit}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())