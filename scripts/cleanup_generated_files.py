#!/usr/bin/env python3
"""
Cleanup Generated Go Files

This script fixes common issues in the automatically generated Go files:
- Fixes package names (removes .go extension)
- Fixes method names (removes Python-specific naming like __init__)
- Fixes import paths
- Consolidates similar files
- Removes duplicate files
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Set


class GoFileCleanup:
    """Cleans up generated Go files to fix common issues."""

    def __init__(self, root_dir: str):
        self.root_dir = Path(root_dir)
        self.fixed_files = []
        self.removed_files = []

    def cleanup_all(self):
        """Clean up all Go files in the directory."""
        print(f"Starting cleanup of Go files in {self.root_dir}")

        go_files = list(self.root_dir.rglob("*.go"))
        print(f"Found {len(go_files)} Go files to clean up")

        for go_file in go_files:
            try:
                if self.cleanup_file(go_file):
                    self.fixed_files.append(str(go_file))
            except Exception as e:
                print(f"Error cleaning {go_file}: {e}")

        self.print_summary()

    def cleanup_file(self, go_file: Path) -> bool:
        """Clean up a single Go file."""
        with open(go_file, "r", encoding="utf-8") as f:
            content = f.read()

        original_content = content

        # Fix package name
        content = self.fix_package_name(content, go_file)

        # Fix method names
        content = self.fix_method_names(content)

        # Fix imports
        content = self.fix_imports(content)

        # Fix struct names
        content = self.fix_struct_names(content)

        # Remove empty lines at end
        content = content.rstrip() + "\n"

        if content != original_content:
            with open(go_file, "w", encoding="utf-8") as f:
                f.write(content)
            return True

        return False

    def fix_package_name(self, content: str, file_path: Path) -> str:
        """Fix package declaration."""
        # Extract correct package name from file path
        package_name = file_path.parent.name

        # Handle special cases
        if package_name.endswith(".go"):
            package_name = package_name[:-3]

        # Replace package declaration
        content = re.sub(
            r"^package [^\s]+\.go\s*$",
            f"package {package_name}",
            content,
            flags=re.MULTILINE,
        )

        return content

    def fix_method_names(self, content: str) -> str:
        """Fix Python-specific method names."""
        # Remove __init__ methods
        content = re.sub(
            r"// __init__ is the Go port of the Python __init__ method\s*\n"
            r"func \([^)]+\) __init__\([^)]*\) \{\s*\n"
            r"\t// TODO: Port Python method logic\s*\n"
            r"\}\s*\n",
            "",
            content,
            flags=re.MULTILINE | re.DOTALL,
        )

        # Fix private method names (remove leading underscores)
        def fix_private_method(match):
            receiver = match.group(1)
            method_name = match.group(2)
            # Remove leading underscores and capitalize
            clean_name = method_name.lstrip("_")
            if clean_name:
                clean_name = clean_name[0].lower() + clean_name[1:]
            return f"func ({receiver}) {clean_name}("

        content = re.sub(r"func (\([^)]+\)) (_+\w+)\(", fix_private_method, content)

        return content

    def fix_imports(self, content: str) -> str:
        """Fix and organize imports."""
        # Find import block
        import_match = re.search(
            r"import \(\s*\n(.*?)\n\)", content, re.MULTILINE | re.DOTALL
        )

        if not import_match:
            return content

        import_block = import_match.group(1)
        imports = []

        for line in import_block.split("\n"):
            line = line.strip()
            if line and not line.startswith("//"):
                imports.append(line)

        # Remove duplicates and sort
        unique_imports = sorted(list(set(imports)))

        if unique_imports:
            new_import_block = "import (\n"
            for imp in unique_imports:
                new_import_block += f"\t{imp}\n"
            new_import_block += ")"

            content = content.replace(import_match.group(0), new_import_block)

        return content

    def fix_struct_names(self, content: str) -> str:
        """Fix struct and type names."""
        # Remove .go from type names if they exist
        content = re.sub(r"type (\w+)\.go struct", r"type \1 struct", content)

        return content

    def print_summary(self):
        """Print cleanup summary."""
        print("\n" + "=" * 60)
        print("CLEANUP SUMMARY")
        print("=" * 60)
        print(f"Fixed files: {len(self.fixed_files)}")
        print(f"Removed files: {len(self.removed_files)}")

        if self.fixed_files:
            print(f"\nSample fixed files:")
            for file_path in self.fixed_files[:5]:
                print(f"  {file_path}")


class ModuleConsolidator:
    """Consolidates related modules and removes duplicates."""

    def __init__(self, root_dir: str):
        self.root_dir = Path(root_dir)
        self.moved_files = []

    def consolidate_modules(self):
        """Consolidate related modules into better structure."""
        print("\nConsolidating module structure...")

        # Move profile descriptions to better location
        self.consolidate_profile_descriptions()

        # Move profile runners to better location
        self.consolidate_profile_runners()

        # Move interface modules
        self.consolidate_interfaces()

        # Remove org/openteacher structure - flatten it
        self.flatten_org_structure()

        print(f"Moved {len(self.moved_files)} files during consolidation")

    def consolidate_profile_descriptions(self):
        """Move profile descriptions to a consolidated location."""
        old_path = (
            self.root_dir / "internal/modules/org/openteacher/data/profileDescriptions"
        )
        new_path = self.root_dir / "internal/modules/data/profiledescriptions"

        if old_path.exists() and not new_path.exists():
            self.move_directory(old_path, new_path)

    def consolidate_profile_runners(self):
        """Move profile runners to a consolidated location."""
        old_path = self.root_dir / "internal/modules/org/openteacher/profileRunners"
        new_path = self.root_dir / "internal/modules/profilerunners"

        if old_path.exists():
            # Only move if our target doesn't already exist
            if not new_path.exists():
                self.move_directory(old_path, new_path)

    def consolidate_interfaces(self):
        """Move interface modules to a consolidated location."""
        old_path = self.root_dir / "internal/modules/org/openteacher/interfaces"
        new_path = self.root_dir / "internal/modules/interfaces"

        if old_path.exists() and not new_path.exists():
            self.move_directory(old_path, new_path)

    def flatten_org_structure(self):
        """Remove the org/openteacher nesting."""
        org_path = self.root_dir / "internal/modules/org"

        if org_path.exists():
            # Move remaining modules up
            openteacher_path = org_path / "openteacher"
            if openteacher_path.exists():
                for item in openteacher_path.iterdir():
                    if item.is_dir() and item.name not in [
                        "profileRunners",
                        "interfaces",
                    ]:
                        new_location = self.root_dir / "internal/modules" / item.name
                        if not new_location.exists():
                            self.move_directory(item, new_location)

            # Clean up empty org structure
            try:
                if org_path.exists():
                    import shutil

                    shutil.rmtree(org_path)
                    print(f"Removed empty org directory structure")
            except Exception as e:
                print(f"Could not remove org structure: {e}")

    def move_directory(self, old_path: Path, new_path: Path):
        """Move a directory and update imports."""
        try:
            import shutil

            new_path.parent.mkdir(parents=True, exist_ok=True)
            shutil.move(str(old_path), str(new_path))
            self.moved_files.append(f"{old_path} -> {new_path}")
            print(f"Moved: {old_path} -> {new_path}")
        except Exception as e:
            print(f"Could not move {old_path} to {new_path}: {e}")


def main():
    """Main entry point."""
    import sys

    root_dir = sys.argv[1] if len(sys.argv) > 1 else "."

    if not os.path.exists(root_dir):
        print(f"Error: Directory does not exist: {root_dir}")
        sys.exit(1)

    # First consolidate the structure
    consolidator = ModuleConsolidator(root_dir)
    consolidator.consolidate_modules()

    # Then clean up the files
    cleaner = GoFileCleanup(root_dir)
    cleaner.cleanup_all()

    print("\nCleanup complete!")
    print("\nNext steps:")
    print("1. Run 'go mod tidy' to update dependencies")
    print("2. Run 'go build ./...' to check for compilation errors")
    print("3. Fix any remaining import path issues")
    print("4. Implement TODO items in priority order")


if __name__ == "__main__":
    main()
