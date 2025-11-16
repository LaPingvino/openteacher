// Package ocrgui provides functionality ported from Python module
//
// Provides a wizard that assists the user in loading a word list
// from a picture or scan of a real-world word list. (E.g. out of a
// book). The actual logic is in the ocrWordListLoader module,
// although some work around that (making the image ready) is done
// by this module.
//
// This is an automated port - implementation may be incomplete.
package ocrgui

import (
	"context"
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/core"
	_ "github.com/therecipe/qt/widgets"
)

// OcrGuiModule is a Go port of the Python OcrGuiModule class
type OcrGuiModule struct {
	*core.BaseModule
	manager *core.Manager
	// TODO: Add module-specific fields
}

// NewOcrGuiModule creates a new OcrGuiModule instance
func NewOcrGuiModule() *OcrGuiModule {
	base := core.NewBaseModule("ui", "ocrgui-module")

	return &OcrGuiModule{
		BaseModule: base,
	}
}

// startwizard is the Go port of the Python _startWizard method
func (mod *OcrGuiModule) startwizard() {
	// TODO: Port Python method logic
}

// loadresultivelesson is the Go port of the Python _loadResultiveLesson method
func (mod *OcrGuiModule) loadresultivelesson() {
	// TODO: Port Python method logic
}

// retranslate is the Go port of the Python _retranslate method
func (mod *OcrGuiModule) retranslate() {
	// TODO: Port Python method logic
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (mod *OcrGuiModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// TODO: Port Python enable logic

	fmt.Println("OcrGuiModule enabled")
	return nil
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (mod *OcrGuiModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// TODO: Port Python disable logic

	fmt.Println("OcrGuiModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *OcrGuiModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitOcrGuiModule creates and returns a new OcrGuiModule instance
// This is the Go equivalent of the Python init function
func InitOcrGuiModule() core.Module {
	return NewOcrGuiModule()
}