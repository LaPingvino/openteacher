// Package inputtyping provides functionality ported from Python module
//
// Provides a typing input widget for teaching modes where students
// need to type their answers. Includes features like spell checking,
// auto-completion, and typing statistics.
//
// This is an automated port - implementation may be incomplete.
package inputtyping

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// TypingStats holds typing statistics
type TypingStats struct {
	StartTime      time.Time
	EndTime        time.Time
	CharsTyped     int
	CorrectChars   int
	Mistakes       int
	WordsPerMinute float64
}

// InputTypingModule is a Go port of the Python InputTypingModule class
type InputTypingModule struct {
	*core.BaseModule
	manager          *core.Manager
	widget           *widgets.QWidget
	inputField       *widgets.QLineEdit
	targetLabel      *widgets.QLabel
	statsLabel       *widgets.QLabel
	submitButton     *widgets.QPushButton
	clearButton      *widgets.QPushButton
	hintButton       *widgets.QPushButton
	targetText       string
	currentInput     string
	startTime        time.Time
	stats            TypingStats
	allowHints       bool
	caseSensitive    bool
	enableSpellCheck bool
	completionModel  *gui.QStringListModel
	completer        *widgets.QCompleter
}

// NewInputTypingModule creates a new InputTypingModule instance
func NewInputTypingModule() *InputTypingModule {
	base := core.NewBaseModule("ui", "input-typing-module")
	base.SetRequires("qtApp")

	return &InputTypingModule{
		BaseModule:       base,
		allowHints:       true,
		caseSensitive:    false,
		enableSpellCheck: true,
	}
}

// GetWidget returns the main widget for this input module
func (mod *InputTypingModule) GetWidget() *widgets.QWidget {
	if mod.widget == nil {
		mod.createWidget()
	}
	return mod.widget
}

// createWidget creates the main typing input widget
func (mod *InputTypingModule) createWidget() {
	mod.widget = widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	mod.widget.SetLayout(layout)

	// Header
	headerLabel := widgets.NewQLabel2("Typing Input", nil, 0)
	headerFont := headerLabel.Font()
	headerFont.SetBold(true)
	headerFont.SetPointSize(14)
	headerLabel.SetFont(headerFont)
	headerLabel.SetAlignment(core.Qt__AlignHCenter)
	layout.AddWidget(headerLabel, 0, 0)

	// Target text display
	targetGroup := widgets.NewQGroupBox2("Target Text", nil)
	targetLayout := widgets.NewQVBoxLayout()
	targetGroup.SetLayout(targetLayout)

	mod.targetLabel = widgets.NewQLabel2("Type the text shown here...", nil, 0)
	mod.targetLabel.SetWordWrap(true)
	mod.targetLabel.SetAlignment(core.Qt__AlignCenter)
	mod.targetLabel.SetStyleSheet("QLabel { background-color: #f0f0f0; padding: 15px; border: 1px solid #ccc; border-radius: 5px; font-size: 16px; }")
	mod.targetLabel.SetMinimumHeight(80)
	targetLayout.AddWidget(mod.targetLabel, 0, 0)

	layout.AddWidget(targetGroup, 0, 0)

	// Input field
	inputGroup := widgets.NewQGroupBox2("Your Input", nil)
	inputLayout := widgets.NewQVBoxLayout()
	inputGroup.SetLayout(inputLayout)

	mod.inputField = widgets.NewQLineEdit(nil)
	mod.inputField.SetPlaceholderText("Start typing here...")
	mod.inputField.SetStyleSheet("QLineEdit { font-size: 16px; padding: 8px; }")
	inputLayout.AddWidget(mod.inputField, 0, 0)

	// Setup auto-completion
	mod.setupAutoCompletion()

	// Connect input field signals
	mod.inputField.ConnectTextChanged(func(text string) {
		mod.onInputChanged(text)
	})

	mod.inputField.ConnectReturnPressed(func() {
		mod.onSubmit()
	})

	layout.AddWidget(inputGroup, 0, 0)

	// Statistics display
	mod.statsLabel = widgets.NewQLabel2("Ready to type...", nil, 0)
	mod.statsLabel.SetAlignment(core.Qt__AlignCenter)
	mod.statsLabel.SetStyleSheet("QLabel { color: #666; font-style: italic; }")
	layout.AddWidget(mod.statsLabel, 0, 0)

	// Button layout
	buttonLayout := widgets.NewQHBoxLayout()

	mod.hintButton = widgets.NewQPushButton2("Show Hint", nil)
	mod.hintButton.ConnectClicked(func(checked bool) {
		mod.showHint()
	})
	buttonLayout.AddWidget(mod.hintButton, 0, 0)

	buttonLayout.AddStretch(1)

	mod.clearButton = widgets.NewQPushButton2("Clear", nil)
	mod.clearButton.ConnectClicked(func(checked bool) {
		mod.clearInput()
	})
	buttonLayout.AddWidget(mod.clearButton, 0, 0)

	mod.submitButton = widgets.NewQPushButton2("Submit", nil)
	mod.submitButton.ConnectClicked(func(checked bool) {
		mod.onSubmit()
	})
	mod.submitButton.SetDefault(true)
	buttonLayout.AddWidget(mod.submitButton, 0, 0)

	layout.AddLayout(buttonLayout, 0)

	// Settings layout
	settingsLayout := widgets.NewQHBoxLayout()

	caseSensitiveCheck := widgets.NewQCheckBox2("Case sensitive", nil)
	caseSensitiveCheck.SetChecked(mod.caseSensitive)
	caseSensitiveCheck.ConnectToggled(func(checked bool) {
		mod.caseSensitive = checked
		mod.updateValidation()
	})
	settingsLayout.AddWidget(caseSensitiveCheck, 0, 0)

	spellCheckCheck := widgets.NewQCheckBox2("Spell checking", nil)
	spellCheckCheck.SetChecked(mod.enableSpellCheck)
	spellCheckCheck.ConnectToggled(func(checked bool) {
		mod.enableSpellCheck = checked
		mod.setupAutoCompletion()
	})
	settingsLayout.AddWidget(spellCheckCheck, 0, 0)

	settingsLayout.AddStretch(1)

	layout.AddLayout(settingsLayout, 0)

	// Initialize with default target
	mod.SetTargetText("Hello World")
}

// setupAutoCompletion sets up auto-completion for the input field
func (mod *InputTypingModule) setupAutoCompletion() {
	if !mod.enableSpellCheck {
		if mod.completer != nil {
			mod.inputField.SetCompleter(nil)
			mod.completer = nil
		}
		return
	}

	// Common English words for completion
	commonWords := []string{
		"hello", "world", "the", "and", "you", "that", "was", "for", "are",
		"with", "his", "they", "this", "have", "from", "one", "had", "but",
		"words", "not", "what", "all", "were", "when", "your", "can", "said",
		"there", "each", "which", "she", "how", "will", "about", "write",
		"would", "like", "these", "her", "long", "make", "thing", "see",
		"him", "two", "more", "has", "way", "could", "people", "than",
		"first", "been", "call", "who", "its", "now", "find", "may",
		"down", "side", "did", "get", "come", "made", "time", "very",
		"after", "back", "other", "many", "then", "them", "well", "some",
	}

	mod.completionModel = gui.NewQStringListModel2(commonWords, nil)
	mod.completer = widgets.NewQCompleter2(mod.completionModel, nil)
	mod.completer.SetCaseSensitivity(core.Qt__CaseInsensitive)
	mod.completer.SetCompletionMode(widgets.QCompleter__PopupCompletion)
	mod.inputField.SetCompleter(mod.completer)
}

// SetTargetText sets the text that the user should type
func (mod *InputTypingModule) SetTargetText(text string) {
	mod.targetText = text
	if mod.targetLabel != nil {
		mod.targetLabel.SetText(text)
	}
	mod.resetStats()
}

// GetTargetText returns the current target text
func (mod *InputTypingModule) GetTargetText() string {
	return mod.targetText
}

// SetAllowHints enables or disables the hint functionality
func (mod *InputTypingModule) SetAllowHints(allow bool) {
	mod.allowHints = allow
	if mod.hintButton != nil {
		mod.hintButton.SetEnabled(allow)
	}
}

// SetCaseSensitive enables or disables case sensitive comparison
func (mod *InputTypingModule) SetCaseSensitive(sensitive bool) {
	mod.caseSensitive = sensitive
	mod.updateValidation()
}

// onInputChanged handles input field text changes
func (mod *InputTypingModule) onInputChanged(text string) {
	mod.currentInput = text

	// Start timing on first character
	if len(text) == 1 && mod.stats.StartTime.IsZero() {
		mod.startTime = time.Now()
		mod.stats.StartTime = mod.startTime
	}

	mod.updateValidation()
	mod.updateStats()
}

// updateValidation updates visual feedback based on input correctness
func (mod *InputTypingModule) updateValidation() {
	if mod.inputField == nil || mod.targetText == "" {
		return
	}

	currentText := mod.currentInput
	targetText := mod.targetText

	if !mod.caseSensitive {
		currentText = strings.ToLower(currentText)
		targetText = strings.ToLower(targetText)
	}

	// Check if input matches the beginning of target
	isCorrect := strings.HasPrefix(targetText, currentText)

	// Update input field styling based on correctness
	if isCorrect {
		if currentText == targetText {
			// Perfect match - green background
			mod.inputField.SetStyleSheet("QLineEdit { font-size: 16px; padding: 8px; background-color: #d4edda; border: 2px solid #28a745; }")
		} else {
			// Partial match - default styling
			mod.inputField.SetStyleSheet("QLineEdit { font-size: 16px; padding: 8px; }")
		}
	} else {
		// Incorrect - red background
		mod.inputField.SetStyleSheet("QLineEdit { font-size: 16px; padding: 8px; background-color: #f8d7da; border: 2px solid #dc3545; }")
	}

	// Enable/disable submit button based on completion
	if mod.submitButton != nil {
		var exactMatch bool
		if mod.caseSensitive {
			exactMatch = mod.currentInput == mod.targetText
		} else {
			exactMatch = strings.ToLower(mod.currentInput) == strings.ToLower(mod.targetText)
		}
		mod.submitButton.SetEnabled(exactMatch)
	}
}

// updateStats updates typing statistics display
func (mod *InputTypingModule) updateStats() {
	if mod.statsLabel == nil {
		return
	}

	if mod.startTime.IsZero() {
		mod.statsLabel.SetText("Ready to type...")
		return
	}

	elapsed := time.Since(mod.startTime)
	if elapsed.Seconds() < 1 {
		mod.statsLabel.SetText("Keep typing...")
		return
	}

	// Calculate stats
	charsTyped := len(mod.currentInput)
	minutes := elapsed.Minutes()

	// Estimate words (assuming average 5 characters per word)
	wordsTyped := float64(charsTyped) / 5.0
	wpm := wordsTyped / minutes

	// Count correct characters
	correctChars := mod.countCorrectCharacters()
	accuracy := 0.0
	if charsTyped > 0 {
		accuracy = float64(correctChars) / float64(charsTyped) * 100.0
	}

	statsText := fmt.Sprintf("Speed: %.1f WPM | Accuracy: %.1f%% | Time: %ds",
		wpm, accuracy, int(elapsed.Seconds()))
	mod.statsLabel.SetText(statsText)
}

// countCorrectCharacters counts how many characters are correctly typed
func (mod *InputTypingModule) countCorrectCharacters() int {
	currentText := mod.currentInput
	targetText := mod.targetText

	if !mod.caseSensitive {
		currentText = strings.ToLower(currentText)
		targetText = strings.ToLower(targetText)
	}

	correct := 0
	for i, char := range currentText {
		if i < len(targetText) && rune(targetText[i]) == char {
			correct++
		}
	}
	return correct
}

// showHint shows a hint to help the user
func (mod *InputTypingModule) showHint() {
	if !mod.allowHints || mod.targetText == "" {
		return
	}

	currentLen := len(mod.currentInput)
	if currentLen >= len(mod.targetText) {
		widgets.QMessageBox_Information(mod.widget, "Hint", "You've typed the complete text!", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Show next character or next few characters
	hintLen := currentLen + 3
	if hintLen > len(mod.targetText) {
		hintLen = len(mod.targetText)
	}

	hint := mod.targetText[:hintLen]
	message := fmt.Sprintf("Next characters: \"%s\"", hint[currentLen:])

	widgets.QMessageBox_Information(mod.widget, "Hint", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}

// clearInput clears the input field and resets statistics
func (mod *InputTypingModule) clearInput() {
	if mod.inputField != nil {
		mod.inputField.Clear()
	}
	mod.currentInput = ""
	mod.resetStats()
	mod.updateValidation()
}

// resetStats resets typing statistics
func (mod *InputTypingModule) resetStats() {
	mod.startTime = time.Time{}
	mod.stats = TypingStats{}
	if mod.statsLabel != nil {
		mod.statsLabel.SetText("Ready to type...")
	}
}

// onSubmit handles submit button click or Enter key press
func (mod *InputTypingModule) onSubmit() {
	if mod.targetText == "" {
		return
	}

	// Check if input matches target
	currentText := mod.currentInput
	targetText := mod.targetText

	if !mod.caseSensitive {
		currentText = strings.ToLower(currentText)
		targetText = strings.ToLower(targetText)
	}

	if currentText == targetText {
		// Success!
		mod.stats.EndTime = time.Now()
		mod.stats.CharsTyped = len(mod.currentInput)
		mod.stats.CorrectChars = mod.countCorrectCharacters()

		elapsed := mod.stats.EndTime.Sub(mod.stats.StartTime)
		wordsTyped := float64(mod.stats.CharsTyped) / 5.0
		mod.stats.WordsPerMinute = wordsTyped / elapsed.Minutes()

		message := fmt.Sprintf("Congratulations! You typed the text correctly!\n\nStatistics:\n• Speed: %.1f WPM\n• Accuracy: %.1f%%\n• Time: %.1f seconds",
			mod.stats.WordsPerMinute,
			float64(mod.stats.CorrectChars)/float64(mod.stats.CharsTyped)*100.0,
			elapsed.Seconds())

		widgets.QMessageBox_Information(mod.widget, "Success!", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)

		// Clear for next attempt
		mod.clearInput()
	} else {
		// Incorrect
		widgets.QMessageBox_Warning(mod.widget, "Not Quite Right", "Your input doesn't match the target text. Please try again.", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	}
}

// GetCurrentInput returns the current input text
func (mod *InputTypingModule) GetCurrentInput() string {
	return mod.currentInput
}

// GetStats returns the current typing statistics
func (mod *InputTypingModule) GetStats() TypingStats {
	return mod.stats
}

// IsCompleted returns true if the user has successfully typed the target text
func (mod *InputTypingModule) IsCompleted() bool {
	if mod.targetText == "" {
		return false
	}

	currentText := mod.currentInput
	targetText := mod.targetText

	if !mod.caseSensitive {
		currentText = strings.ToLower(currentText)
		targetText = strings.ToLower(targetText)
	}

	return currentText == targetText
}

// Enable activates the module
func (mod *InputTypingModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("InputTypingModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *InputTypingModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up widget
	if mod.widget != nil {
		mod.widget.Close()
		mod.widget = nil
		mod.inputField = nil
		mod.targetLabel = nil
		mod.statsLabel = nil
		mod.submitButton = nil
		mod.clearButton = nil
		mod.hintButton = nil
		mod.completer = nil
		mod.completionModel = nil
	}

	fmt.Println("InputTypingModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *InputTypingModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitInputTypingModule creates and returns a new InputTypingModule instance
func InitInputTypingModule() core.Module {
	return NewInputTypingModule()
}
