#!/bin/bash

# Migration script to update from therecipe/qt to miqt
# This script updates all Go files to use miqt instead of therecipe/qt

set -e

echo "🚀 Starting migration from therecipe/qt to miqt..."

# Function to update imports in a file
update_file() {
    local file="$1"
    echo "  Updating $file"

    # Create a backup
    cp "$file" "$file.backup"

    # Replace therecipe/qt imports with miqt equivalents (using qt package for both Qt5/Qt6)
    sed -i \
        -e 's|"github.com/therecipe/qt/core"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/gui"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/widgets"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/network"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/multimedia"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/webengine"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/webchannel"|"github.com/mappu/miqt/qt"|g' \
        -e 's|"github.com/therecipe/qt/printsupport"|"github.com/mappu/miqt/qt"|g' \
        "$file"

    # Handle aliased imports - remove qtcore/qtgui aliases since miqt/qt has everything
    sed -i \
        -e 's|qtcore "github.com/mappu/miqt/qt"|"github.com/mappu/miqt/qt"|g' \
        -e 's|qtgui "github.com/mappu/miqt/qt"|"github.com/mappu/miqt/qt"|g' \
        -e 's|qtcore\.|qt.|g' \
        -e 's|qtgui\.|qt.|g' \
        "$file"
}

# Find all Go files that import therecipe/qt
echo "📁 Finding Go files with therecipe/qt imports..."
files_to_update=$(find . -name "*.go" -exec grep -l "github.com/therecipe/qt" {} \;)

if [ -z "$files_to_update" ]; then
    echo "✅ No files found with therecipe/qt imports"
    exit 0
fi

echo "📋 Found $(echo "$files_to_update" | wc -l) files to update:"
echo "$files_to_update"
echo

# Update each file
for file in $files_to_update; do
    update_file "$file"
done

echo
echo "🔧 Updating go.mod..."
go mod tidy

echo
echo "🗑️  Removing qtbox (no longer needed with miqt)..."
if [ -f "qtbox" ]; then
    rm -f qtbox
    echo "  Removed qtbox file"
fi

echo
echo "🧹 Cleaning up backup files..."
find . -name "*.go.backup" -delete

echo
echo "✅ Migration completed!"
echo
echo "📦 Next steps:"
echo "1. Install Qt6 development packages if not already installed:"
echo "   sudo apt install qt6-base-dev qt6-multimedia-dev qt6-webengine-dev"
echo "2. Test the build: go build ./cmd/recuerdo"
echo "3. If there are API compatibility issues, check the migration log above"
echo
echo "🎯 The application should now work properly with Wayland!"
