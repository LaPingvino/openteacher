#!/bin/bash

# Fix MIQT migration issues
# This script fixes the problems from the initial migration

set -e

echo "🔧 Fixing MIQT migration issues..."

# Function to fix a single Go file
fix_file() {
    local file="$1"
    echo "  Fixing $file"

    # Create backup
    cp "$file" "$file.backup"

    # Remove duplicate qt imports and fix import section
    awk '
    BEGIN { in_import = 0; qt_imported = 0; }
    /^import \(/ { in_import = 1; print; next }
    /^\)/ && in_import {
        if (!qt_imported) print "\t\"github.com/mappu/miqt/qt\""
        in_import = 0; print; next
    }
    in_import && /github\.com\/mappu\/miqt\/qt/ {
        if (!qt_imported) {
            print "\t\"github.com/mappu/miqt/qt\""
            qt_imported = 1
        }
        next
    }
    { print }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

    # Fix type references - remove widgets. prefixes since everything is in qt package
    sed -i \
        -e 's/\*widgets\./\*qt\./g' \
        -e 's/widgets\.New/qt.New/g' \
        -e 's/widgets\.Q/qt.Q/g' \
        -e 's/gui\.New/qt.New/g' \
        -e 's/gui\.Q/qt.Q/g' \
        -e 's/qtcore\./qt./g' \
        -e 's/qtgui\./qt./g' \
        "$file"

    # Fix Qt constants - miqt uses different naming
    sed -i \
        -e 's/qt\.Qt__ApplicationModal/qt.ApplicationModal/g' \
        -e 's/qt\.Qt__AlignHCenter/qt.AlignHCenter/g' \
        -e 's/qt\.Qt__AlignVCenter/qt.AlignVCenter/g' \
        -e 's/qt\.Qt__AA_UseHighDpiPixmaps/qt.AA_UseHighDpiPixmaps/g' \
        -e 's/qt\.QEventLoop__AllEvents/qt.AllEvents/g' \
        "$file"
}

# Find all Go files that likely need fixing
echo "📁 Finding Go files to fix..."
files_to_fix=$(find ./internal/modules/interfaces/qt -name "*.go" -type f)

if [ -z "$files_to_fix" ]; then
    echo "✅ No files found to fix"
    exit 0
fi

echo "📋 Found $(echo "$files_to_fix" | wc -l) files to fix"

# Fix each file
for file in $files_to_fix; do
    if grep -q "github.com/mappu/miqt/qt" "$file"; then
        fix_file "$file"
    fi
done

echo ""
echo "🧪 Testing build..."
if go build ./cmd/recuerdo > /dev/null 2>&1; then
    echo "✅ Build successful!"

    echo ""
    echo "🧹 Cleaning up backup files..."
    find . -name "*.go.backup" -delete

    echo ""
    echo "✅ Migration fixes completed successfully!"
    echo ""
    echo "🎯 The application should now work with Wayland! Try running:"
    echo "   ./recuerdo"
    echo "   # or with explicit Wayland support:"
    echo "   QT_QPA_PLATFORM=wayland ./recuerdo"
else
    echo "❌ Build still has errors. Manual fixes needed."
    echo "Backup files preserved with .backup extension"
    echo ""
    echo "Running go build to show errors:"
    go build ./cmd/recuerdo
fi
