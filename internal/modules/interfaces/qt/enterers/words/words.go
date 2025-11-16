// Package words provides functionality ported from Python module
//
// Provides a word list enterer widget for creating and editing word pairs.
// This module allows users to enter question/answer pairs for vocabulary lessons.
//
// This is an automated port - implementation may be incomplete.
package words

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// WordPair represents a question/answer pair
type WordPair struct {
	Question string
	Answer   string
	Comment  string
}

// WordsEntererModule is a Go port of the Python WordsEntererModule class
type WordsEntererModule struct {
	*core.BaseModule
	manager     *core.Manager
	widget      *qt.QWidget
	tableView   *qt.QTableWidget
	addButton   *qt.QPushButton
	delButton   *qt.QPushButton
	clearButton *qt.QPushButton
	wordPairs   []WordPair
}

// NewWordsEntererModule creates a new WordsEntererModule instance
func NewWordsEntererModule() *WordsEntererModule {
	base := core.NewBaseModule("ui", "words-enterer-module")
	base.SetRequires("qtApp")

	return &WordsEntererModule{
		BaseModule: base,
		wordPairs:  make([]WordPair, 0),
	}
}

// GetWidget returns the main widget for this enterer
func (mod *WordsEntererModule) GetWidget() *qt.QWidget {
	if mod.widget == nil {
		mod.createWidget()
	}
	return mod.widget
}

// createWidget creates the main word enterer widget
func (mod *WordsEntererModule) createWidget() {
	mod.widget = qt.NewQWidget(nil, 0)
	layout := qt.NewQVBoxLayout()
	mod.widget.SetLayout(layout)

	// Header
	headerLabel := qt.NewQLabel2("Word List Editor", nil, 0)
	headerFont := headerLabel.Font()
	headerFont.SetBold(true)
	headerFont.SetPointSize(14)
	headerLabel.SetFont(headerFont)
	layout.AddWidget(headerLabel, 0, 0)

	// Instructions
	instructionsLabel := qt.NewQLabel2("Enter word pairs below. Click 'Add Row' to add new pairs.", nil, 0)
	instructionsLabel.SetWordWrap(true)
	layout.AddWidget(instructionsLabel, 0, 0)

	// Table widget
	mod.tableView = qt.NewQTableWidget3(0, 3, nil)
	mod.tableView.SetHorizontalHeaderLabels([]string{"Question", "Answer", "Comment"})

	// Set column widths
	header := mod.tableView.HorizontalHeader()
	header.SetStretchLastSection(false)
	header.ResizeSection(0, 200) // Question column
	header.ResizeSection(1, 200) // Answer column
	header.ResizeSection(2, 150) // Comment column

	// Enable sorting
	mod.tableView.SetSortingEnabled(true)
	mod.tableView.SetAlternatingRowColors(true)
	mod.tableView.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)

	layout.AddWidget(mod.tableView, 1, 0)

	// Button layout
	buttonLayout := qt.NewQHBoxLayout()

	mod.addButton = qt.NewQPushButton2("Add Row", nil)
	mod.addButton.ConnectClicked(func(checked bool) {
		mod.addWordPair()
	})
	buttonLayout.AddWidget(mod.addButton, 0, 0)

	mod.delButton = qt.NewQPushButton2("Delete Selected", nil)
	mod.delButton.ConnectClicked(func(checked bool) {
		mod.deleteSelectedRows()
	})
	buttonLayout.AddWidget(mod.delButton, 0, 0)

	mod.clearButton = qt.NewQPushButton2("Clear All", nil)
	mod.clearButton.ConnectClicked(func(checked bool) {
		mod.clearAllRows()
	})
	buttonLayout.AddWidget(mod.clearButton, 0, 0)

	buttonLayout.AddStretch(1)

	// Import/Export buttons
	importButton := qt.NewQPushButton2("Import from Text", nil)
	importButton.ConnectClicked(func(checked bool) {
		mod.showImportDialog()
	})
	buttonLayout.AddWidget(importButton, 0, 0)

	exportButton := qt.NewQPushButton2("Export to Text", nil)
	exportButton.ConnectClicked(func(checked bool) {
		mod.showExportDialog()
	})
	buttonLayout.AddWidget(exportButton, 0, 0)

	layout.AddLayout(buttonLayout, 0)

	// Connect table signals
	mod.tableView.ConnectCellChanged(func(row, column int) {
		mod.updateWordPairFromTable(row)
	})

	// Add some initial rows
	mod.addWordPair()
	mod.addWordPair()
	mod.addWordPair()
}

// addWordPair adds a new word pair row to the table
func (mod *WordsEntererModule) addWordPair() {
	if mod.tableView == nil {
		return
	}

	row := mod.tableView.RowCount()
	mod.tableView.InsertRow(row)

	// Create editable items
	questionItem := qt.NewQTableWidgetItem2("", 0)
	answerItem := qt.NewQTableWidgetItem2("", 0)
	commentItem := qt.NewQTableWidgetItem2("", 0)

	mod.tableView.SetItem(row, 0, questionItem)
	mod.tableView.SetItem(row, 1, answerItem)
	mod.tableView.SetItem(row, 2, commentItem)

	// Add to internal storage
	mod.wordPairs = append(mod.wordPairs, WordPair{})

	// Focus on the new row's first column
	mod.tableView.SetCurrentCell(row, 0)
}

// deleteSelectedRows deletes the currently selected rows
func (mod *WordsEntererModule) deleteSelectedRows() {
	if mod.tableView == nil {
		return
	}

	selectedRanges := mod.tableView.SelectedRanges()
	if len(selectedRanges) == 0 {
		return
	}

	// Get all selected row numbers
	selectedRows := make([]int, 0)
	for _, selRange := range selectedRanges {
		for row := selRange.TopRow(); row <= selRange.BottomRow(); row++ {
			selectedRows = append(selectedRows, row)
		}
	}

	// Remove duplicates and sort in descending order
	uniqueRows := make(map[int]bool)
	for _, row := range selectedRows {
		uniqueRows[row] = true
	}

	sortedRows := make([]int, 0, len(uniqueRows))
	for row := range uniqueRows {
		sortedRows = append(sortedRows, row)
	}

	// Sort in descending order to remove from bottom up
	for i := 0; i < len(sortedRows)-1; i++ {
		for j := i + 1; j < len(sortedRows); j++ {
			if sortedRows[i] < sortedRows[j] {
				sortedRows[i], sortedRows[j] = sortedRows[j], sortedRows[i]
			}
		}
	}

	// Remove rows
	for _, row := range sortedRows {
		if row >= 0 && row < len(mod.wordPairs) {
			mod.tableView.RemoveRow(row)
			// Remove from internal storage
			mod.wordPairs = append(mod.wordPairs[:row], mod.wordPairs[row+1:]...)
		}
	}
}

// clearAllRows removes all word pairs
func (mod *WordsEntererModule) clearAllRows() {
	if mod.tableView == nil {
		return
	}

	// Ask for confirmation
	reply := qt.QMessageBox_Question(mod.widget,
		"Clear All",
		"Are you sure you want to clear all word pairs?",
		qt.QMessageBox__Yes|qt.QMessageBox__No,
		qt.QMessageBox__No)

	if reply == qt.QMessageBox__Yes {
		mod.tableView.SetRowCount(0)
		mod.wordPairs = make([]WordPair, 0)

		// Add one empty row
		mod.addWordPair()
	}
}

// updateWordPairFromTable updates internal storage when table cell changes
func (mod *WordsEntererModule) updateWordPairFromTable(row int) {
	if mod.tableView == nil || row >= len(mod.wordPairs) {
		return
	}

	questionItem := mod.tableView.Item(row, 0)
	answerItem := mod.tableView.Item(row, 1)
	commentItem := mod.tableView.Item(row, 2)

	if questionItem != nil && answerItem != nil && commentItem != nil {
		mod.wordPairs[row].Question = strings.TrimSpace(questionItem.Text())
		mod.wordPairs[row].Answer = strings.TrimSpace(answerItem.Text())
		mod.wordPairs[row].Comment = strings.TrimSpace(commentItem.Text())
	}
}

// showImportDialog shows a dialog for importing word pairs from text
func (mod *WordsEntererModule) showImportDialog() {
	dialog := qt.NewQDialog(mod.widget, 0)
	dialog.SetWindowTitle("Import Word Pairs")
	dialog.SetFixedSize2(500, 400)
	dialog.SetWindowModality(core.Qt__ApplicationModal)

	layout := qt.NewQVBoxLayout()
	dialog.SetLayout(layout)

	// Instructions
	instructLabel := qt.NewQLabel2("Enter word pairs, one per line, separated by tabs or commas:", nil, 0)
	layout.AddWidget(instructLabel, 0, 0)

	// Text area
	textEdit := qt.NewQTextEdit(nil)
	textEdit.SetPlaceholderText("hello\tworld\tgreeting\ngoodbye\tauf wiedersehen\tparting")
	layout.AddWidget(textEdit, 1, 0)

	// Separator options
	sepLayout := qt.NewQHBoxLayout()
	sepLabel := qt.NewQLabel2("Separator:", nil, 0)
	sepLayout.AddWidget(sepLabel, 0, 0)

	sepCombo := qt.NewQComboBox(nil)
	sepCombo.AddItems([]string{"Tab", "Comma", "Semicolon", "Pipe (|)"})
	sepLayout.AddWidget(sepCombo, 0, 0)

	sepLayout.AddStretch(1)
	layout.AddLayout(sepLayout, 0)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox3(
		qt.QDialogButtonBox__Ok|qt.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		text := textEdit.ToPlainText()
		separator := sepCombo.CurrentText()
		mod.importFromText(text, separator)
		dialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		dialog.Reject()
	})

	dialog.Exec()
}

// showExportDialog shows a dialog for exporting word pairs to text
func (mod *WordsEntererModule) showExportDialog() {
	// Gather current word pairs
	mod.updateAllWordPairsFromTable()

	if len(mod.wordPairs) == 0 {
		qt.QMessageBox_Information(mod.widget, "Export", "No word pairs to export.", qt.QMessageBox__Ok, qt.QMessageBox__Ok)
		return
	}

	dialog := qt.NewQDialog(mod.widget, 0)
	dialog.SetWindowTitle("Export Word Pairs")
	dialog.SetFixedSize2(500, 400)
	dialog.SetWindowModality(core.Qt__ApplicationModal)

	layout := qt.NewQVBoxLayout()
	dialog.SetLayout(layout)

	// Instructions
	instructLabel := qt.NewQLabel2("Word pairs exported as text:", nil, 0)
	layout.AddWidget(instructLabel, 0, 0)

	// Text area
	textEdit := qt.NewQTextEdit(nil)
	textEdit.SetReadOnly(true)

	// Generate export text
	exportText := ""
	for _, pair := range mod.wordPairs {
		if pair.Question != "" || pair.Answer != "" {
			line := pair.Question + "\t" + pair.Answer
			if pair.Comment != "" {
				line += "\t" + pair.Comment
			}
			exportText += line + "\n"
		}
	}
	textEdit.SetPlainText(exportText)
	layout.AddWidget(textEdit, 1, 0)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox3(
		qt.QDialogButtonBox__Close,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	// Add copy button
	copyButton := qt.NewQPushButton2("Copy to Clipboard", nil)
	copyButton.ConnectClicked(func(checked bool) {
		clipboard := qt.QApplication_Clipboard()
		clipboard.SetText(exportText, qt.QClipboard__Clipboard)
		qt.QMessageBox_Information(dialog, "Copied", "Text copied to clipboard!", qt.QMessageBox__Ok, qt.QMessageBox__Ok)
	})
	buttonBox.AddButton(copyButton, qt.QDialogButtonBox__ActionRole)

	buttonBox.ConnectRejected(func() {
		dialog.Close()
	})

	dialog.Exec()
}

// importFromText imports word pairs from text string
func (mod *WordsEntererModule) importFromText(text string, separatorType string) {
	if text == "" {
		return
	}

	// Determine separator
	separator := "\t"
	switch separatorType {
	case "Comma":
		separator = ","
	case "Semicolon":
		separator = ";"
	case "Pipe (|)":
		separator = "|"
	}

	// Clear existing data
	mod.tableView.SetRowCount(0)
	mod.wordPairs = make([]WordPair, 0)

	// Parse lines
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, separator)
		pair := WordPair{}

		if len(parts) >= 1 {
			pair.Question = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			pair.Answer = strings.TrimSpace(parts[1])
		}
		if len(parts) >= 3 {
			pair.Comment = strings.TrimSpace(parts[2])
		}

		// Add to table
		row := mod.tableView.RowCount()
		mod.tableView.InsertRow(row)

		questionItem := qt.NewQTableWidgetItem2(pair.Question, 0)
		answerItem := qt.NewQTableWidgetItem2(pair.Answer, 0)
		commentItem := qt.NewQTableWidgetItem2(pair.Comment, 0)

		mod.tableView.SetItem(row, 0, questionItem)
		mod.tableView.SetItem(row, 1, answerItem)
		mod.tableView.SetItem(row, 2, commentItem)

		mod.wordPairs = append(mod.wordPairs, pair)
	}

	// Add one empty row at the end
	mod.addWordPair()
}

// updateAllWordPairsFromTable updates all word pairs from current table state
func (mod *WordsEntererModule) updateAllWordPairsFromTable() {
	if mod.tableView == nil {
		return
	}

	rowCount := mod.tableView.RowCount()
	mod.wordPairs = make([]WordPair, rowCount)

	for row := 0; row < rowCount; row++ {
		mod.updateWordPairFromTable(row)
	}
}

// GetWordPairs returns the current list of word pairs
func (mod *WordsEntererModule) GetWordPairs() []WordPair {
	mod.updateAllWordPairsFromTable()
	return mod.wordPairs
}

// SetWordPairs sets the word pairs and updates the table
func (mod *WordsEntererModule) SetWordPairs(pairs []WordPair) {
	if mod.tableView == nil {
		mod.wordPairs = pairs
		return
	}

	// Clear existing table
	mod.tableView.SetRowCount(0)
	mod.wordPairs = make([]WordPair, 0)

	// Add pairs
	for _, pair := range pairs {
		row := mod.tableView.RowCount()
		mod.tableView.InsertRow(row)

		questionItem := qt.NewQTableWidgetItem2(pair.Question, 0)
		answerItem := qt.NewQTableWidgetItem2(pair.Answer, 0)
		commentItem := qt.NewQTableWidgetItem2(pair.Comment, 0)

		mod.tableView.SetItem(row, 0, questionItem)
		mod.tableView.SetItem(row, 1, answerItem)
		mod.tableView.SetItem(row, 2, commentItem)

		mod.wordPairs = append(mod.wordPairs, pair)
	}

	// Add one empty row
	mod.addWordPair()
}

// GetWordCount returns the number of non-empty word pairs
func (mod *WordsEntererModule) GetWordCount() int {
	mod.updateAllWordPairsFromTable()
	count := 0
	for _, pair := range mod.wordPairs {
		if pair.Question != "" || pair.Answer != "" {
			count++
		}
	}
	return count
}

// Enable activates the module
func (mod *WordsEntererModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("WordsEntererModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *WordsEntererModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up widget
	if mod.widget != nil {
		mod.widget.Close()
		mod.widget = nil
		mod.tableView = nil
		mod.addButton = nil
		mod.delButton = nil
		mod.clearButton = nil
	}

	fmt.Println("WordsEntererModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *WordsEntererModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitWordsEntererModule creates and returns a new WordsEntererModule instance
func InitWordsEntererModule() core.Module {
	return NewWordsEntererModule()
}
