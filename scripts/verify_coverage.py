#!/usr/bin/env python3
"""
Verify Coverage of Python to Go Conversion

This script verifies that every Python file in the legacy directory
has a corresponding Go file in the internal/modules directory.
"""

import os
from pathlib import Path


def get_python_files(legacy_dir):
    """Get all Python files in the legacy directory."""
    legacy_path = Path(legacy_dir)
    python_files = []

    for py_file in legacy_path.rglob("*.py"):
        if py_file.name != "__init__.py":  # Skip __init__.py files
            python_files.append(py_file)

    return python_files


def calculate_go_path(py_file, legacy_root, target_root):
    """Calculate expected Go file path from Python file path."""
    rel_path = py_file.relative_to(legacy_root)

    # Transform path components
    parts = list(rel_path.parts)

    # Handle the flattened structure after consolidation
    if parts and parts[0] == "modules":
        parts = parts[1:]  # Remove 'modules'

    # Handle the flattened org/openteacher structure - this was consolidated
    if len(parts) >= 2 and parts[0:2] == ["org", "openteacher"]:
        parts = parts[2:]  # Remove 'org/openteacher'

    # Map to actual consolidated locations
    if parts and len(parts) >= 2:
        if parts[0] == "data" and parts[1] == "profileDescriptions":
            # data/profileDescriptions -> data/profiledescriptions
            parts = ["data", "profiledescriptions"] + parts[2:]
        elif parts[0] == "profileRunners":
            # profileRunners -> profilerunners
            parts = ["profilerunners"] + parts[1:]
        # interfaces, logic, misc are moved directly under modules/

    # Convert Python file to Go file
    if parts:
        parts[-1] = parts[-1].replace(".py", ".go")

    # Construct target path
    go_path = target_root / "internal" / "modules" / Path(*parts)

    return go_path


def verify_coverage(legacy_dir="legacy", target_dir="."):
    """Verify that all Python files have corresponding Go files."""
    legacy_path = Path(legacy_dir)
    target_path = Path(target_dir)

    python_files = get_python_files(legacy_path)

    print(f"Found {len(python_files)} Python files to verify")
    print("=" * 60)

    missing_files = []
    existing_files = []

    for py_file in python_files:
        go_file = calculate_go_path(py_file, legacy_path, target_path)

        if go_file.exists():
            existing_files.append((str(py_file), str(go_file)))
        else:
            missing_files.append((str(py_file), str(go_file)))

    # Print results
    print(f"✅ Files with Go equivalents: {len(existing_files)}")
    print(f"❌ Missing Go files: {len(missing_files)}")
    print(
        f"📊 Coverage: {len(existing_files)}/{len(python_files)} ({100 * len(existing_files) / len(python_files):.1f}%)"
    )

    if missing_files:
        print("\nMissing Go files:")
        print("-" * 40)
        for py_file, go_file in missing_files[:10]:  # Show first 10
            print(f"Python: {py_file}")
            print(f"  Missing Go: {go_file}")
            print()

        if len(missing_files) > 10:
            print(f"... and {len(missing_files) - 10} more missing files")

    # Show some examples of successful mappings
    if existing_files:
        print("\nSuccessful mappings (sample):")
        print("-" * 40)
        for py_file, go_file in existing_files[:5]:
            print(f"✅ {py_file}")
            print(f"   -> {go_file}")
            print()

    # Analyze directory coverage
    print("\nDirectory Coverage Analysis:")
    print("-" * 40)
    analyze_directory_coverage(python_files, legacy_path, target_path)

    return len(existing_files), len(missing_files)


def analyze_directory_coverage(python_files, legacy_path, target_path):
    """Analyze coverage by directory."""
    dir_stats = {}

    for py_file in python_files:
        # Get the first directory level under legacy
        rel_path = py_file.relative_to(legacy_path)
        if rel_path.parts:
            first_dir = rel_path.parts[0]

            if first_dir not in dir_stats:
                dir_stats[first_dir] = {"total": 0, "existing": 0}

            dir_stats[first_dir]["total"] += 1

            # Check if Go file exists
            go_file = calculate_go_path(py_file, legacy_path, target_path)
            if go_file.exists():
                dir_stats[first_dir]["existing"] += 1

    # Print directory stats
    for directory, stats in sorted(dir_stats.items()):
        total = stats["total"]
        existing = stats["existing"]
        percentage = 100 * existing / total if total > 0 else 0

        status = "✅" if percentage == 100 else "🚧" if percentage > 50 else "❌"
        print(
            f"{status} {directory:<20} {existing:>3}/{total:<3} ({percentage:>5.1f}%)"
        )


def main():
    """Main entry point."""
    print("OpenTeacher Python to Go Conversion Coverage Verification")
    print("=" * 60)

    existing, missing = verify_coverage()

    print("\n" + "=" * 60)
    print("VERIFICATION SUMMARY")
    print("=" * 60)

    if missing == 0:
        print("🎉 PERFECT COVERAGE! All Python files have Go equivalents.")
    elif missing < 5:
        print(f"🎯 EXCELLENT COVERAGE! Only {missing} files missing.")
    elif missing < 20:
        print(f"👍 GOOD COVERAGE! {missing} files missing.")
    else:
        print(f"⚠️  PARTIAL COVERAGE: {missing} files still need conversion.")

    total = existing + missing
    coverage_percent = 100 * existing / total if total > 0 else 0

    print(
        f"\nFinal Score: {existing}/{total} files converted ({coverage_percent:.1f}%)"
    )

    if missing == 0:
        print("\n🚀 Ready for implementation phase!")
        print("Next steps:")
        print("1. Run 'go build ./...' to check compilation")
        print("2. Implement TODO items in priority order")
        print("3. Add comprehensive tests")
        print("4. Begin functionality implementation")


if __name__ == "__main__":
    main()
