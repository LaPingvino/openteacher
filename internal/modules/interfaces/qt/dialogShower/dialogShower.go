// Package dialogshower provides functionality ported from Python module
//
// Provides a service for showing various dialogs throughout the application.
// This module acts as a central hub for dialog management.
//
// This is an automated port - implementation may be incomplete.
package dialogshower

import (
	"context"
	"fmt"

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// DialogShowerModule is a Go port of the Python DialogShowerModule class
type DialogShowerModule struct {
	*core.BaseModule
	manager      *core.Manager
	mainWindow   *widgets.QMainWindow
	messageBoxes []*widgets.QMessageBox
}

// NewDialogShowerModule creates a new DialogShowerModule instance
func NewDialogShowerModule() *DialogShowerModule {
	base := core.NewBaseModule("ui", "dialog-shower-module")
	base.SetRequires("qtApp")

	return &DialogShowerModule{
		BaseModule:   base,
		messageBoxes: make([]*widgets.QMessageBox, 0),
	}
}

// SetMainWindow sets the main window to use as parent for dialogs
func (mod *DialogShowerModule) SetMainWindow(window *widgets.QMainWindow) {
	mod.mainWindow = window
}

// ShowInformation displays an information dialog
func (mod *DialogShowerModule) ShowInformation(title string, message string) {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Information)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	msgBox.Exec()
	mod.removeMessageBox(msgBox)
}

// ShowWarning displays a warning dialog
func (mod *DialogShowerModule) ShowWarning(title string, message string) {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Warning)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	msgBox.Exec()
	mod.removeMessageBox(msgBox)
}

// ShowError displays an error dialog
func (mod *DialogShowerModule) ShowError(title string, message string) {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Critical)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)
	msgBox.SetStandardButtons(widgets.QMessageBox__Ok)

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	msgBox.Exec()
	mod.removeMessageBox(msgBox)
}

// ShowQuestion displays a yes/no question dialog and returns true if Yes was clicked
func (mod *DialogShowerModule) ShowQuestion(title string, message string) bool {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Question)
	msgBox.SetStandardButtons(widgets.QMessageBox__Yes | widgets.QMessageBox__No)
	msgBox.SetDefaultButton2(widgets.QMessageBox__No)

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	result := msgBox.Exec()
	mod.removeMessageBox(msgBox)

	return result == int(widgets.QMessageBox__Yes)
}

// ShowQuestionWithCancel displays a yes/no/cancel question dialog
// Returns: 1 for Yes, 0 for No, -1 for Cancel
func (mod *DialogShowerModule) ShowQuestionWithCancel(title string, message string) int {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Question)
	msgBox.SetStandardButtons(widgets.QMessageBox__Yes | widgets.QMessageBox__No | widgets.QMessageBox__Cancel)
	msgBox.SetDefaultButton2(widgets.QMessageBox__Cancel)

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	result := msgBox.Exec()
	mod.removeMessageBox(msgBox)

	switch result {
	case int(widgets.QMessageBox__Yes):
		return 1
	case int(widgets.QMessageBox__No):
		return 0
	default: // Cancel or close
		return -1
	}
}

// ShowCustomDialog displays a custom dialog with specified buttons
func (mod *DialogShowerModule) ShowCustomDialog(title string, message string, buttons []string, defaultButton int) int {
	msgBox := widgets.NewQMessageBox(mod.mainWindow)
	msgBox.SetWindowTitle(title)
	msgBox.SetText(message)
	msgBox.SetIcon(widgets.QMessageBox__Question)

	// Add custom buttons
	buttonRefs := make([]*widgets.QPushButton, len(buttons))
	for i, buttonText := range buttons {
		button := msgBox.AddButton2(buttonText, widgets.QMessageBox__ActionRole)
		buttonRefs[i] = button
		if i == defaultButton {
			msgBox.SetDefaultButton3(button)
		}
	}

	mod.messageBoxes = append(mod.messageBoxes, msgBox)
	msgBox.Exec()

	// Find which button was clicked
	clickedButton := msgBox.ClickedButton()
	result := -1
	for i, button := range buttonRefs {
		if button.Pointer() == clickedButton.Pointer() {
			result = i
			break
		}
	}

	mod.removeMessageBox(msgBox)
	return result
}

// ShowInputDialog displays an input dialog and returns the entered text
func (mod *DialogShowerModule) ShowInputDialog(title string, label string, defaultText string) (string, bool) {
	ok := false
	text := widgets.QInputDialog_GetText(mod.mainWindow, title, label, widgets.QLineEdit__Normal, defaultText, &ok, 0, 0)
	return text, ok
}

// ShowPasswordDialog displays a password input dialog
func (mod *DialogShowerModule) ShowPasswordDialog(title string, label string) (string, bool) {
	ok := false
	text := widgets.QInputDialog_GetText(mod.mainWindow, title, label, widgets.QLineEdit__Password, "", &ok, 0, 0)
	return text, ok
}

// ShowProgressDialog creates and returns a progress dialog
func (mod *DialogShowerModule) ShowProgressDialog(title string, label string, maximum int) *widgets.QProgressDialog {
	progressDialog := widgets.NewQProgressDialog2(label, "Cancel", 0, maximum, mod.mainWindow, 0)
	progressDialog.SetWindowTitle(title)
	progressDialog.SetWindowModality(qtcore.Qt__ApplicationModal)
	progressDialog.SetMinimumDuration(0)
	progressDialog.Show()
	return progressDialog
}

// ShowFileDialog displays a file selection dialog
func (mod *DialogShowerModule) ShowFileDialog(title string, filter string, save bool) string {
	if save {
		return widgets.QFileDialog_GetSaveFileName(mod.mainWindow, title, "", filter, "", 0)
	} else {
		return widgets.QFileDialog_GetOpenFileName(mod.mainWindow, title, "", filter, "", 0)
	}
}

// ShowDirectoryDialog displays a directory selection dialog
func (mod *DialogShowerModule) ShowDirectoryDialog(title string) string {
	return widgets.QFileDialog_GetExistingDirectory(mod.mainWindow, title, "", 0)
}

// ShowAbout displays an about dialog with application information
func (mod *DialogShowerModule) ShowAbout(title string, text string) {
	widgets.QMessageBox_About(mod.mainWindow, title, text)
}

// ShowAboutQt displays the standard Qt about dialog
func (mod *DialogShowerModule) ShowAboutQt() {
	widgets.QMessageBox_AboutQt(mod.mainWindow, "OpenTeacher")
}

// CloseAllDialogs closes all open message boxes
func (mod *DialogShowerModule) CloseAllDialogs() {
	for _, msgBox := range mod.messageBoxes {
		if msgBox != nil {
			msgBox.Close()
		}
	}
	mod.messageBoxes = mod.messageBoxes[:0] // Clear the slice
}

// removeMessageBox removes a message box from tracking
func (mod *DialogShowerModule) removeMessageBox(msgBox *widgets.QMessageBox) {
	for i, box := range mod.messageBoxes {
		if box == msgBox {
			// Remove from slice
			mod.messageBoxes = append(mod.messageBoxes[:i], mod.messageBoxes[i+1:]...)
			break
		}
	}
}

// Enable activates the module
func (mod *DialogShowerModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Try to get the main window from the gui module
	if mod.manager != nil {
		if guiModule, exists := mod.manager.GetDefaultModule("ui"); exists {
			if guiMod, ok := guiModule.(interface{ GetMainWindow() *widgets.QMainWindow }); ok {
				mod.mainWindow = guiMod.GetMainWindow()
			}
		}
	}

	fmt.Println("DialogShowerModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *DialogShowerModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Close all open dialogs
	mod.CloseAllDialogs()
	mod.mainWindow = nil

	fmt.Println("DialogShowerModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *DialogShowerModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitDialogShowerModule creates and returns a new DialogShowerModule instance
func InitDialogShowerModule() core.Module {
	return NewDialogShowerModule()
}
