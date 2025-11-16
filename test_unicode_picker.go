package main

import (
	"fmt"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

func main() {
	// Create Qt application
	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("Unicode Picker Test")

	// Set UTF-8 encoding for proper Unicode handling
	app.SetAttribute(core.Qt__AA_UseHighDpiPixmaps, true)
	app.SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Unicode Picker Test")
	window.Resize2(400, 200)

	// Create central widget
	central := widgets.NewQWidget(window, 0)
	window.SetCentralWidget(central)

	layout := widgets.NewQVBoxLayout()
	central.SetLayout(layout)

	// Instructions
	instructions := widgets.NewQLabel(nil, 0)
	instructions.SetText("Click the button to test the Unicode character picker.")
	instructions.SetWordWrap(true)
	layout.AddWidget(instructions, 0, 0)

	// Test input field
	testEdit := widgets.NewQLineEdit(central)
	testEdit.SetPlaceholderText("Characters will be inserted here...")
	layout.AddWidget(testEdit, 0, 0)

	// Button to show picker
	showButton := widgets.NewQPushButton2("Show Unicode Picker", central)
	layout.AddWidget(showButton, 0, 0)

	// Create the picker window
	picker := widgets.NewQWidget(nil, 0)
	picker.SetWindowTitle("Unicode Character Picker")
	picker.SetFixedSize2(500, 300)
	picker.SetWindowFlags(core.Qt__Window | core.Qt__WindowStaysOnTopHint)

	pickerLayout := widgets.NewQVBoxLayout()
	picker.SetLayout(pickerLayout)

	// Title
	pickerTitle := widgets.NewQLabel2("Click a character to insert it:", picker, 0)
	pickerLayout.AddWidget(pickerTitle, 0, 0)

	// Character buttons
	charWidget := widgets.NewQWidget(picker, 0)
	charLayout := widgets.NewQGridLayout2()
	charWidget.SetLayout(charLayout)

	// Set UTF-8 friendly font
	charFont := charWidget.Font()
	charFont.SetFamily("DejaVu Sans, Liberation Sans, Arial, sans-serif")
	charFont.SetPointSize(14)
	charWidget.SetFont(charFont)

	// Spanish characters
	spanishChars := []string{"á", "é", "í", "ó", "ú", "ñ", "ü", "Á", "É", "Í", "Ó", "Ú", "Ñ", "Ü", "¿", "¡"}

	row, col := 0, 0
	maxCols := 8

	for _, char := range spanishChars {
		button := widgets.NewQPushButton2(char, charWidget)
		button.SetMinimumSize2(40, 40)
		button.SetStyleSheet("font-size: 16px; font-weight: bold;")
		button.SetFont(charFont)

		// Set tooltip with Unicode info
		unicodePoint := fmt.Sprintf("U+%04X", []rune(char)[0])
		button.SetToolTip(fmt.Sprintf("%s (%s)", char, unicodePoint))

		// Capture character in closure
		character := char
		button.ConnectClicked(func(checked bool) {
			// Insert character at cursor position with proper UTF-8 handling
			cursorPos := testEdit.CursorPosition()
			currentText := testEdit.Text()

			// Convert to runes for proper Unicode handling
			currentRunes := []rune(currentText)
			charRunes := []rune(character)

			// Ensure cursor position is within bounds
			if cursorPos > len(currentRunes) {
				cursorPos = len(currentRunes)
			}
			if cursorPos < 0 {
				cursorPos = 0
			}

			// Insert character at cursor position
			newRunes := make([]rune, 0, len(currentRunes)+len(charRunes))
			newRunes = append(newRunes, currentRunes[:cursorPos]...)
			newRunes = append(newRunes, charRunes...)
			newRunes = append(newRunes, currentRunes[cursorPos:]...)

			newText := string(newRunes)
			testEdit.SetText(newText)
			testEdit.SetCursorPosition(cursorPos + len(charRunes))
			testEdit.SetFocus2()
			fmt.Printf("Inserted character: %s (runes: %d)\n", character, len(charRunes))
		})

		charLayout.AddWidget3(button, row, col, 1, 1, 0)

		col++
		if col >= maxCols {
			col = 0
			row++
		}
	}

	pickerLayout.AddWidget(charWidget, 0, 0)

	// Close button for picker
	closeButton := widgets.NewQPushButton2("Close", picker)
	closeButton.SetStyleSheet("background-color: #ff6666; color: white; font-weight: bold;")
	closeButton.ConnectClicked(func(checked bool) {
		picker.Hide()
		fmt.Println("Picker hidden")
	})
	pickerLayout.AddWidget(closeButton, 0, 0)

	// Connect show button
	showButton.ConnectClicked(func(checked bool) {
		fmt.Println("Show button clicked")

		// Center the picker on screen
		screen := widgets.QApplication_Desktop().AvailableGeometry2(nil)
		centerX := screen.Width()/2 - picker.Width()/2
		centerY := screen.Height()/2 - picker.Height()/2
		picker.Move2(centerX, centerY)

		picker.Show()
		picker.Raise()
		picker.ActivateWindow()
		fmt.Println("Picker should be visible now")
	})

	// Show main window
	window.Show()

	fmt.Println("Unicode Picker Test started")
	fmt.Println("Click 'Show Unicode Picker' to test character insertion")

	// Run event loop
	app.Exec()
}
