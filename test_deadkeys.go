package main

import (
	"fmt"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func main() {
	// Create Qt application
	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("Dead Keys Test")

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Dead Keys & AltGr Test - OpenTeacher")
	window.Resize2(600, 400)

	// Create central widget
	central := widgets.NewQWidget(window, 0)
	window.SetCentralWidget(central)

	layout := widgets.NewQVBoxLayout()
	central.SetLayout(layout)

	// Instructions
	instructions := widgets.NewQLabel(nil, 0)
	instructions.SetText(`Test Dead Keys & AltGr Input:

Dead Keys Test:
• Type: ´ then o → should get: ó
• Type: ` + "`" + ` then a → should get: à
• Type: ~ then n → should get: ñ
• Type: ¨ then u → should get: ü

AltGr Test (if available):
• Try AltGr + key combinations for special characters

Type in the field below to test:`)

	instructions.SetWordWrap(true)
	instructions.SetAlignment(core.Qt__AlignTop)
	layout.AddWidget(instructions, 0, 0)

	// Test input field - MINIMAL configuration
	testEdit := widgets.NewQLineEdit(central)
	testEdit.SetPlaceholderText("Type accented characters here using dead keys...")

	// Set Unicode font only - NO input method configuration
	font := gui.NewQFont()
	font.SetFamily("DejaVu Sans, Liberation Sans, Arial, sans-serif")
	font.SetPointSize(14)
	testEdit.SetFont(font)

	// DON'T set any input method hints or attributes
	// Let Qt use system defaults for dead key processing

	layout.AddWidget(testEdit, 0, 0)

	// Output display
	output := widgets.NewQTextEdit(central)
	output.SetReadOnly(true)
	output.SetMaximumHeight(150)
	output.SetFont(font)
	layout.AddWidget(output, 0, 0)

	// Connect text changes to show what was typed
	testEdit.ConnectTextChanged(func(text string) {
		if text != "" {
			output.Append(fmt.Sprintf("Typed: '%s' (length: %d bytes, %d runes)",
				text, len(text), len([]rune(text))))

			// Show individual characters
			for i, r := range text {
				output.Append(fmt.Sprintf("  [%d]: '%c' (U+%04X)", i, r, r))
			}
			output.Append("---")
		}
	})

	// Clear button
	clearBtn := widgets.NewQPushButton2("Clear", central)
	clearBtn.ConnectClicked(func(checked bool) {
		testEdit.Clear()
		output.Clear()
		testEdit.SetFocus2()
	})
	layout.AddWidget(clearBtn, 0, 0)

	// Status
	status := widgets.NewQLabel(nil, 0)
	status.SetText("Environment: " + os.Getenv("QT_IM_MODULE") + " | Platform: " + os.Getenv("QT_QPA_PLATFORM"))
	layout.AddWidget(status, 0, 0)

	// Show window and focus input
	window.Show()
	testEdit.SetFocus2()

	fmt.Println("Dead Keys Test Application Started")
	fmt.Println("Window should be visible - test typing accented characters")
	fmt.Println("Expected behavior:")
	fmt.Println("  - Dead acute (´) + o = ó")
	fmt.Println("  - Dead grave (`) + a = à")
	fmt.Println("  - Dead tilde (~) + n = ñ")
	fmt.Println("  - Dead diaeresis (¨) + u = ü")

	// Run event loop
	app.Exec()
}
