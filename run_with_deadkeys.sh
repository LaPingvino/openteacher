#!/bin/bash

# OpenTeacher launcher script with proper dead key and AltGr support
# This ensures Qt doesn't interfere with normal keyboard input processing

echo "=== OpenTeacher Dead Keys & AltGr Launcher ==="
echo "Configuring environment for proper keyboard input..."

# Ensure Qt uses the system's native input method handling
export QT_IM_MODULE=""  # Don't force any specific input module
export QT_AUTO_SCREEN_SCALE_FACTOR=1

# Make sure Qt doesn't override keyboard input processing
export QT_LOGGING_RULES="qt.qpa.input.events=true"

# Use native platform integration
if [ "$XDG_SESSION_TYPE" = "wayland" ]; then
    echo "Detected Wayland session - using Qt Wayland platform"
    export QT_QPA_PLATFORM=wayland
else
    echo "Using X11 platform"
    export QT_QPA_PLATFORM=xcb
fi

# Ensure proper locale for character handling
export LC_CTYPE=en_US.UTF-8

echo "Environment configured:"
echo "  QT_IM_MODULE: $QT_IM_MODULE"
echo "  QT_QPA_PLATFORM: $QT_QPA_PLATFORM"
echo "  LC_CTYPE: $LC_CTYPE"
echo ""
echo "Dead keys test: Try typing ´+o=ó, `+a=à, ~+n=ñ"
echo "AltGr test: Try AltGr+key combinations for special characters"
echo ""
echo "Starting OpenTeacher..."

# Run OpenTeacher with proper environment
./openteacher "$@"
