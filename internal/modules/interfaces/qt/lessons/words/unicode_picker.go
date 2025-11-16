package words

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// CharacterSet represents a named group of characters
type CharacterSet struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Characters    []string `json:"characters"`
	UnicodeBlocks []string `json:"unicode_blocks,omitempty"`
	Categories    []string `json:"categories,omitempty"`
}

// UnicodePickerConfig holds the configuration for character sets
type UnicodePickerConfig struct {
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	CharacterSets []CharacterSet `json:"character_sets"`
}

// UnicodePicker provides a configurable Unicode character picker
type UnicodePicker struct {
	*widgets.QWidget

	logger     *logging.Logger
	config     *UnicodePickerConfig
	targetEdit *widgets.QLineEdit
	isVisible  bool
	configPath string

	// UI components
	tabWidget     *widgets.QTabWidget
	searchEdit    *widgets.QLineEdit
	searchResults *widgets.QWidget
}

// NewUnicodePicker creates a new Unicode character picker
func NewUnicodePicker(configPath string, parent widgets.QWidget_ITF) *UnicodePicker {
	picker := &UnicodePicker{
		QWidget:    widgets.NewQWidget(parent, 0),
		logger:     logging.NewLogger("UnicodePicker"),
		configPath: configPath,
	}

	picker.loadConfig()
	picker.setupUI()
	picker.Hide()

	return picker
}

// loadConfig loads character sets from the configuration file
func (up *UnicodePicker) loadConfig() {
	// Set default config path if not provided
	if up.configPath == "" {
		homeDir, _ := os.UserHomeDir()
		up.configPath = filepath.Join(homeDir, ".openteacher", "character_sets.json")
	}

	// Create default config if file doesn't exist
	if _, err := os.Stat(up.configPath); os.IsNotExist(err) {
		up.createDefaultConfig()
	}

	// Load configuration
	data, err := ioutil.ReadFile(up.configPath)
	if err != nil {
		up.logger.Error("Failed to read config file: %v", err)
		up.createDefaultConfig()
		return
	}

	up.config = &UnicodePickerConfig{}
	if err := json.Unmarshal(data, up.config); err != nil {
		up.logger.Error("Failed to parse config file: %v", err)
		up.createDefaultConfig()
		return
	}

	// Expand Unicode blocks and categories
	up.expandUnicodeDefinitions()

	up.logger.Success("Loaded character picker config with %d sets", len(up.config.CharacterSets))
}

// createDefaultConfig creates a default character picker configuration
func (up *UnicodePicker) createDefaultConfig() {
	up.config = &UnicodePickerConfig{
		Version:     "1.0",
		Description: "OpenTeacher Unicode Character Picker Configuration",
		CharacterSets: []CharacterSet{
			{
				Name:        "Spanish",
				Description: "Common Spanish characters and symbols",
				Characters:  []string{"á", "é", "í", "ó", "ú", "ñ", "ü", "Á", "É", "Í", "Ó", "Ú", "Ñ", "Ü", "¿", "¡"},
			},
			{
				Name:        "French",
				Description: "Common French characters and symbols",
				Characters:  []string{"à", "â", "ç", "è", "é", "ê", "ë", "î", "ï", "ô", "ù", "û", "ü", "ÿ", "À", "Â", "Ç", "È", "É", "Ê", "Ë", "Î", "Ï", "Ô", "Ù", "Û", "Ü", "Ÿ", "œ", "Œ", "æ", "Æ"},
			},
			{
				Name:        "German",
				Description: "Common German characters",
				Characters:  []string{"ä", "ö", "ü", "ß", "Ä", "Ö", "Ü"},
			},
			{
				Name:        "Math & Symbols",
				Description: "Mathematical and special symbols",
				Characters:  []string{"±", "×", "÷", "≈", "≠", "≤", "≥", "∞", "π", "Σ", "Δ", "√", "∫", "∂", "€", "£", "¥", "¢"},
			},
			{
				Name:          "Latin Extended",
				Description:   "Extended Latin characters from Unicode blocks",
				UnicodeBlocks: []string{"Latin Extended-A", "Latin Extended-B"},
			},
			{
				Name:        "Greek Letters",
				Description: "Greek alphabet for mathematical and scientific use",
				Categories:  []string{"Greek"},
			},
			{
				Name:        "Common Punctuation",
				Description: "Extended punctuation marks",
				Characters:  []string{"–", "—", "\"", "\"", "'", "'", "…", "‚", "„", "‹", "›", "«", "»"},
			},
		},
	}

	// Save default config
	up.saveConfig()
}

// expandUnicodeDefinitions expands Unicode block and category definitions
func (up *UnicodePicker) expandUnicodeDefinitions() {
	for i := range up.config.CharacterSets {
		set := &up.config.CharacterSets[i]

		// Expand Unicode blocks
		for _, blockName := range set.UnicodeBlocks {
			chars := up.getCharactersFromBlock(blockName)
			set.Characters = append(set.Characters, chars...)
		}

		// Expand Unicode categories
		for _, category := range set.Categories {
			chars := up.getCharactersFromCategory(category)
			set.Characters = append(set.Characters, chars...)
		}

		// Remove duplicates
		set.Characters = up.removeDuplicates(set.Characters)
	}
}

// getCharactersFromBlock returns characters from a Unicode block
func (up *UnicodePicker) getCharactersFromBlock(blockName string) []string {
	var chars []string

	// Define Unicode block ranges (simplified - in real implementation, use unicode/rangetable)
	blockRanges := map[string][2]rune{
		"Latin Extended-A": {0x0100, 0x017F},
		"Latin Extended-B": {0x0180, 0x024F},
		"Greek and Coptic": {0x0370, 0x03FF},
		"Cyrillic":         {0x0400, 0x04FF},
	}

	if rng, exists := blockRanges[blockName]; exists {
		for r := rng[0]; r <= rng[1]; r++ {
			if unicode.IsPrint(r) && utf8.ValidRune(r) {
				chars = append(chars, string(r))
			}
		}
	}

	up.logger.Debug("Generated %d characters from Unicode block '%s'", len(chars), blockName)
	return chars
}

// getCharactersFromCategory returns characters from a Unicode category
func (up *UnicodePicker) getCharactersFromCategory(category string) []string {
	var chars []string

	// Sample Greek letters (in real implementation, iterate through Unicode ranges)
	greekLetters := []string{
		"Α", "Β", "Γ", "Δ", "Ε", "Ζ", "Η", "Θ", "Ι", "Κ", "Λ", "Μ",
		"Ν", "Ξ", "Ο", "Π", "Ρ", "Σ", "Τ", "Υ", "Φ", "Χ", "Ψ", "Ω",
		"α", "β", "γ", "δ", "ε", "ζ", "η", "θ", "ι", "κ", "λ", "μ",
		"ν", "ξ", "ο", "π", "ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω",
	}

	switch strings.ToLower(category) {
	case "greek":
		chars = greekLetters
	}

	up.logger.Debug("Generated %d characters from category '%s'", len(chars), category)
	return chars
}

// removeDuplicates removes duplicate characters from a slice
func (up *UnicodePicker) removeDuplicates(chars []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, char := range chars {
		if !seen[char] {
			seen[char] = true
			result = append(result, char)
		}
	}

	return result
}

// saveConfig saves the current configuration to file
func (up *UnicodePicker) saveConfig() {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(up.configPath), 0755); err != nil {
		up.logger.Error("Failed to create config directory: %v", err)
		return
	}

	data, err := json.MarshalIndent(up.config, "", "  ")
	if err != nil {
		up.logger.Error("Failed to marshal config: %v", err)
		return
	}

	if err := ioutil.WriteFile(up.configPath, data, 0644); err != nil {
		up.logger.Error("Failed to write config file: %v", err)
		return
	}

	up.logger.Success("Saved character picker config to: %s", up.configPath)
}

// setupUI creates the Unicode picker interface
func (up *UnicodePicker) setupUI() {
	up.SetFixedSize2(600, 400)
	up.SetWindowTitle("Unicode Character Picker")

	// Use same window flags as working test
	up.SetWindowFlags(core.Qt__Window | core.Qt__WindowStaysOnTopHint)

	up.SetStyleSheet(`
		QWidget {
			background-color: #f8f8f8;
			border: 2px solid #666;
			border-radius: 8px;
		}
		QPushButton {
			background-color: white;
			border: 1px solid #ccc;
			border-radius: 4px;
			padding: 6px;
			margin: 1px;
			font-size: 14px;
			min-width: 32px;
			min-height: 32px;
		}
		QPushButton:hover {
			background-color: #e6f3ff;
			border-color: #0078d4;
		}
		QPushButton:pressed {
			background-color: #cce4f7;
		}
		QLineEdit {
			padding: 8px;
			border: 1px solid #ccc;
			border-radius: 4px;
			font-size: 14px;
		}
		QTabWidget::pane {
			border: 1px solid #ccc;
			background-color: white;
		}
		QTabBar::tab {
			background-color: #e0e0e0;
			padding: 8px 16px;
			margin: 2px;
			border-radius: 4px 4px 0 0;
		}
		QTabBar::tab:selected {
			background-color: white;
			border-bottom: 2px solid #0078d4;
		}
	`)

	mainLayout := widgets.NewQVBoxLayout()
	up.SetLayout(mainLayout)

	// Title and search
	headerLayout := widgets.NewQHBoxLayout()

	titleLabel := widgets.NewQLabel2("Unicode Character Picker", up, 0)
	titleLabel.SetStyleSheet("font-size: 16px; font-weight: bold; margin: 5px;")
	headerLayout.AddWidget(titleLabel, 0, 0)

	headerLayout.AddStretch(0)

	up.searchEdit = widgets.NewQLineEdit(up)
	up.searchEdit.SetPlaceholderText("Search characters...")
	up.searchEdit.SetMaximumWidth(200)
	headerLayout.AddWidget(up.searchEdit, 0, 0)

	mainLayout.AddLayout(headerLayout, 0)

	// Character tabs
	up.tabWidget = widgets.NewQTabWidget(up)
	mainLayout.AddWidget(up.tabWidget, 0, 0)

	// Create tabs for each character set
	up.createCharacterTabs()

	// Search functionality
	up.searchEdit.ConnectTextChanged(func(text string) {
		up.performSearch(text)
	})

	// Close button
	closeButton := widgets.NewQPushButton2("Close", up)
	closeButton.SetStyleSheet(`
		QPushButton {
			background-color: #dc3545;
			color: white;
			font-weight: bold;
			padding: 10px 20px;
			border: none;
		}
		QPushButton:hover {
			background-color: #c82333;
		}
	`)
	closeButton.ConnectClicked(func(checked bool) {
		up.Hide()
		up.isVisible = false
	})
	mainLayout.AddWidget(closeButton, 0, 0)

	up.logger.Success("Unicode picker UI created with %d character sets", len(up.config.CharacterSets))
}

// createCharacterTabs creates tabs for each character set
func (up *UnicodePicker) createCharacterTabs() {
	for _, charSet := range up.config.CharacterSets {
		tab := up.createCharacterTab(charSet)
		up.tabWidget.AddTab(tab, charSet.Name)
	}
}

// createCharacterTab creates a tab with character buttons
func (up *UnicodePicker) createCharacterTab(charSet CharacterSet) *widgets.QWidget {
	tab := widgets.NewQWidget(nil, 0)

	mainLayout := widgets.NewQVBoxLayout()
	tab.SetLayout(mainLayout)

	// Description
	if charSet.Description != "" {
		descLabel := widgets.NewQLabel2(charSet.Description, tab, 0)
		descLabel.SetStyleSheet("font-style: italic; margin: 5px; color: #666;")
		descLabel.SetWordWrap(true)
		mainLayout.AddWidget(descLabel, 0, 0)
	}

	// Scrollable character area
	scrollArea := widgets.NewQScrollArea(tab)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetVerticalScrollBarPolicy(core.Qt__ScrollBarAsNeeded)
	scrollArea.SetHorizontalScrollBarPolicy(core.Qt__ScrollBarAsNeeded)

	charWidget := widgets.NewQWidget(nil, 0)
	gridLayout := widgets.NewQGridLayout2()
	charWidget.SetLayout(gridLayout)

	// Add character buttons
	maxCols := 12
	row, col := 0, 0

	for _, char := range charSet.Characters {
		if strings.TrimSpace(char) == "" {
			continue
		}

		button := widgets.NewQPushButton2(char, charWidget)
		button.SetToolTip(fmt.Sprintf("Insert: %s (U+%04X)", char, []rune(char)[0]))

		// Capture character in closure
		character := char
		button.ConnectClicked(func(checked bool) {
			up.insertCharacter(character)
		})

		gridLayout.AddWidget3(button, row, col, 1, 1, 0)

		col++
		if col >= maxCols {
			col = 0
			row++
		}
	}

	// Add stretch to push buttons to top
	gridLayout.SetRowStretch(row+1, 1)

	scrollArea.SetWidget(charWidget)
	mainLayout.AddWidget(scrollArea, 0, 0)

	up.logger.Debug("Created character tab '%s' with %d characters", charSet.Name, len(charSet.Characters))
	return tab
}

// performSearch searches for characters matching the query
func (up *UnicodePicker) performSearch(query string) {
	if query == "" {
		return
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if len(query) < 2 {
		return
	}

	var matchingChars []string

	// Search through all character sets
	for _, charSet := range up.config.CharacterSets {
		for _, char := range charSet.Characters {
			charName := strings.ToLower(char)
			if strings.Contains(charName, query) {
				matchingChars = append(matchingChars, char)
			}
		}
	}

	up.logger.Debug("Search for '%s' found %d matching characters", query, len(matchingChars))
	// TODO: Show search results in a separate tab or overlay
}

// insertCharacter inserts a character into the target edit widget
func (up *UnicodePicker) insertCharacter(char string) {
	if up.targetEdit == nil {
		up.logger.Warning("No target edit widget set")
		return
	}

	cursorPos := up.targetEdit.CursorPosition()
	currentText := up.targetEdit.Text()

	// Convert to runes for proper Unicode handling
	currentRunes := []rune(currentText)
	charRunes := []rune(char)

	// Ensure cursor position is within bounds
	if cursorPos > len(currentRunes) {
		cursorPos = len(currentRunes)
	}
	if cursorPos < 0 {
		cursorPos = 0
	}

	// Insert character at cursor position using rune slicing
	newRunes := make([]rune, 0, len(currentRunes)+len(charRunes))
	newRunes = append(newRunes, currentRunes[:cursorPos]...)
	newRunes = append(newRunes, charRunes...)
	newRunes = append(newRunes, currentRunes[cursorPos:]...)

	newText := string(newRunes)
	up.targetEdit.SetText(newText)

	// Move cursor to after inserted character (in rune positions)
	newCursorPos := cursorPos + len(charRunes)
	up.targetEdit.SetCursorPosition(newCursorPos)

	// Focus back on edit widget
	up.targetEdit.SetFocus2()

	up.logger.Info("Inserted '%s' at position %d (runes: %d)", char, cursorPos, len(charRunes))
}

// SetTargetEdit sets the target line edit for character insertion
func (up *UnicodePicker) SetTargetEdit(edit *widgets.QLineEdit) {
	up.targetEdit = edit
	up.logger.Debug("Set target edit widget")
}

// ShowPicker shows the Unicode picker
func (up *UnicodePicker) ShowPicker() {
	up.logger.Debug("ShowPicker called - using working test pattern")

	// Use same pattern as working test
	up.Show()
	up.Raise()
	up.ActivateWindow()
	up.isVisible = true

	up.logger.Success("Unicode picker shown using working test pattern")
}

// HidePicker hides the Unicode picker
func (up *UnicodePicker) HidePicker() {
	up.logger.Debug("HidePicker called")
	up.Hide()
	up.isVisible = false
	up.logger.Debug("Unicode picker hidden")
}

// ToggleVisibility toggles picker visibility
func (up *UnicodePicker) ToggleVisibility() {
	if up.isVisible {
		up.HidePicker()
	} else {
		up.ShowPicker()
	}
}

// ReloadConfig reloads the configuration file
func (up *UnicodePicker) ReloadConfig() {
	up.loadConfig()

	// Clear existing tabs
	up.tabWidget.Clear()

	// Recreate tabs with new config
	up.createCharacterTabs()

	up.logger.Success("Reloaded Unicode picker configuration")
}

// GetConfigPath returns the current configuration file path
func (up *UnicodePicker) GetConfigPath() string {
	return up.configPath
}

// IsVisible returns whether the picker is currently visible
func (up *UnicodePicker) IsVisible() bool {
	return up.isVisible
}

// PositionNearWidget positions the picker near another widget (simplified to center positioning)
func (up *UnicodePicker) PositionNearWidget(widget widgets.QWidget_ITF) {
	up.logger.Debug("Positioning Unicode picker at screen center")

	// Use same center positioning as working test
	screen := widgets.QApplication_Desktop().AvailableGeometry2(nil)
	centerX := screen.Width()/2 - up.Width()/2
	centerY := screen.Height()/2 - up.Height()/2

	up.Move2(centerX, centerY)
	up.logger.Debug("Positioned Unicode picker at center of screen")
}
