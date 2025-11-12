// Package charskeyboard provides functionality ported from Python module
//
// Provides a virtual character keyboard for entering special characters
// and diacritics that may not be easily accessible on a standard keyboard.
//
// This is an automated port - implementation may be incomplete.
package charskeyboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/therecipe/qt/widgets"
)

// CharsKeyboardModule is a Go port of the Python CharsKeyboardModule class
type CharsKeyboardModule struct {
	*core.BaseModule
	manager        *core.Manager
	widget         *widgets.QWidget
	charButtons    []*widgets.QPushButton
	targetLineEdit *widgets.QLineEdit
	currentChars   []string
}

// NewCharsKeyboardModule creates a new CharsKeyboardModule instance
func NewCharsKeyboardModule() *CharsKeyboardModule {
	base := core.NewBaseModule("ui", "chars-keyboard-module")
	base.SetRequires("qtApp")

	return &CharsKeyboardModule{
		BaseModule:   base,
		charButtons:  make([]*widgets.QPushButton, 0),
		currentChars: make([]string, 0),
	}
}

// GetWidget returns the main widget for this keyboard
func (mod *CharsKeyboardModule) GetWidget() *widgets.QWidget {
	if mod.widget == nil {
		mod.createWidget()
	}
	return mod.widget
}

// createWidget creates the character keyboard widget
func (mod *CharsKeyboardModule) createWidget() {
	mod.widget = widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	mod.widget.SetLayout(layout)

	// Header
	headerLabel := widgets.NewQLabel2("Special Characters", nil, 0)
	headerFont := headerLabel.Font()
	headerFont.SetBold(true)
	headerLabel.SetFont(headerFont)
	layout.AddWidget(headerLabel, 0, 0)

	// Character sets
	mod.createCharacterSet("Common", []string{"á", "é", "í", "ó", "ú", "ñ", "ü", "ç"}, layout)
	mod.createCharacterSet("German", []string{"ä", "ö", "ü", "ß", "Ä", "Ö", "Ü"}, layout)
	mod.createCharacterSet("French", []string{"à", "è", "ù", "â", "ê", "î", "ô", "û", "ë", "ï", "ÿ", "ç"}, layout)
	mod.createCharacterSet("Spanish", []string{"á", "é", "í", "ó", "ú", "ñ", "¿", "¡"}, layout)

	// Clear button
	clearButton := widgets.NewQPushButton2("Clear", nil)
	clearButton.ConnectClicked(func(checked bool) {
		mod.clearTarget()
	})
	layout.AddWidget(clearButton, 0, 0)
}

// createCharacterSet creates a set of character buttons
func (mod *CharsKeyboardModule) createCharacterSet(setName string, chars []string, parentLayout *widgets.QVBoxLayout) {
	groupBox := widgets.NewQGroupBox2(setName, nil)
	layout := widgets.NewQHBoxLayout()
	groupBox.SetLayout(layout)

	for _, char := range chars {
		button := widgets.NewQPushButton2(char, nil)
		button.SetFixedSize2(40, 40)
		button.SetToolTip(fmt.Sprintf("Insert character: %s", char))

		// Capture the character for the closure
		character := char
		button.ConnectClicked(func(checked bool) {
			mod.insertCharacter(character)
		})

		layout.AddWidget(button, 0, 0)
		mod.charButtons = append(mod.charButtons, button)
	}

	layout.AddStretch(1)
	parentLayout.AddWidget(groupBox, 0, 0)
}

// SetTargetLineEdit sets the line edit widget that will receive the characters
func (mod *CharsKeyboardModule) SetTargetLineEdit(lineEdit *widgets.QLineEdit) {
	mod.targetLineEdit = lineEdit
}

// insertCharacter inserts a character into the target line edit
func (mod *CharsKeyboardModule) insertCharacter(char string) {
	if mod.targetLineEdit == nil {
		fmt.Printf("Character typed: %s (no target set)\n", char)
		return
	}

	// Get current text and cursor position
	currentText := mod.targetLineEdit.Text()
	cursorPos := mod.targetLineEdit.CursorPosition()

	// Insert character at cursor position
	newText := currentText[:cursorPos] + char + currentText[cursorPos:]
	mod.targetLineEdit.SetText(newText)

	// Move cursor after inserted character
	mod.targetLineEdit.SetCursorPosition(cursorPos + len(char))

	// Give focus back to the line edit
	mod.targetLineEdit.SetFocus2()
}

// clearTarget clears the target line edit
func (mod *CharsKeyboardModule) clearTarget() {
	if mod.targetLineEdit != nil {
		mod.targetLineEdit.Clear()
		mod.targetLineEdit.SetFocus2()
	}
}

// AddCustomCharacters adds custom characters to the keyboard
func (mod *CharsKeyboardModule) AddCustomCharacters(setName string, chars []string) {
	// Store characters for later use, avoid layout casting issues for now
	mod.currentChars = append(mod.currentChars, chars...)
	fmt.Printf("Added custom character set '%s': %v\n", setName, chars)
}

// GetAvailableCharacters returns all available characters
func (mod *CharsKeyboardModule) GetAvailableCharacters() []string {
	chars := []string{}
	for _, button := range mod.charButtons {
		chars = append(chars, button.Text())
	}
	return chars
}

// FilterCharacters shows only characters matching the filter
func (mod *CharsKeyboardModule) FilterCharacters(filter string) {
	filter = strings.ToLower(filter)

	for _, button := range mod.charButtons {
		char := strings.ToLower(button.Text())
		if filter == "" || strings.Contains(char, filter) {
			button.SetVisible(true)
		} else {
			button.SetVisible(false)
		}
	}
}

// ResetFilter shows all characters
func (mod *CharsKeyboardModule) ResetFilter() {
	for _, button := range mod.charButtons {
		button.SetVisible(true)
	}
}

// Enable activates the module
func (mod *CharsKeyboardModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("CharsKeyboardModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *CharsKeyboardModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up widget
	if mod.widget != nil {
		mod.widget.Close()
		mod.widget = nil
		mod.charButtons = make([]*widgets.QPushButton, 0)
	}

	mod.targetLineEdit = nil

	fmt.Println("CharsKeyboardModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *CharsKeyboardModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitCharsKeyboardModule creates and returns a new CharsKeyboardModule instance
func InitCharsKeyboardModule() core.Module {
	return NewCharsKeyboardModule()
}
