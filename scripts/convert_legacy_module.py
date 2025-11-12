#!/usr/bin/env python3
"""
OpenTeacher Legacy Module Converter

This script automatically converts Python modules from the legacy codebase
to Go modules following the established OpenTeacher module system patterns.

Usage:
    python scripts/convert_legacy_module.py <legacy_module_path> [output_path]

Examples:
    python scripts/convert_legacy_module.py legacy/modules/org/openteacher/interfaces/qt/startWidget/startWidget.py
    python scripts/convert_legacy_module.py legacy/modules/org/openteacher/logic/lesson/lesson.py internal/modules/logic/
"""

import argparse
import ast
import os
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple


@dataclass
class ModuleInfo:
    """Information extracted from a Python module"""

    name: str
    module_type: str
    class_name: str
    requires: List[str]
    uses: List[str]
    priority: Optional[int]
    methods: List[str]
    properties: List[str]
    imports: List[str]
    doc_string: Optional[str]


@dataclass
class ConversionContext:
    """Context for the conversion process"""

    legacy_root: Path
    output_root: Path
    module_registry: Dict[str, ModuleInfo]
    type_mappings: Dict[str, str]
    method_mappings: Dict[str, str]


class PythonASTAnalyzer(ast.NodeVisitor):
    """Analyzes Python AST to extract module information"""

    def __init__(self):
        self.classes = []
        self.functions = []
        self.imports = []
        self.module_info = None
        self.current_class = None

    def visit_Import(self, node):
        for alias in node.names:
            self.imports.append(alias.name)

    def visit_ImportFrom(self, node):
        if node.module:
            for alias in node.names:
                self.imports.append(f"{node.module}.{alias.name}")

    def visit_ClassDef(self, node):
        self.current_class = {
            "name": node.name,
            "methods": [],
            "properties": [],
            "doc": ast.get_docstring(node),
            "bases": [
                base.id if isinstance(base, ast.Name) else str(base)
                for base in node.bases
            ],
        }
        self.classes.append(self.current_class)
        self.generic_visit(node)
        self.current_class = None

    def visit_FunctionDef(self, node):
        if self.current_class:
            self.current_class["methods"].append(node.name)
        else:
            self.functions.append(node.name)

    def visit_Assign(self, node):
        # Look for module metadata assignments
        for target in node.targets:
            if isinstance(target, ast.Name):
                if target.id == "type" and isinstance(node.value, ast.Str):
                    if self.module_info is None:
                        self.module_info = {}
                    self.module_info["type"] = node.value.s
                elif target.id == "requires":
                    if self.module_info is None:
                        self.module_info = {}
                    self.module_info["requires"] = self._extract_list_strings(
                        node.value
                    )
                elif target.id == "uses":
                    if self.module_info is None:
                        self.module_info = {}
                    self.module_info["uses"] = self._extract_list_strings(node.value)

    def _extract_list_strings(self, node):
        """Extract string values from list or tuple nodes"""
        if isinstance(node, (ast.List, ast.Tuple)):
            return [elt.s for elt in node.elts if isinstance(elt, ast.Str)]
        return []


class LegacyModuleConverter:
    """Converts legacy Python modules to Go"""

    def __init__(self, legacy_root: str, output_root: str):
        self.context = ConversionContext(
            legacy_root=Path(legacy_root),
            output_root=Path(output_root),
            module_registry={},
            type_mappings=self._init_type_mappings(),
            method_mappings=self._init_method_mappings(),
        )

    def _init_type_mappings(self) -> Dict[str, str]:
        """Initialize Python to Go type mappings"""
        return {
            "str": "string",
            "int": "int",
            "float": "float64",
            "bool": "bool",
            "list": "[]interface{}",
            "dict": "map[string]interface{}",
            "tuple": "[]interface{}",
            "None": "nil",
            "object": "interface{}",
            "QtCore.QObject": "interface{}",
            "QtWidgets.QWidget": "widgets.QWidget_ITF",
            "QtWidgets.QMainWindow": "*widgets.QMainWindow",
            "QtWidgets.QDialog": "*widgets.QDialog",
        }

    def _init_method_mappings(self) -> Dict[str, str]:
        """Initialize Python to Go method name mappings"""
        return {
            "__init__": "New{ClassName}",
            "__str__": "String",
            "__repr__": "String",
            "__call__": "Call",
            "__iter__": "Iterator",
            "__len__": "Len",
            "__getitem__": "Get",
            "__setitem__": "Set",
            "__contains__": "Contains",
        }

    def analyze_legacy_module(self, file_path: Path) -> ModuleInfo:
        """Analyze a legacy Python module and extract information"""
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        try:
            tree = ast.parse(content)
        except SyntaxError as e:
            print(f"Warning: Could not parse {file_path}: {e}")
            return self._create_fallback_module_info(file_path, content)

        analyzer = PythonASTAnalyzer()
        analyzer.visit(tree)

        # Extract module information
        main_class = None
        if analyzer.classes:
            # Look for the main module class (usually ending with 'Module')
            for cls in analyzer.classes:
                if cls["name"].endswith("Module"):
                    main_class = cls
                    break
            if not main_class:
                main_class = analyzer.classes[0]

        module_name = self._derive_module_name(file_path)
        module_type = self._derive_module_type(
            analyzer.module_info, main_class, file_path
        )

        return ModuleInfo(
            name=module_name,
            module_type=module_type,
            class_name=main_class["name"]
            if main_class
            else self._derive_class_name(module_name),
            requires=analyzer.module_info.get("requires", [])
            if analyzer.module_info
            else [],
            uses=analyzer.module_info.get("uses", []) if analyzer.module_info else [],
            priority=None,  # Will be derived or set to default
            methods=main_class["methods"] if main_class else analyzer.functions,
            properties=[],  # TODO: Extract properties
            imports=analyzer.imports,
            doc_string=main_class["doc"] if main_class else None,
        )

    def _create_fallback_module_info(self, file_path: Path, content: str) -> ModuleInfo:
        """Create module info when AST parsing fails"""
        module_name = self._derive_module_name(file_path)

        # Try to extract basic info with regex
        type_match = re.search(r'self\.type\s*=\s*["\']([^"\']+)["\']', content)
        module_type = type_match.group(1) if type_match else "unknown"

        requires_match = re.search(
            r"self\.requires\s*=\s*\((.*?)\)", content, re.DOTALL
        )
        requires = []
        if requires_match:
            requires_text = requires_match.group(1)
            requires = re.findall(r'["\']([^"\']+)["\']', requires_text)

        return ModuleInfo(
            name=module_name,
            module_type=module_type,
            class_name=self._derive_class_name(module_name),
            requires=requires,
            uses=[],
            priority=None,
            methods=[],
            properties=[],
            imports=[],
            doc_string=f"Package {module_name} provides functionality ported from Python module\n{file_path}\n\nThis is an automated port - implementation may be incomplete.",
        )

    def _derive_module_name(self, file_path: Path) -> str:
        """Derive Go module name from file path"""
        # Get the last directory name and file name
        parent_dir = file_path.parent.name
        file_stem = file_path.stem

        # If they're the same, just use one
        if parent_dir == file_stem:
            return self._to_go_name(file_stem)

        return self._to_go_name(f"{parent_dir}_{file_stem}")

    def _derive_module_type(
        self, module_info: Optional[Dict], main_class: Optional[Dict], file_path: Path
    ) -> str:
        """Derive module type from various sources"""
        if module_info and "type" in module_info:
            return module_info["type"]

        # Derive from path
        path_parts = file_path.parts
        if "interfaces" in path_parts:
            if "qt" in path_parts:
                return "ui"
            return "interface"
        elif "logic" in path_parts:
            return "logic"
        elif "data" in path_parts:
            return "data"
        elif "profileRunners" in path_parts:
            return file_path.parent.name  # Use directory name as type

        return "unknown"

    def _derive_class_name(self, module_name: str) -> str:
        """Derive Go class name from module name"""
        return self._to_go_class_name(module_name) + "Module"

    def _to_go_name(self, name: str) -> str:
        """Convert Python name to Go name (camelCase)"""
        # Handle special cases
        name = re.sub(
            r"[^a-zA-Z0-9_]", "", name
        )  # Remove non-alphanumeric except underscore
        parts = name.split("_")
        if len(parts) == 1:
            return parts[0].lower()
        return parts[0].lower() + "".join(word.capitalize() for word in parts[1:])

    def _to_go_class_name(self, name: str) -> str:
        """Convert Python name to Go class name (PascalCase)"""
        name = re.sub(r"[^a-zA-Z0-9_]", "", name)
        parts = name.split("_")
        return "".join(word.capitalize() for word in parts)

    def _to_go_method_name(self, py_method: str) -> str:
        """Convert Python method name to Go method name"""
        if py_method in self.context.method_mappings:
            return self.context.method_mappings[py_method]

        # Handle special Python methods
        if py_method.startswith("__") and py_method.endswith("__"):
            method_name = py_method[2:-2]
            return self._to_go_class_name(method_name)

        # Handle private methods (start with _)
        if py_method.startswith("_"):
            return self._to_go_name(py_method[1:])  # Make it private by lowercase

        # Convert to Go convention (PascalCase for public methods)
        return self._to_go_class_name(py_method)

    def generate_go_module(self, module_info: ModuleInfo, output_path: Path) -> None:
        """Generate Go module code from module information"""

        # Ensure output directory exists
        output_path.parent.mkdir(parents=True, exist_ok=True)

        go_code = self._generate_go_code(module_info)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(go_code)

        print(f"Generated Go module: {output_path}")

    def _generate_go_code(self, info: ModuleInfo) -> str:
        """Generate the complete Go module code"""

        lines = []

        # Package declaration and doc comment
        lines.append(
            f"// Package {info.name} provides functionality ported from Python module"
        )
        if info.doc_string:
            lines.append("//")
            for line in info.doc_string.split("\n"):
                lines.append(f"// {line}" if line.strip() else "//")
        lines.append("//")
        lines.append("// This is an automated port - implementation may be incomplete.")
        lines.append(f"package {self._get_package_name(info.name)}")
        lines.append("")

        # Imports
        lines.extend(self._generate_imports(info))
        lines.append("")

        # Main struct
        lines.extend(self._generate_struct(info))
        lines.append("")

        # Constructor
        lines.extend(self._generate_constructor(info))
        lines.append("")

        # Methods
        for method in info.methods:
            if method not in ["__init__", "enable", "disable"]:
                lines.extend(self._generate_method(info, method))
                lines.append("")

        # Standard module methods
        lines.extend(self._generate_enable_method(info))
        lines.append("")
        lines.extend(self._generate_disable_method(info))
        lines.append("")
        lines.extend(self._generate_set_manager_method(info))
        lines.append("")

        # Init function
        lines.extend(self._generate_init_function(info))

        return "\n".join(lines)

    def _get_package_name(self, module_name: str) -> str:
        """Get the Go package name from module name"""
        # For now, use a simple approach
        return module_name.lower().replace("_", "").replace("-", "")

    def _generate_imports(self, info: ModuleInfo) -> List[str]:
        """Generate import statements"""
        lines = []
        lines.append("import (")
        lines.append('\t"context"')
        lines.append('\t"fmt"')
        lines.append('\t"github.com/LaPingvino/openteacher/internal/core"')

        # Add Qt imports if this appears to be a UI module
        if info.module_type == "ui" or any("Qt" in imp for imp in info.imports):
            lines.append('\t"github.com/therecipe/qt/widgets"')

        lines.append(")")
        return lines

    def _generate_struct(self, info: ModuleInfo) -> List[str]:
        """Generate the main struct definition"""
        lines = []
        lines.append(
            f"// {info.class_name} is a Go port of the Python {info.class_name} class"
        )
        lines.append(f"type {info.class_name} struct {{")
        lines.append("\t*core.BaseModule")
        lines.append("\tmanager *core.Manager")
        lines.append("\t// TODO: Add module-specific fields")
        lines.append("}")
        return lines

    def _generate_constructor(self, info: ModuleInfo) -> List[str]:
        """Generate the constructor function"""
        lines = []
        constructor_name = f"New{info.class_name}"
        lines.append(f"// {constructor_name} creates a new {info.class_name} instance")
        lines.append(f"func {constructor_name}() *{info.class_name} {{")

        # Determine module type and name
        module_type = info.module_type if info.module_type != "unknown" else "unknown"
        module_name = f"{info.name}-module"

        lines.append(f'\tbase := core.NewBaseModule("{module_type}", "{module_name}")')

        # Add requirements if any
        if info.requires:
            req_list = ", ".join(f'"{req}"' for req in info.requires)
            lines.append(f"\tbase.SetRequires({req_list})")

        # Add uses if any
        if info.uses:
            uses_list = ", ".join(f'"{use}"' for use in info.uses)
            lines.append(f"\tbase.SetUses({uses_list})")

        lines.append("")
        lines.append(f"\treturn &{info.class_name}{{")
        lines.append("\t\tBaseModule: base,")
        lines.append("\t}")
        lines.append("}")
        return lines

    def _generate_method(self, info: ModuleInfo, method_name: str) -> List[str]:
        """Generate a method from Python method"""
        lines = []
        go_method_name = self._to_go_method_name(method_name)

        lines.append(
            f"// {go_method_name} is the Go port of the Python {method_name} method"
        )
        lines.append(f"func (mod *{info.class_name}) {go_method_name}() {{")
        lines.append("\t// TODO: Port Python method logic")
        lines.append("}")
        return lines

    def _generate_enable_method(self, info: ModuleInfo) -> List[str]:
        """Generate the Enable method"""
        lines = []
        lines.append("// Enable activates the module")
        lines.append("// This is the Go equivalent of the Python enable method")
        lines.append(
            f"func (mod *{info.class_name}) Enable(ctx context.Context) error {{"
        )
        lines.append("\tif err := mod.BaseModule.Enable(ctx); err != nil {")
        lines.append("\t\treturn err")
        lines.append("\t}")
        lines.append("")
        lines.append("\t// TODO: Port Python enable logic")
        lines.append("")
        lines.append(f'\tfmt.Println("{info.class_name} enabled")')
        lines.append("\treturn nil")
        lines.append("}")
        return lines

    def _generate_disable_method(self, info: ModuleInfo) -> List[str]:
        """Generate the Disable method"""
        lines = []
        lines.append("// Disable deactivates the module")
        lines.append("// This is the Go equivalent of the Python disable method")
        lines.append(
            f"func (mod *{info.class_name}) Disable(ctx context.Context) error {{"
        )
        lines.append("\tif err := mod.BaseModule.Disable(ctx); err != nil {")
        lines.append("\t\treturn err")
        lines.append("\t}")
        lines.append("")
        lines.append("\t// TODO: Port Python disable logic")
        lines.append("")
        lines.append(f'\tfmt.Println("{info.class_name} disabled")')
        lines.append("\treturn nil")
        lines.append("}")
        return lines

    def _generate_set_manager_method(self, info: ModuleInfo) -> List[str]:
        """Generate the SetManager method"""
        lines = []
        lines.append("// SetManager sets the module manager")
        lines.append(
            f"func (mod *{info.class_name}) SetManager(manager *core.Manager) {{"
        )
        lines.append("\tmod.manager = manager")
        lines.append("}")
        return lines

    def _generate_init_function(self, info: ModuleInfo) -> List[str]:
        """Generate the Init function"""
        lines = []
        init_name = f"Init{info.class_name}"
        lines.append(
            f"// {init_name} creates and returns a new {info.class_name} instance"
        )
        lines.append("// This is the Go equivalent of the Python init function")
        lines.append(f"func {init_name}() core.Module {{")
        lines.append(f"\treturn New{info.class_name}()")
        lines.append("}")
        return lines

    def convert_module(self, legacy_file: str, output_file: str = None) -> None:
        """Convert a legacy Python module to Go"""

        legacy_path = Path(legacy_file)
        if not legacy_path.exists():
            raise FileNotFoundError(f"Legacy module not found: {legacy_file}")

        # Analyze the legacy module
        print(f"Analyzing legacy module: {legacy_path}")
        module_info = self.analyze_legacy_module(legacy_path)

        # Determine output path
        if output_file:
            output_path = Path(output_file)
        else:
            # Generate output path based on module structure
            relative_path = legacy_path.relative_to(self.context.legacy_root)
            # Convert path: remove 'modules/org/openteacher', change .py to .go
            path_parts = relative_path.parts
            if len(path_parts) >= 3 and path_parts[:3] == (
                "modules",
                "org",
                "openteacher",
            ):
                new_parts = path_parts[3:]  # Remove 'modules/org/openteacher'
            else:
                new_parts = path_parts

            # Create output path
            output_path = (
                self.context.output_root
                / Path(*new_parts[:-1])
                / f"{new_parts[-1].replace('.py', '.go')}"
            )

        # Generate Go code
        print(f"Generating Go module: {output_path}")
        self.generate_go_module(module_info, output_path)

        # Automatically register the module in main.go
        self.register_module_in_main(module_info, output_path)

        return module_info

    def register_module_in_main(
        self, module_info: ModuleInfo, output_path: Path
    ) -> None:
        """Automatically register the generated module in cmd/openteacher/main.go"""
        main_go_path = Path("cmd/openteacher/main.go")

        if not main_go_path.exists():
            print(
                f"Warning: main.go not found at {main_go_path}, skipping registration"
            )
            return

        # Read the main.go file
        with open(main_go_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Determine import path for the module
        import_path = self._get_import_path(output_path)
        constructor_name = f"New{module_info.class_name}"
        variable_name = self._to_go_name(module_info.name) + "Module"

        # Check if already imported/registered
        if import_path in content and constructor_name in content:
            print(
                f"Module {module_info.name} appears to already be registered in main.go"
            )
            return

        # Add import if needed
        if import_path not in content:
            content = self._add_import_to_main(content, import_path)

        # Add module registration
        content = self._add_module_registration(
            content, module_info, constructor_name, variable_name
        )

        # Write back to file
        with open(main_go_path, "w", encoding="utf-8") as f:
            f.write(content)

        print(f"Registered module {module_info.name} in main.go")

    def _get_import_path(self, output_path: Path) -> str:
        """Generate the import path for a Go module"""
        # Convert file path to Go import path
        path_parts = output_path.parts

        # Find the internal/modules part
        if "internal" in path_parts and "modules" in path_parts:
            internal_index = path_parts.index("internal")
            return "github.com/LaPingvino/openteacher/" + "/".join(
                path_parts[internal_index:-1]
            )

        return f"github.com/LaPingvino/openteacher/internal/modules/{output_path.parent.name}"

    def _add_import_to_main(self, content: str, import_path: str) -> str:
        """Add an import statement to main.go"""
        # Find the import block
        import_pattern = r"(\nimport \(\n)(.*?\n)(\))"
        match = re.search(import_pattern, content, re.DOTALL)

        if match:
            # Add the import to the existing import block
            import_start, imports, import_end = match.groups()
            new_import = f'\t"{import_path}"\n'
            return content.replace(
                match.group(0), f"{import_start}{imports}{new_import}{import_end}"
            )
        else:
            # Try to add after existing single imports
            single_import_pattern = r'(\nimport\s+"[^"]+"\n)'
            if re.search(single_import_pattern, content):
                return re.sub(
                    r'(\nimport\s+"[^"]+"\n)',
                    f'\\1import "{import_path}"\n',
                    content,
                    count=1,
                )
            else:
                # Add import block after package declaration
                return re.sub(
                    r"(package main\n)",
                    f'\\1\nimport (\n\t"{import_path}"\n)\n',
                    content,
                )

    def _add_module_registration(
        self,
        content: str,
        module_info: ModuleInfo,
        constructor_name: str,
        variable_name: str,
    ) -> str:
        """Add module registration to main.go"""
        # Find the registerAllModules function
        register_func_pattern = r"(func registerAllModules\(manager \*core\.Manager\) error \{.*?)([ \t]*return nil\n\})"
        match = re.search(register_func_pattern, content, re.DOTALL)

        if match:
            func_body, func_end = match.groups()

            # Create registration code
            registration_code = f"""
	// Register {module_info.name} module
	{variable_name} := {module_info.name}.{constructor_name}()
	if err := manager.Register({variable_name}); err != nil {{
		return fmt.Errorf("failed to register {module_info.name} module: %w", err)
	}}
"""

            return content.replace(
                match.group(0), f"{func_body}{registration_code}\n{func_end}"
            )
        else:
            print(f"Warning: Could not find registerAllModules function in main.go")
            return content


def main():
    """Main function for the conversion script"""
    parser = argparse.ArgumentParser(
        description="Convert OpenTeacher legacy Python modules to Go"
    )
    parser.add_argument("legacy_module", help="Path to the legacy Python module file")
    parser.add_argument(
        "output_file",
        nargs="?",
        help="Output Go file path (optional, will be auto-generated if not provided)",
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

    args = parser.parse_args()

    try:
        converter = LegacyModuleConverter(args.legacy_root, args.output_root)
        module_info = converter.convert_module(args.legacy_module, args.output_file)

        print("\nConversion Summary:")
        print(f"  Module Name: {module_info.name}")
        print(f"  Module Type: {module_info.module_type}")
        print(f"  Class Name: {module_info.class_name}")
        print(f"  Requires: {module_info.requires}")
        print(f"  Uses: {module_info.uses}")
        print(f"  Methods: {len(module_info.methods)}")

        if module_info.requires:
            print("\nNote: This module requires the following types:")
            for req in module_info.requires:
                print(f"  - {req}")
            print("Make sure these modules are available and registered.")

    except Exception as e:
        print(f"Error during conversion: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
