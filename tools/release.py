#!/usr/bin/env python3
"""Tag a CalVer release, push it, then open GitHub's release form for the changelog."""
from __future__ import annotations

import re
import subprocess
import webbrowser
from datetime import datetime, timezone
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent


def git(*args: str) -> str:
    """Run git in the repository root and return its stripped output."""
    return subprocess.check_output(["git", *args], cwd=REPO_ROOT, text=True).strip()


def next_tag(tags: list[str], prefix: str) -> str:
    """This month's tag: YY.M.N, the counter restarting each month.

    A v-prefixed tag counts too. Nothing here mints one, but if a stray tag exists
    the number is taken, and reissuing it would point two releases at one version.
    """
    pattern = re.compile(rf"v?{re.escape(prefix)}\.(\d+)")
    used = [int(m[1]) for t in tags if (m := pattern.fullmatch(t))]
    return f"{prefix}.{max(used, default=0) + 1}"


def release_url(tag: str) -> str:
    """GitHub's "draft a new release" page for origin, pre-filled with the tag."""
    slug = re.sub(r"^.*github\.com[:/]|\.git$", "", git("remote", "get-url", "origin"))
    return f"https://github.com/{slug}/releases/new?tag={tag}"


def main() -> int | str:
    git("fetch", "origin", "main", "--tags")

    if git("status", "--porcelain"):
        print("Warning: working tree is dirty; the tag points at HEAD, not at your uncommitted changes")

    ancestry = ["git", "merge-base", "--is-ancestor", "HEAD", "origin/main"]
    if subprocess.call(ancestry, cwd=REPO_ROOT) != 0:
        return "ERROR: HEAD is not on origin/main; push your commits first"

    now = datetime.now(timezone.utc)
    tag = next_tag(git("tag", "--list").splitlines(), f"{now.year % 100}.{now.month}")

    print(f"Tagging {tag} at {git('log', '-1', '--format=%h %s')}")
    if input("Push it? [Y/n] ").strip().lower() not in ("", "y"):
        return "Aborted"

    git("tag", "-a", tag, "-m", f"Release {tag}")
    git("push", "origin", tag)

    url = release_url(tag)
    print(f"Pushed {tag}. Write the changelog at:\n  {url}")
    webbrowser.open(url)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as exc:
        raise SystemExit(f"ERROR: {' '.join(exc.cmd)} exited with {exc.returncode}")
