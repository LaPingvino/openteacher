// Package results provides functionality ported from Python module
//
// Provides a results dialog for displaying test/lesson results and statistics.
//
// This is an automated port - implementation may be incomplete.
package results

import (
	"context"
	"fmt"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/therecipe/qt/widgets"
)

// ResultsDialogModule is a Go port of the Python ResultsDialogModule class
type ResultsDialogModule struct {
	*core.BaseModule
	manager *core.Manager
	dialog  *widgets.QDialog
}

// NewResultsDialogModule creates a new ResultsDialogModule instance
func NewResultsDialogModule() *ResultsDialogModule {
	base := core.NewBaseModule("ui", "results-dialog-module")
	base.SetRequires("qtApp")

	return &ResultsDialogModule{
		BaseModule: base,
	}
}

// Show displays the results dialog with given data
func (mod *ResultsDialogModule) Show(results map[string]interface{}) {
	if mod.dialog == nil {
		mod.createDialog()
	}

	if mod.dialog != nil {
		mod.loadResults(results)
		mod.dialog.Show()
		mod.dialog.Raise()
		mod.dialog.ActivateWindow()
	}
}

// createDialog creates the results dialog
func (mod *ResultsDialogModule) createDialog() {
	mod.dialog = widgets.NewQDialog(nil, 0)
	mod.dialog.SetWindowTitle("Test Results")
	mod.dialog.SetFixedSize2(500, 400)

	layout := widgets.NewQVBoxLayout()
	mod.dialog.SetLayout(layout)

	// Results display
	resultsText := widgets.NewQTextEdit(nil)
	resultsText.SetReadOnly(true)
	resultsText.SetObjectName("resultsText")
	layout.AddWidget(resultsText, 1, 0)

	// Close button
	buttonBox := widgets.NewQDialogButtonBox3(widgets.QDialogButtonBox__Close, nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectRejected(func() {
		mod.dialog.Close()
	})
}

// loadResults loads results data into the dialog
func (mod *ResultsDialogModule) loadResults(results map[string]interface{}) {
	if mod.dialog == nil {
		return
	}

	resultsText := widgets.NewQTextEdit(nil)
	resultsText.SetObjectName("resultsText")

	// For now, just print results to console to avoid Qt API issues
	fmt.Printf("Results loaded: %+v\n", results)

	// Simple implementation without widget finding
	text := "Test Results\n"
	text += "============\n\n"

	if score, ok := results["score"].(float64); ok {
		text += fmt.Sprintf("Score: %.1f%%\n", score)
	}

	if correct, ok := results["correct"].(int); ok {
		text += fmt.Sprintf("Correct answers: %d\n", correct)
	}

	if total, ok := results["total"].(int); ok {
		text += fmt.Sprintf("Total questions: %d\n", total)
	}

	fmt.Printf("Results text: %s\n", text)
	return

}

// Enable activates the module
func (mod *ResultsDialogModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("ResultsDialogModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *ResultsDialogModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	if mod.dialog != nil {
		mod.dialog.Close()
		mod.dialog = nil
	}

	fmt.Println("ResultsDialogModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *ResultsDialogModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitResultsDialogModule creates and returns a new ResultsDialogModule instance
func InitResultsDialogModule() core.Module {
	return NewResultsDialogModule()
}
