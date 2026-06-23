#!/usr/bin/env python

# Compile the per-change fragment files under release-notes/current/ into a single
# versioned release notes YAML (release-notes/<version>.yaml), then clear the
# fragments so the directory is ready for the next development cycle.
#
# Usage: python compile.py <version> [date]
#   <version>  e.g. v1.9.0 (also accepts 1.9.0)
#   [date]     e.g. "June 23, 2026"; defaults to "Pending"

import os
import re
import sys

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "..", ".."))
CURRENT_DIR = os.path.join(ROOT_DIR, "release-notes", "current")

# (YAML key, header comment or None, fragment subdirectory). Order is preserved
# in the output and the keys match what tools/src/release-notes-docs/yml2md.py reads.
SECTIONS = [
    ("breaking changes",
     "# Changes that are expected to cause an incompatibility with previous versions, such as deletions or modifications to existing APIs.",
     "breaking_changes"),
    ("security updates",
     "# Updates addressing vulnerabilities, security flaws, or compliance requirements.",
     "security_updates"),
    ("new features",
     "# New features or capabilities added in this release.",
     "new_features"),
    ("bug fixes", None, "bug_fixes"),
    ("performance improvements",
     "# Enhancements that improve performance.",
     "performance_improvements"),
    ("deprecations",
     "# Deprecated features or APIs.",
     "deprecations"),
    ("Other changes",
     "# Other notable changes not covered by the above sections.",
     "other_changes"),
]


def fragment_sort_key(filename):
    # Files are named <pr-number>-<slug>.md; sort numerically by PR then by name.
    m = re.match(r"^(\d+)", filename)
    return (0, int(m.group(1)), filename) if m else (1, 0, filename)


def read_fragments(section_dir):
    path = os.path.join(CURRENT_DIR, section_dir)
    if not os.path.isdir(path):
        return [], []
    files = [f for f in os.listdir(path)
             if f.endswith(".md") and f != "README.md"]
    files.sort(key=fragment_sort_key)
    bullets, paths = [], []
    for f in files:
        fp = os.path.join(path, f)
        with open(fp) as fh:
            # One change per file: collapse any internal whitespace into a single
            # line so it renders as a single bullet, matching the existing format.
            text = " ".join(fh.read().split())
        if text:
            bullets.append(text)
            paths.append(fp)
    return bullets, paths


def main():
    if len(sys.argv) not in (2, 3):
        print("Usage: python compile.py <version> [date]")
        sys.exit(1)

    version = sys.argv[1]
    if not version.startswith("v"):
        version = "v" + version
    date = sys.argv[2].strip() if len(sys.argv) == 3 else ""
    if not date:
        date = "Pending"

    out_lines = ["date: {}".format(date), ""]
    consumed = []
    for key, comment, section_dir in SECTIONS:
        bullets, paths = read_fragments(section_dir)
        consumed.extend(paths)
        if comment:
            out_lines.append(comment)
        out_lines.append("{}: |".format(key))
        for b in bullets:
            out_lines.append("  {}".format(b))
        out_lines.append("")

    output = "\n".join(out_lines).rstrip("\n") + "\n"
    out_file = os.path.join(ROOT_DIR, "release-notes", "{}.yaml".format(version))
    with open(out_file, "w") as f:
        f.write(output)
    print("Wrote {} ({} fragments)".format(out_file, len(consumed)))

    # Clear the fragments now that they are compiled into the release file.
    for p in consumed:
        os.remove(p)
    if consumed:
        print("Removed {} fragment file(s) from release-notes/current/".format(len(consumed)))


if __name__ == "__main__":
    main()
