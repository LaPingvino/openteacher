// Package plaintextwords provides functionality ported from Python module
//
// Provides a plain text word enterer for simple text-based word list input.
// This module allows users to enter word lists in plain text format.
//
// This is an automated port - implementation may be incomplete.
package plaintextwords

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/therecipe/qt/widgets"
)

// PlainTextWordsEntererModule is a Go port of the Python PlainTextWordsEntererModule class
type PlainTextWordsEntererModule struct {
	*core.BaseModule
	manager     *core.Manager
	widget      *widgets.QWidget
	textEdit    *widgets.QTextEdit
	parseButton *widgets.QPushButton
	clearButton *widgets.QPushButton
	wordPairs   []string
}

// NewPlainTextWordsEntererModule creates a new PlainTextWordsEntererModule instance
func NewPlainTextWordsEntererModule() *PlainTextWordsEntererModule {
	base := core.NewBaseModule("ui", "plaintext-words-enterer-module")
	base.SetRequires("qtApp")

	return &PlainTextWordsEntererModule{
		BaseModule: base,
		wordPairs:  make([]string, 0),
	}
}

// GetWidget returns the main widget for this enterer
func (mod *PlainTextWordsEntererModule) GetWidget() *widgets.QWidget {
	if mod.widget == nil {
		mod.createWidget()
	}
	return mod.widget
}

// createWidget creates the main plain text enterer widget
func (mod *PlainTextWordsEntererModule) createWidget() {
	mod.widget = widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	mod.widget.SetLayout(layout)

	// Header
	headerLabel := widgets.NewQLabel2("Plain Text Word Enterer", nil, 0)
	headerFont := headerLabel.Font()
	headerFont.SetBold(true)
	headerFont.SetPointSize(14)
	headerLabel.SetFont(headerFont)
	layout.AddWidget(headerLabel, 0, 0)

	// Instructions
	instructionsLabel := widgets.NewQLabel2("Enter word pairs, one per line. Separate questions and answers with a tab or comma.", nil, 0)
	instructionsLabel.SetWordWrap(true)
	layout.AddWidget(instructionsLabel, 0, 0)

	// Text editor
	mod.textEdit = widgets.NewQTextEdit(nil)
	mod.textEdit.SetPlaceholderText("Example:\nhello\tworld\ngood morning\tguten Morgen\nhow are you?\twie geht es dir?")
	layout.AddWidget(mod.textEdit, 1, 0)

	// Button layout
	buttonLayout := widgets.NewQHBoxLayout()

	mod.parseButton = widgets.NewQPushButton2("Parse Words", nil)
	mod.parseButton.ConnectClicked(func(checked bool) {
		mod.parseWords()
	})
	buttonLayout.AddWidget(mod.parseButton, 0, 0)

	mod.clearButton = widgets.NewQPushButton2("Clear All", nil)
	mod.clearButton.ConnectClicked(func(checked bool) {
		mod.clearText()
	})
	buttonLayout.AddWidget(mod.clearButton, 0, 0)

	buttonLayout.AddStretch(1)

	// Status label
	statusLabel := widgets.NewQLabel2("Enter your word list above and click 'Parse Words' to process.", nil, 0)
	statusLabel.SetObjectName("statusLabel")
	buttonLayout.AddWidget(statusLabel, 1, 0)

	layout.AddLayout(buttonLayout, 0)
}

// parseWords parses the text content into word pairs
func (mod *PlainTextWordsEntererModule) parseWords() {
	if mod.textEdit == nil {
		return
	}

	text := mod.textEdit.ToPlainText()
	if strings.TrimSpace(text) == "" {
		mod.showStatus("Please enter some text first.")
		return
	}

	lines := strings.Split(text, "\n")
	mod.wordPairs = make([]string, 0)
	validPairs := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try tab separator first, then comma
		var parts []string
		if strings.Contains(line, "\t") {
			parts = strings.Split(line, "\t")
		} else if strings.Contains(line, ",") {
			parts = strings.Split(line, ",")
		} else {
			// Single word - use as both question and answer
			parts = []string{line, line}
		}

		if len(parts) >= 2 {
			question := strings.TrimSpace(parts[0])
			answer := strings.TrimSpace(parts[1])
			if question != "" && answer != "" {
				mod.wordPairs = append(mod.wordPairs, fmt.Sprintf("%s -> %s", question, answer))
				validPairs++
			}
		}
	}

	mod.showStatus(fmt.Sprintf("Parsed %d valid word pairs from %d lines.", validPairs, len(lines)))
}

// clearText clears all text content
func (mod *PlainTextWordsEntererModule) clearText() {
	if mod.textEdit != nil {
		reply := widgets.QMessageBox_Question(mod.widget,
			"Clear Text",
			"Are you sure you want to clear all text?",
			widgets.QMessageBox__Yes|widgets.QMessageBox__No,
			widgets.QMessageBox__No)

		if reply == widgets.QMessageBox__Yes {
			mod.textEdit.Clear()
			mod.wordPairs = make([]string, 0)
			mod.showStatus("Text cleared.")
		}
	}
}

// showStatus updates the status message
func (mod *PlainTextWordsEntererModule) showStatus(message string) {
	// For now, just print to console to avoid Qt API issues
	fmt.Printf("Status: %s\n", message)
}

// GetWordPairs returns the parsed word pairs
func (mod *PlainTextWordsEntererModule) GetWordPairs() []string {
	return mod.wordPairs
}

// SetText sets the text content
func (mod *PlainTextWordsEntererModule) SetText(text string) {
	if mod.textEdit != nil {
		mod.textEdit.SetPlainText(text)
	}
}

// GetText returns the current text content
func (mod *PlainTextWordsEntererModule) GetText() string {
	if mod.textEdit != nil {
		return mod.textEdit.ToPlainText()
	}
	return ""
}

// Enable activates the module
func (mod *PlainTextWordsEntererModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("PlainTextWordsEntererModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *PlainTextWordsEntererModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up widget
	if mod.widget != nil {
		mod.widget.Close()
		mod.widget = nil
		mod.textEdit = nil
		mod.parseButton = nil
		mod.clearButton = nil
	}

	fmt.Println("PlainTextWordsEntererModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *PlainTextWordsEntererModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitPlainTextWordsEntererModule creates and returns a new PlainTextWordsEntererModule instance
func InitPlainTextWordsEntererModule() core.Module {
	return NewPlainTextWordsEntererModule()
}
