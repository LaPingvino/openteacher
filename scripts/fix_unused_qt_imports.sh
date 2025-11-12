#!/bin/bash

# Script to fix unused Qt widgets imports by adding blank identifier
# This allows the code to compile while preserving imports for future implementation

echo "Fixing unused Qt widgets imports..."

# Find all Go files with unused qt/widgets imports
files_with_unused_imports=(
    "internal/modules/interfaces/qt/progressViewer/progressViewer.go"
    "internal/modules/interfaces/qt/printer/printer.go"
    "internal/modules/interfaces/qt/ocrGui/ocrGui.go"
    "internal/modules/interfaces/qt/mediaTypes/text/text.go"
    "internal/modules/interfaces/qt/mediaDisplay/mediaDisplay.go"
    "internal/modules/interfaces/qt/mediaTypes/image/image.go"
    "internal/modules/interfaces/qt/mediaTypes/youtube/youtube.go"
    "internal/modules/interfaces/qt/recentlyOpenedViewer/recentlyOpenedViewer.go"
    "internal/modules/interfaces/qt/mediaTypes/video/video.go"
    "internal/modules/interfaces/qt/mediaTypes/audio/audio.go"
    "internal/modules/interfaces/qt/mediaTypes/liveleak/liveleak.go"
    "internal/modules/interfaces/qt/percentNotesViewer/percentNotesViewer.go"
    "internal/modules/interfaces/qt/mediaTypes/dailymotion/dailymotion.go"
    "internal/modules/interfaces/qt/mediaTypes/vimeo/vimeo.go"
    "internal/modules/interfaces/qt/loaderGui/loaderGui.go"
    "internal/modules/interfaces/qt/settingsWidget/boolean/boolean.go"
    "internal/modules/interfaces/qt/settingsWidget/number/number.go"
    "internal/modules/interfaces/qt/settingsWidget/characterTable/characterTable.go"
    "internal/modules/interfaces/qt/settingsWidget/language/language.go"
    "internal/modules/interfaces/qt/mediaTypes/website/website.go"
    "internal/modules/interfaces/qt/settingsWidget/multiOption/multiOption.go"
    "internal/modules/interfaces/qt/settingsWidget/longText/longText.go"
    "internal/modules/interfaces/qt/settingsWidget/option/option.go"
    "internal/modules/interfaces/qt/settingsWidget/shortText/shortText.go"
    "internal/modules/interfaces/qt/settingsWidgets/settingsWidgets.go"
    "internal/modules/interfaces/qt/settingsWidget/password/password.go"
    "internal/modules/interfaces/qt/settingsWidget/profile/profile.go"
    "internal/modules/interfaces/qt/teachTypes/inMind/inMind.go"
    "internal/modules/interfaces/qt/testMode/menu/menu.go"
    "internal/modules/interfaces/qt/teachTypes/shuffleAnswer/shuffleAnswer.go"
    "internal/modules/interfaces/qt/teachTypes/hangman/hangman.go"
    "internal/modules/interfaces/qt/teachTypes/typing/typing.go"
    "internal/modules/interfaces/qt/teachTypes/repeatAnswer/repeatAnswer.go"
    "internal/modules/interfaces/qt/topoMaps/topoMaps.go"
    "internal/modules/interfaces/qt/theme/theme.go"
    "internal/modules/interfaces/qt/testMode/studentsView/studentsView.go"
    "internal/modules/interfaces/qt/typingTutor/keyboard/keyboard.go"
)

fixed_count=0
skipped_count=0

for file in "${files_with_unused_imports[@]}"; do
    if [ ! -f "$file" ]; then
        echo "Warning: File not found: $file"
        continue
    fi

    # Check if file contains the unused import pattern
    if grep -q '"github.com/therecipe/qt/widgets"' "$file"; then
        # Check if it already has the blank identifier
        if grep -q '_ "github.com/therecipe/qt/widgets"' "$file"; then
            echo "Already fixed: $file"
            ((skipped_count++))
        else
            # Replace the import with blank identifier version
            sed -i 's|"github.com/therecipe/qt/widgets"|_ "github.com/therecipe/qt/widgets"|g' "$file"
            echo "Fixed: $file"
            ((fixed_count++))
        fi
    else
        echo "No matching import found in: $file"
        ((skipped_count++))
    fi
done

echo "Summary:"
echo "  Fixed: $fixed_count files"
echo "  Skipped: $skipped_count files"
echo "  Total: $((fixed_count + skipped_count)) files processed"

# Test compilation after fixes
echo ""
echo "Testing compilation..."
if go build ./cmd/openteacher >/dev/null 2>&1; then
    echo "✅ Compilation successful!"
else
    echo "❌ Compilation still has errors. Running 'go build' for details:"
    go build ./cmd/openteacher
fi
