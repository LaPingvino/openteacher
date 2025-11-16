#!/bin/bash

# Script to remove unused Qt imports from stub modules
# This will help reduce compilation errors during the miqt migration

set -e

echo "🧹 Removing unused Qt imports from stub modules..."

# Function to check if a file actually uses Qt and remove import if not
clean_unused_qt_import() {
    local file="$1"
    echo "  Checking $file"

    # Create backup
    cp "$file" "$file.backup"

    # Check if file actually uses qt. (methods, types, etc.)
    if ! grep -q "qt\." "$file"; then
        echo "    → Removing unused qt import from $file"
        # Remove the qt import line
        sed -i '/^[[:space:]]*"github\.com\/mappu\/miqt\/qt"$/d' "$file"

        # Clean up any empty import blocks
        awk '
        BEGIN { in_import = 0; import_lines = ""; other_imports = 0 }
        /^import \(/ { in_import = 1; print; next }
        /^\)/ && in_import {
            if (other_imports > 0) print
            in_import = 0; other_imports = 0; next
        }
        in_import {
            if ($0 !~ /^[[:space:]]*$/ && $0 !~ /github\.com\/mappu\/miqt\/qt/) {
                other_imports = 1
            }
            if ($0 !~ /github\.com\/mappu\/miqt\/qt/) print
            next
        }
        { print }
        ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    else
        echo "    → Qt import needed in $file"
    fi
}

# List of directories that likely contain stub modules with unused imports
stub_dirs=(
    "internal/modules/interfaces/qt/mediaTypes"
    "internal/modules/interfaces/qt/settingsWidget"
    "internal/modules/interfaces/qt/teachTypes"
    "internal/modules/interfaces/qt/print"
    "internal/modules/interfaces/qt/testMode"
    "internal/modules/interfaces/qt/typingTutor"
    "internal/modules/interfaces/qt/teachers"
)

# Additional individual files that might have unused imports
individual_files=(
    "internal/modules/interfaces/qt/mediaDisplay/mediaDisplay.go"
    "internal/modules/interfaces/qt/loaderGui/loaderGui.go"
    "internal/modules/interfaces/qt/ocrGui/ocrGui.go"
    "internal/modules/interfaces/qt/percentNotesViewer/percentNotesViewer.go"
    "internal/modules/interfaces/qt/printer/printer.go"
    "internal/modules/interfaces/qt/progressViewer/progressViewer.go"
    "internal/modules/interfaces/qt/recentlyOpenedViewer/recentlyOpenedViewer.go"
    "internal/modules/interfaces/qt/settingsWidgets/settingsWidgets.go"
    "internal/modules/interfaces/qt/theme/theme.go"
    "internal/modules/interfaces/qt/topoMaps/topoMaps.go"
)

# Process directories
for dir in "${stub_dirs[@]}"; do
    if [ -d "$dir" ]; then
        echo "Processing directory: $dir"
        find "$dir" -name "*.go" -type f | while read -r file; do
            if grep -q "github.com/mappu/miqt/qt" "$file"; then
                clean_unused_qt_import "$file"
            fi
        done
    fi
done

# Process individual files
echo "Processing individual stub files..."
for file in "${individual_files[@]}"; do
    if [ -f "$file" ] && grep -q "github.com/mappu/miqt/qt" "$file"; then
        clean_unused_qt_import "$file"
    fi
done

echo ""
echo "🧪 Testing build after cleanup..."
if go build ./cmd/recuerdo > build_output.log 2>&1; then
    echo "✅ Build successful after removing unused imports!"

    echo ""
    echo "🧹 Cleaning up backup files..."
    find . -name "*.go.backup" -delete
    rm -f build_output.log

    echo ""
    echo "✅ Unused Qt import cleanup completed successfully!"
else
    echo "❌ Build still has errors. Check build_output.log for details."
    echo "Backup files preserved with .backup extension"

    echo ""
    echo "📋 Showing first few build errors:"
    head -20 build_output.log
fi
