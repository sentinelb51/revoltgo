import glob
import os
import re
import subprocess
import sys
from typing import List, Optional, Set, Tuple

# Configuration
OUTPUT_FILENAME = "revoltgo_combined_src.go"  # Temporary combined file
GEN_OUTPUT_NAME = "revoltgo_msgp_gen.go"      # Final generated file name
IGNORE_FILES = {"_gen.go", "_test.go", "gen_msgp.py", "msgp_codegen.py", "generate_code.py"}

# Regex patterns
RX_PACKAGE = re.compile(r'^\s*package\s+(\w+)', re.MULTILINE)
RX_IMPORT_BLOCK = re.compile(r'^\s*import\s*\((.*?)\)', re.DOTALL | re.MULTILINE)
RX_IMPORT_SINGLE = re.compile(r'^\s*import\s+(".*?"|[\w_.]+\s+".*?")', re.MULTILINE)
RX_GENERATE = re.compile(r'^\s*//go:generate\s+msgp', re.MULTILINE)


def get_target_files() -> List[str]:
    target_files: List[str] = []
    source_files = glob.glob("*.go")

    for fname in source_files:
        if any(fname.endswith(i) for i in IGNORE_FILES):
            continue
        if fname in (OUTPUT_FILENAME, GEN_OUTPUT_NAME):
            continue

        try:
            with open(fname, 'r', encoding='utf-8') as f:
                content = f.read()
                if RX_GENERATE.search(content):
                    target_files.append(fname)
        except OSError as e:
            print(f"Warning: Could not read {fname}: {e}")

    return target_files


def parse_file(filename: str) -> Tuple[Optional[str], Set[str], str]:
    with open(filename, 'r', encoding='utf-8') as f:
        content = f.read()

    pkg_match = RX_PACKAGE.search(content)
    pkg = pkg_match.group(1) if pkg_match else None

    imports: Set[str] = set()

    for match in RX_IMPORT_BLOCK.findall(content):
        lines = match.split('\n')
        for line in lines:
            line = line.strip()
            if line and not line.startswith('//'):
                imports.add(line)

    for match in RX_IMPORT_SINGLE.findall(content):
        imports.add(match.strip())

    lines = content.splitlines()
    body_lines: List[str] = []
    in_import_block = False

    for line in lines:
        stripped = line.strip()
        if not stripped:
            body_lines.append(line)
            continue
        if RX_PACKAGE.match(line):
            continue
        if stripped.startswith("import ("):
            in_import_block = True
            continue
        if in_import_block:
            if stripped == ")":
                in_import_block = False
            continue
        if stripped.startswith("import ") and not in_import_block:
            if '"' in stripped:
                continue
        body_lines.append(line)

    return pkg, imports, "\n".join(body_lines)


def main() -> None:
    files = get_target_files()
    if not files:
        print("No .go files with '//go:generate msgp' found.")
        return

    all_imports: Set[str] = set()
    all_body: List[str] = []
    package_name: Optional[str] = None

    print(f"Detected {len(files)} files with generation directives.")

    for fname in files:
        pkg, imports, body = parse_file(fname)

        if pkg is None:
            print(f"Warning: Could not detect package in {fname}, skipping.")
            continue

        if not package_name:
            package_name = pkg
        elif pkg != package_name:
            if pkg.endswith("_test"):
                continue
            print(f"Error: File {fname} has package {pkg}, expected {package_name}")
            return

        all_imports.update(imports)
        all_body.append(f"\n// --- Content from {fname} ---")
        all_body.append(body)

    if not package_name:
        print("Error: Could not determine package name.")
        return

    print(f"Merging into {OUTPUT_FILENAME}...")
    try:
        with open(OUTPUT_FILENAME, 'w', encoding='utf-8') as f:
            f.write(f"package {package_name}\n\n")
            f.write("import (\n")
            for imp in sorted(all_imports):
                if imp:
                    f.write(f"\t{imp}\n")
            f.write(")\n")
            f.write("\n".join(all_body))
    except OSError as e:
        print(f"Error writing combined file: {e}")
        return

    # Run msgp
    cmd = ["msgp", "-file", OUTPUT_FILENAME, "-o", GEN_OUTPUT_NAME, "-io=false", "-tests=false", "-v=true"]
    print(f"Running: {' '.join(cmd)}")

    all_output_lines = []
    warnings_and_errors = []

    try:
        process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            encoding='utf-8'
        )

        if process.stdout:
            for line in process.stdout:
                line_stripped = line.strip()
                all_output_lines.append(line_stripped)

                # --- Filtering Logic ---
                lower_line = line_stripped.lower()

                # Filter out empty lines
                if not lower_line:
                    continue

                # Filter out standard 'info', 'input', and 'wrote' messages
                # This ensures we only catch actual warnings or unexpected output
                # Ignore "unresolved identifier: Timestamp" because we defined custom unmarshal/marshal
                is_standard_msg = (
                    "info:" in lower_line or
                    "unresolved identifier: timestamp" in lower_line or
                    lower_line.startswith("generated") or
                    lower_line.startswith("input:") or
                    lower_line.startswith("wrote")
                )

                if not is_standard_msg:
                    cleaned = line_stripped.replace("warn: revoltgo_combined_src.go: ", "")
                    warnings_and_errors.append(line_stripped)

        return_code = process.wait()

        if return_code != 0:
            print(f"\nmsgp failed with exit code {return_code}")
            print(" --- Output --- ")
            for l in all_output_lines:
                print(l)
            sys.exit(return_code)

        print(f"Successfully generated {GEN_OUTPUT_NAME}")

        # --- SUMMARY SECTION ---
        # Only print if there are actual warnings (ignoring the standard info noise)
        if warnings_and_errors:
            print("--- Warnings ---")
            for w in warnings_and_errors:
                print(w)

    except Exception as e:
        print(f"Error executing msgp: {e}")
        sys.exit(1)
    finally:
        if os.path.exists(OUTPUT_FILENAME):
            try:
                os.remove(OUTPUT_FILENAME)
                print(f"Cleaned up {OUTPUT_FILENAME}")
            except OSError:
                pass


if __name__ == "__main__":
    main()