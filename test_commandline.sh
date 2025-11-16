#!/bin/bash

# Test script for Recuerdo command-line functionality
# This script demonstrates the various command-line options available

set -e  # Exit on any error

echo "🚀 Recuerdo Command-Line Interface Test Script"
echo "=============================================="
echo

# Build the application
echo "📦 Building Recuerdo..."
go build ./cmd/recuerdo
echo "✅ Build complete"
echo

# Test 1: Show help
echo "📖 Test 1: Showing help information"
echo "Command: ./recuerdo --help"
./recuerdo --help
echo
echo "✅ Help test complete"
echo

# Test 2: List available commands
echo "📋 Test 2: Listing available commands"
echo "Command: ./recuerdo --list-commands"
./recuerdo --list-commands
echo
echo "✅ Command list test complete"
echo

# Test 3: Load lesson file only
echo "📂 Test 3: Loading lesson file"
echo "Command: timeout 8 ./recuerdo testdata/lessons/sample.ot"
echo "This will load the Spanish vocabulary lesson and show the GUI for 8 seconds"
echo "Press Ctrl+C or wait for timeout..."
timeout 8 ./recuerdo --lesson=testdata/lessons/sample.ot 2>/dev/null || true
echo
echo "✅ Lesson loading test complete"
echo

# Test 4: Execute single command
echo "🎯 Test 4: Execute single command (new lesson dialog)"
echo "Command: timeout 5 ./recuerdo --commands=new-lesson"
echo "This will show the new lesson dialog for 5 seconds"
echo "Press Ctrl+C or wait for timeout..."
timeout 5 ./recuerdo --commands=new-lesson 2>/dev/null || true
echo
echo "✅ Single command test complete"
echo

# Test 5: Load lesson and show properties
echo "🔧 Test 5: Load lesson and show properties dialog"
echo "Command: timeout 6 ./recuerdo testdata/lessons/sample.ot --commands=show-properties"
echo "This will load the lesson and automatically open the properties dialog"
echo "Press Ctrl+C or wait for timeout..."
timeout 6 ./recuerdo testdata/lessons/sample.ot --commands=show-properties 2>/dev/null || true
echo
echo "✅ Lesson + properties test complete"
echo

# Test 6: Multiple commands
echo "🎪 Test 6: Execute multiple commands"
echo "Command: timeout 8 ./recuerdo --commands=new-lesson,show-settings"
echo "This will show new lesson dialog, then settings dialog"
echo "Press Ctrl+C or wait for timeout..."
timeout 8 ./recuerdo --commands=new-lesson,show-settings 2>/dev/null || true
echo
echo "✅ Multiple commands test complete"
echo

# Test 7: Non-existent lesson file
echo "❌ Test 7: Error handling - non-existent file"
echo "Command: timeout 3 ./recuerdo nonexistent.ot"
echo "This should show an error message about the missing file"
timeout 3 ./recuerdo nonexistent.ot 2>&1 | grep -E "(Error|error|not exist)" || echo "No error message found"
echo
echo "✅ Error handling test complete"
echo

echo "🎉 All command-line tests completed!"
echo
echo "📚 Summary of available functionality:"
echo "  • ./recuerdo                                    - Start normally"
echo "  • ./recuerdo file.ot                            - Load lesson file on startup"
echo "  • ./recuerdo --commands=cmd1,cmd2              - Execute commands automatically"
echo "  • ./recuerdo file.ot --commands=show-properties - Load file and show properties"
echo "  • ./recuerdo --list-commands                    - List all available commands"
echo "  • ./recuerdo --help                             - Show help"
echo
echo "🔧 Available commands:"
echo "  • show-properties - Show lesson properties dialog"
echo "  • show-settings   - Show application settings dialog"
echo "  • show-about      - Show about dialog"
echo "  • new-lesson      - Create a new lesson"
echo "  • open-file       - Show open file dialog"
echo "  • exit            - Exit the application"
echo
echo "🎯 Example use cases for testing:"
echo "  # Quick property editing:"
echo "  ./recuerdo my_lesson.ot --commands=show-properties"
echo
echo "  # Batch testing of dialogs:"
echo "  ./recuerdo --commands=show-about,show-settings,new-lesson"
echo
echo "  # Automated testing workflow:"
echo "  ./recuerdo test.ot --commands=show-properties,show-settings,exit"
echo
echo "✨ The command-line interface makes testing much more efficient!"
echo "   No more manual clicking through the GUI to test specific features."
