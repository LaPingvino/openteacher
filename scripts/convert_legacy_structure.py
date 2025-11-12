#!/usr/bin/env python3
"""
Automated Legacy Structure Converter

This script scans all Python files in the legacy/ directory and creates
corresponding Go file placeholders in the internal/ directory structure.

It preserves the module hierarchy and creates basic Go implementations
with the same class/function structure as placeholders.
"""

import ast
import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple


class PythonModuleAnalyzer:
    """Analyzes Python modules and extracts structural information."""

    def __init__(self, file_path: str):
        self.file_path = file_path
        self.classes = []
        self.functions = []
        self.imports = []
        self.module_type = None
        self.dependencies = []

    def analyze(self) -> Dict:
        """Analyze the Python file and extract structural information."""
        try:
            with open(self.file_path, "r", encoding="utf-8") as f:
                content = f.read()

            tree = ast.parse(content)
            self._extract_info(tree)

            # Try to detect module type from common patterns
            self._detect_module_type(content)

            return {
                "classes": self.classes,
                "functions": self.functions,
                "imports": self.imports,
                "module_type": self.module_type,
                "dependencies": self.dependencies,
            }
        except Exception as e:
            print(f"Warning: Could not parse {self.file_path}: {e}")
            return {
                "classes": [],
                "functions": [],
                "imports": [],
                "module_type": "unknown",
                "dependencies": [],
            }

    def _extract_info(self, tree):
        """Extract classes, functions, and imports from AST."""
        for node in ast.walk(tree):
            if isinstance(node, ast.ClassDef):
                self.classes.append(
                    {
                        "name": node.name,
                        "methods": [
                            n.name for n in node.body if isinstance(n, ast.FunctionDef)
                        ],
                        "bases": [self._get_name(base) for base in node.bases],
                    }
                )
            elif isinstance(node, ast.FunctionDef) and not self._is_method(node):
                self.functions.append(node.name)
            elif isinstance(node, ast.Import):
                self.imports.extend([alias.name for alias in node.names])
            elif isinstance(node, ast.ImportFrom):
                if node.module:
                    self.imports.append(node.module)

    def _get_name(self, node) -> str:
        """Extract name from AST node."""
        if isinstance(node, ast.Name):
            return node.id
        elif isinstance(node, ast.Attribute):
            return f"{self._get_name(node.value)}.{node.attr}"
        return str(node)

    def _is_method(self, node) -> bool:
        """Check if function is inside a class (method)."""
        # This is a simplified check - in a full implementation we'd track scope
        return False

    def _detect_module_type(self, content: str):
        """Detect module type based on content patterns."""
        if "class" in content and "Module" in content:
            if "type = " in content:
                # Try to extract the module type
                type_match = re.search(r'self\.type = ["\']([^"\']+)["\']', content)
                if type_match:
                    self.module_type = type_match.group(1)
                    return

            # Fallback to class name pattern
            class_match = re.search(r"class (\w+)Module", content)
            if class_match:
                name = class_match.group(1)
                # Convert CamelCase to snake_case
                self.module_type = re.sub(r"([A-Z])", r"_\1", name).lower().lstrip("_")

        if not self.module_type:
            self.module_type = "unknown"


class GoFileGenerator:
    """Generates Go file content from Python module analysis."""

    def __init__(self, python_path: str, analysis: Dict, go_path: str):
        self.python_path = python_path
        self.analysis = analysis
        self.go_path = go_path
        self.package_name = self._extract_package_name(go_path)

    def _extract_package_name(self, go_path: str) -> str:
        """Extract Go package name from path."""
        return Path(go_path).name.lower()

    def generate(self) -> str:
        """Generate Go file content."""
        content = []

        # Package declaration
        content.append(
            f"// Package {self.package_name} provides functionality ported from Python module"
        )
        content.append(f"// {self.python_path}")
        content.append(f"//")
        content.append(
            f"// This is an automated port - implementation may be incomplete."
        )
        content.append(f"package {self.package_name}")
        content.append("")

        # Standard imports
        imports = self._generate_imports()
        if imports:
            content.append("import (")
            for imp in imports:
                content.append(f"\t{imp}")
            content.append(")")
            content.append("")

        # Generate types and structs for Python classes
        for cls in self.analysis["classes"]:
            content.extend(self._generate_class(cls))
            content.append("")

        # Generate functions
        for func in self.analysis["functions"]:
            content.extend(self._generate_function(func))
            content.append("")

        # Generate init function if this looks like a module
        if self.analysis["classes"] and any(
            "Module" in cls["name"] for cls in self.analysis["classes"]
        ):
            content.extend(self._generate_init_function())

        return "\n".join(content)

    def _generate_imports(self) -> List[str]:
        """Generate Go imports based on detected patterns."""
        imports = [
            '"context"',
            '"fmt"',
        ]

        # Add OpenTeacher core imports if this looks like a module
        if any("Module" in cls["name"] for cls in self.analysis["classes"]):
            imports.append('"github.com/LaPingvino/openteacher/internal/core"')

        return imports

    def _generate_class(self, cls: Dict) -> List[str]:
        """Generate Go struct and methods for Python class."""
        lines = []

        # Generate struct
        struct_name = cls["name"]
        lines.append(f"// {struct_name} is a Go port of the Python {cls['name']} class")
        lines.append(f"type {struct_name} struct {{")

        # If this looks like an OpenTeacher module, embed BaseModule
        if "Module" in cls["name"]:
            lines.append("\t*core.BaseModule")
            lines.append("\tmanager *core.Manager")
            lines.append("\t// TODO: Add module-specific fields")
        else:
            lines.append("\t// TODO: Add struct fields based on Python class")

        lines.append("}")
        lines.append("")

        # Generate constructor
        constructor_name = f"New{struct_name}"
        lines.append(f"// {constructor_name} creates a new {struct_name} instance")
        if "Module" in cls["name"]:
            lines.append(f"func {constructor_name}() *{struct_name} {{")

            # Try to extract module type
            module_type = self.analysis.get("module_type", "unknown")
            if module_type and module_type != "unknown":
                lines.append(
                    f'\tbase := core.NewBaseModule("{module_type}", "{module_type.replace("_", "-")}")'
                )
            else:
                lines.append(
                    '\tbase := core.NewBaseModule("unknown", "unknown-module")'
                )

            lines.append("")
            lines.append(f"\treturn &{struct_name}{{")
            lines.append("\t\tBaseModule: base,")
            lines.append("\t}")
        else:
            lines.append(f"func {constructor_name}() *{struct_name} {{")
            lines.append(f"\treturn &{struct_name}{{")
            lines.append("\t\t// TODO: Initialize fields")
            lines.append("\t}")
        lines.append("}")
        lines.append("")

        # Generate methods
        for method in cls["methods"]:
            lines.extend(self._generate_method(struct_name, method))
            lines.append("")

        # Generate standard module methods if this is a module
        if "Module" in cls["name"]:
            lines.extend(self._generate_module_methods(struct_name))

        return lines

    def _generate_method(self, struct_name: str, method_name: str) -> List[str]:
        """Generate Go method from Python method."""
        lines = []

        receiver = struct_name.lower()[:3]  # Use first 3 letters as receiver

        # Handle special method names
        go_method_name = method_name
        if method_name == "enable":
            go_method_name = "Enable"
            lines.append(
                f"// {go_method_name} is the Go port of the Python {method_name} method"
            )
            lines.append(
                f"func ({receiver} *{struct_name}) {go_method_name}(ctx context.Context) error {{"
            )
            lines.append("\t// TODO: Port Python enable logic")
            lines.append("\treturn nil")
        elif method_name == "disable":
            go_method_name = "Disable"
            lines.append(
                f"// {go_method_name} is the Go port of the Python {method_name} method"
            )
            lines.append(
                f"func ({receiver} *{struct_name}) {go_method_name}(ctx context.Context) error {{"
            )
            lines.append("\t// TODO: Port Python disable logic")
            lines.append("\treturn nil")
        elif method_name.startswith("_") and not method_name.startswith("__"):
            # Private method - make it unexported in Go
            go_method_name = method_name[1:]  # Remove underscore
            lines.append(
                f"// {go_method_name} is the Go port of the Python {method_name} method"
            )
            lines.append(f"func ({receiver} *{struct_name}) {go_method_name}() {{")
            lines.append("\t// TODO: Port Python private method logic")
        else:
            # Capitalize first letter for public method
            go_method_name = method_name[0].upper() + method_name[1:]
            lines.append(
                f"// {go_method_name} is the Go port of the Python {method_name} method"
            )
            lines.append(f"func ({receiver} *{struct_name}) {go_method_name}() {{")
            lines.append("\t// TODO: Port Python method logic")

        lines.append("}")

        return lines

    def _generate_module_methods(self, struct_name: str) -> List[str]:
        """Generate standard OpenTeacher module methods."""
        lines = []
        receiver = struct_name.lower()[:3]

        # SetManager method
        lines.append("// SetManager sets the module manager")
        lines.append(
            f"func ({receiver} *{struct_name}) SetManager(manager *core.Manager) {{"
        )
        lines.append(f"\t{receiver}.manager = manager")
        lines.append("}")
        lines.append("")

        return lines

    def _generate_function(self, func_name: str) -> List[str]:
        """Generate Go function from Python function."""
        lines = []

        # Capitalize first letter for public function
        go_func_name = func_name[0].upper() + func_name[1:]

        lines.append(
            f"// {go_func_name} is the Go port of the Python {func_name} function"
        )
        lines.append(f"func {go_func_name}() {{")
        lines.append("\t// TODO: Port Python function logic")
        lines.append("}")

        return lines

    def _generate_init_function(self) -> List[str]:
        """Generate init function for module."""
        lines = []

        lines.append("// Init creates and returns a new module instance")
        lines.append("// This is the Go equivalent of the Python init function")
        lines.append("func Init() core.Module {")

        # Find the main module class
        module_class = None
        for cls in self.analysis["classes"]:
            if "Module" in cls["name"]:
                module_class = cls["name"]
                break

        if module_class:
            lines.append(f"\treturn New{module_class}()")
        else:
            lines.append("\t// TODO: Return appropriate module instance")
            lines.append("\treturn nil")

        lines.append("}")

        return lines


class StructureConverter:
    """Main converter that orchestrates the entire conversion process."""

    def __init__(self, legacy_root: str, target_root: str):
        self.legacy_root = Path(legacy_root)
        self.target_root = Path(target_root)
        self.converted_files = []
        self.skipped_files = []

    def convert_all(self):
        """Convert all Python files in the legacy structure."""
        print(f"Starting conversion from {self.legacy_root} to {self.target_root}")

        # Find all Python files
        python_files = list(self.legacy_root.rglob("*.py"))
        print(f"Found {len(python_files)} Python files to convert")

        for py_file in python_files:
            try:
                self.convert_file(py_file)
            except Exception as e:
                print(f"Error converting {py_file}: {e}")
                self.skipped_files.append(str(py_file))

        self.print_summary()

    def convert_file(self, py_file: Path):
        """Convert a single Python file to Go."""
        # Skip __init__.py files
        if py_file.name == "__init__.py":
            return

        # Calculate target Go file path
        go_path = self._calculate_go_path(py_file)

        # Skip if Go file already exists and is substantial
        if go_path.exists() and go_path.stat().st_size > 1000:
            print(f"Skipping {py_file} - Go file already exists: {go_path}")
            return

        # Analyze Python file
        analyzer = PythonModuleAnalyzer(str(py_file))
        analysis = analyzer.analyze()

        # Generate Go file
        generator = GoFileGenerator(str(py_file), analysis, str(go_path))
        go_content = generator.generate()

        # Ensure target directory exists
        go_path.parent.mkdir(parents=True, exist_ok=True)

        # Write Go file
        with open(go_path, "w", encoding="utf-8") as f:
            f.write(go_content)

        self.converted_files.append((str(py_file), str(go_path)))
        print(f"Converted: {py_file} -> {go_path}")

    def _calculate_go_path(self, py_file: Path) -> Path:
        """Calculate the target Go file path from Python file path."""
        # Get relative path from legacy root
        rel_path = py_file.relative_to(self.legacy_root)

        # Transform path components
        parts = list(rel_path.parts)

        # Skip common Python package structure parts
        if parts and parts[0] == "modules":
            parts = parts[1:]  # Remove 'modules'
        if len(parts) >= 3 and parts[0:3] == ["org", "openteacher"]:
            parts = parts[3:]  # Remove 'org/openteacher'

        # Convert Python file to Go file
        if parts:
            parts[-1] = parts[-1].replace(".py", ".go")

        # Construct target path
        go_path = self.target_root / "internal" / "modules" / Path(*parts)

        return go_path

    def print_summary(self):
        """Print conversion summary."""
        print("\n" + "=" * 60)
        print("CONVERSION SUMMARY")
        print("=" * 60)
        print(f"Successfully converted: {len(self.converted_files)} files")
        print(f"Skipped files: {len(self.skipped_files)} files")

        if self.converted_files:
            print("\nConverted files:")
            for py_file, go_file in self.converted_files[-10:]:  # Show last 10
                print(f"  {py_file}")
                print(f"  -> {go_file}")
                print()

        if self.skipped_files:
            print(f"\nSkipped files: {self.skipped_files[:5]}...")  # Show first 5

        print("\nNext steps:")
        print("1. Review generated Go files")
        print("2. Update imports and dependencies")
        print("3. Implement TODO items")
        print("4. Add proper error handling")
        print("5. Write tests for new modules")


def main():
    """Main entry point."""
    import sys

    if len(sys.argv) != 3:
        print("Usage: python convert_legacy_structure.py <legacy_root> <target_root>")
        print("Example: python convert_legacy_structure.py legacy .")
        sys.exit(1)

    legacy_root = sys.argv[1]
    target_root = sys.argv[2]

    if not os.path.exists(legacy_root):
        print(f"Error: Legacy root directory does not exist: {legacy_root}")
        sys.exit(1)

    converter = StructureConverter(legacy_root, target_root)
    converter.convert_all()


if __name__ == "__main__":
    main()
