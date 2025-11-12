#!/usr/bin/env python3
"""
OpenTeacher Batch Legacy Module Converter

This script automatically converts multiple Python modules from the legacy codebase
to Go modules following the established OpenTeacher module system patterns.

Usage:
    python scripts/batch_convert_modules.py [options]

Examples:
    # Convert all UI modules
    python scripts/batch_convert_modules.py --category interfaces/qt

    # Convert specific modules by pattern
    python scripts/batch_convert_modules.py --pattern "*startWidget*"

    # Convert all modules (be careful!)
    python scripts/batch_convert_modules.py --all

    # Dry run to see what would be converted
    python scripts/batch_convert_modules.py --category interfaces/qt --dry-run
"""

import argparse
import fnmatch
import json
import os
import sys
import time
from pathlib import Path
from typing import Dict, List, Set

# Import the main converter
from convert_legacy_module import LegacyModuleConverter, ModuleInfo


class BatchConverter:
    """Handles batch conversion of legacy modules"""

    def __init__(
        self, legacy_root: str = "legacy", output_root: str = "internal/modules"
    ):
        self.converter = LegacyModuleConverter(legacy_root, output_root)
        self.legacy_root = Path(legacy_root)
        self.conversion_log = []
        self.failed_conversions = []

    def find_legacy_modules(
        self, category: str = None, pattern: str = None
    ) -> List[Path]:
        """Find legacy Python modules to convert"""
        modules_root = self.legacy_root / "modules" / "org" / "openteacher"

        if not modules_root.exists():
            print(f"Legacy modules root not found: {modules_root}")
            return []

        python_files = []

        if category:
            # Look for modules in specific category
            category_path = modules_root / category
            if category_path.exists():
                python_files.extend(category_path.rglob("*.py"))
        else:
            # Find all Python modules
            python_files.extend(modules_root.rglob("*.py"))

        # Filter by pattern if provided
        if pattern:
            python_files = [f for f in python_files if fnmatch.fnmatch(f.name, pattern)]

        # Filter out __init__.py and test files
        python_files = [
            f
            for f in python_files
            if f.name != "__init__.py" and "test" not in f.name.lower()
        ]

        return sorted(python_files)

    def get_conversion_priorities(self) -> Dict[str, int]:
        """Get conversion priorities for different module types"""
        return {
            # Core infrastructure (highest priority)
            "event": 1,
            "settings": 1,
            "metadata": 1,
            # UI infrastructure
            "ui": 2,
            "buttonRegister": 2,
            "startWidget": 2,
            "translator": 2,
            # Data and logic
            "data": 3,
            "logic": 3,
            "loader": 3,
            "saver": 3,
            # Specific interfaces
            "interface": 4,
            "dialog": 4,
            "enterer": 4,
            "teacher": 4,
            # Specialized modules
            "profileRunner": 5,
            "test": 6,
            "unknown": 10,
        }

    def prioritize_modules(self, modules: List[Path]) -> List[Path]:
        """Sort modules by conversion priority"""
        priorities = self.get_conversion_priorities()

        def get_priority(module_path: Path) -> int:
            # Analyze module to determine type
            try:
                module_info = self.converter.analyze_legacy_module(module_path)
                return priorities.get(module_info.module_type, 10)
            except:
                return 10

        return sorted(modules, key=get_priority)

    def check_dependencies(self, module_info: ModuleInfo) -> Set[str]:
        """Check if module dependencies are satisfied"""
        missing_deps = set()

        # Check if required modules exist
        for req in module_info.requires:
            # This is a simplified check - would need more sophisticated dependency resolution
            if req not in [
                "event",
                "settings",
                "metadata",
                "ui",
            ]:  # Known available modules
                missing_deps.add(req)

        return missing_deps

    def convert_modules(self, modules: List[Path], dry_run: bool = False) -> Dict:
        """Convert a list of modules"""
        results = {"converted": [], "failed": [], "skipped": [], "total": len(modules)}

        print(f"Found {len(modules)} modules to convert")

        if dry_run:
            print("\nDRY RUN - No files will be created")

        # Prioritize modules
        prioritized_modules = self.prioritize_modules(modules)

        for i, module_path in enumerate(prioritized_modules, 1):
            print(f"\n[{i}/{len(modules)}] Converting {module_path.name}...")

            if dry_run:
                try:
                    module_info = self.converter.analyze_legacy_module(module_path)
                    print(f"  Module: {module_info.name} ({module_info.module_type})")
                    print(f"  Requires: {module_info.requires}")
                    print(f"  Methods: {len(module_info.methods)}")
                    results["converted"].append(str(module_path))
                except Exception as e:
                    print(f"  ERROR: {e}")
                    results["failed"].append(str(module_path))
                continue

            try:
                # Convert the module
                module_info = self.converter.convert_module(str(module_path))

                # Check dependencies
                missing_deps = self.check_dependencies(module_info)
                if missing_deps:
                    print(f"  WARNING: Missing dependencies: {missing_deps}")

                # Log successful conversion
                conversion_entry = {
                    "legacy_path": str(module_path),
                    "module_name": module_info.name,
                    "module_type": module_info.module_type,
                    "requires": module_info.requires,
                    "uses": module_info.uses,
                    "methods_count": len(module_info.methods),
                    "timestamp": time.time(),
                }
                self.conversion_log.append(conversion_entry)
                results["converted"].append(str(module_path))

                print(
                    f"  ✓ Converted to {module_info.name} ({module_info.module_type})"
                )

            except Exception as e:
                print(f"  ✗ Failed: {e}")
                results["failed"].append(str(module_path))
                self.failed_conversions.append(
                    {
                        "path": str(module_path),
                        "error": str(e),
                        "timestamp": time.time(),
                    }
                )

        return results

    def save_conversion_log(self, output_file: str = "conversion_log.json") -> None:
        """Save conversion log to file"""
        log_data = {
            "conversions": self.conversion_log,
            "failures": self.failed_conversions,
            "timestamp": time.time(),
            "summary": {
                "total_attempted": len(self.conversion_log)
                + len(self.failed_conversions),
                "successful": len(self.conversion_log),
                "failed": len(self.failed_conversions),
            },
        }

        with open(output_file, "w") as f:
            json.dump(log_data, f, indent=2)

        print(f"\nConversion log saved to {output_file}")

    def generate_registration_script(
        self, output_file: str = "register_converted_modules.go"
    ) -> None:
        """Generate a Go file with all module registrations"""
        if not self.conversion_log:
            print("No successful conversions to generate registration script")
            return

        lines = []
        lines.append("package main")
        lines.append("")
        lines.append("// Auto-generated module registrations")
        lines.append("// This file was generated by batch_convert_modules.py")
        lines.append("")
        lines.append("import (")
        lines.append('\t"fmt"')
        lines.append('\t"github.com/LaPingvino/openteacher/internal/core"')

        # Add imports for all converted modules
        for entry in self.conversion_log:
            package_name = entry["module_name"].lower()
            lines.append(
                f'\t"{package_name}" "github.com/LaPingvino/openteacher/internal/modules/{package_name}"'
            )

        lines.append(")")
        lines.append("")
        lines.append(
            "// RegisterAllConvertedModules registers all auto-converted modules"
        )
        lines.append("func RegisterAllConvertedModules(manager *core.Manager) error {")

        for entry in self.conversion_log:
            module_name = entry["module_name"]
            package_name = module_name.lower()
            class_name = module_name.title() + "Module"

            lines.append(f"\t// Register {module_name} module")
            lines.append(f"\t{module_name}Module := {package_name}.New{class_name}()")
            lines.append(
                f"\tif err := manager.Register({module_name}Module); err != nil {{"
            )
            lines.append(
                f'\t\treturn fmt.Errorf("failed to register {module_name} module: %w", err)'
            )
            lines.append(f"\t}}")
            lines.append("")

        lines.append("\treturn nil")
        lines.append("}")

        with open(output_file, "w") as f:
            f.write("\n".join(lines))

        print(f"Generated registration script: {output_file}")

    def analyze_dependencies(self) -> Dict:
        """Analyze dependency relationships between converted modules"""
        if not self.conversion_log:
            return {}

        deps = {}
        for entry in self.conversion_log:
            module_name = entry["module_name"]
            requires = entry["requires"]
            uses = entry["uses"]

            deps[module_name] = {
                "requires": requires,
                "uses": uses,
                "required_by": [],
                "used_by": [],
            }

        # Build reverse dependencies
        for module_name, info in deps.items():
            for req in info["requires"]:
                if req in deps:
                    deps[req]["required_by"].append(module_name)
            for use in info["uses"]:
                if use in deps:
                    deps[use]["used_by"].append(module_name)

        return deps


def main():
    """Main function for batch conversion"""
    parser = argparse.ArgumentParser(
        description="Batch convert OpenTeacher legacy Python modules to Go"
    )
    parser.add_argument(
        "--category",
        help="Convert modules in specific category (e.g., 'interfaces/qt', 'logic')",
    )
    parser.add_argument(
        "--pattern", help="Convert modules matching pattern (e.g., '*Widget*')"
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Convert all legacy modules (use with caution)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be converted without actually converting",
    )
    parser.add_argument(
        "--legacy-root",
        default="legacy",
        help="Root directory of legacy modules (default: legacy)",
    )
    parser.add_argument(
        "--output-root",
        default="internal/modules",
        help="Root directory for output modules (default: internal/modules)",
    )
    parser.add_argument(
        "--log-file",
        default="conversion_log.json",
        help="File to save conversion log (default: conversion_log.json)",
    )

    args = parser.parse_args()

    if not any([args.category, args.pattern, args.all]):
        print("Error: Must specify --category, --pattern, or --all")
        sys.exit(1)

    batch_converter = BatchConverter(args.legacy_root, args.output_root)

    # Find modules to convert
    modules = batch_converter.find_legacy_modules(
        category=args.category, pattern=args.pattern
    )

    if not modules:
        print("No modules found matching criteria")
        sys.exit(1)

    if not args.all and len(modules) > 50:
        print(
            f"Found {len(modules)} modules. Use --all to convert them all, or be more specific with --category or --pattern"
        )
        sys.exit(1)

    # Convert modules
    results = batch_converter.convert_modules(modules, dry_run=args.dry_run)

    # Print summary
    print(f"\n{'=' * 60}")
    print("CONVERSION SUMMARY")
    print(f"{'=' * 60}")
    print(f"Total modules processed: {results['total']}")
    print(f"Successfully converted: {len(results['converted'])}")
    print(f"Failed conversions: {len(results['failed'])}")
    print(f"Skipped: {len(results['skipped'])}")

    if results["failed"]:
        print(f"\nFailed conversions:")
        for failed in results["failed"]:
            print(f"  - {failed}")

    if not args.dry_run and batch_converter.conversion_log:
        # Save conversion log
        batch_converter.save_conversion_log(args.log_file)

        # Generate registration script
        batch_converter.generate_registration_script()

        # Analyze dependencies
        deps = batch_converter.analyze_dependencies()
        print(f"\nDependency Analysis:")
        print(f"Converted {len(deps)} modules with dependencies")

        missing_deps = set()
        for module_name, info in deps.items():
            for req in info["requires"]:
                if req not in deps and req not in [
                    "event",
                    "settings",
                    "metadata",
                    "ui",
                ]:
                    missing_deps.add(req)

        if missing_deps:
            print(f"Missing dependencies that need to be converted:")
            for dep in sorted(missing_deps):
                print(f"  - {dep}")


if __name__ == "__main__":
    main()
