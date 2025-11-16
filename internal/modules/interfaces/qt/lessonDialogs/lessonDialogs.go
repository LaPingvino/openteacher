// Package lessonDialogs provides functionality ported from Python module
//
// Provides dialogs for lesson management including new lesson creation,
// lesson properties, and lesson import/export dialogs.
//
// This is an automated port - implementation may be incomplete.
package lessonDialogs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// LessonDialogsModule is a Go port of the Python LessonDialogsModule class
type LessonDialogsModule struct {
	*core.BaseModule
	manager          *core.Manager
	newLessonDialog  *qt.QDialog
	propertiesDialog *qt.QDialog
	importDialog     *qt.QDialog
}

// NewLessonDialogsModule creates a new LessonDialogsModule instance
func NewLessonDialogsModule() *LessonDialogsModule {
	base := core.NewBaseModule("lessonDialogs", "lesson-dialogs-module")
	base.SetRequires("qtApp")

	return &LessonDialogsModule{
		BaseModule: base,
	}
}

// ShowNewLessonDialog displays the new lesson creation dialog
func (mod *LessonDialogsModule) ShowNewLessonDialog() map[string]interface{} {
	log.Printf("[SUCCESS] LessonDialogsModule.ShowNewLessonDialog() - creating and showing new lesson dialog")

	if mod.manager == nil {
		log.Printf("[ERROR] LessonDialogsModule.ShowNewLessonDialog() - manager is nil")
		return nil
	}

	// Get the main window as parent
	var parentWidget *qt.QWidget
	uiModules := mod.manager.GetModulesByType("ui")
	if len(uiModules) > 0 {
		if guiMod, ok := uiModules[0].(interface{ GetMainWindow() *qt.QMainWindow }); ok {
			parentWidget = guiMod.GetMainWindow().QWidget_PTR()
			log.Printf("[SUCCESS] LessonDialogsModule got parent window from GUI module")
		}
	}

	if mod.newLessonDialog == nil {
		mod.createNewLessonDialog(parentWidget)
	}

	// Reset form
	mod.resetNewLessonForm()

	log.Printf("[SUCCESS] LessonDialogsModule showing new lesson dialog")
	result := mod.newLessonDialog.Exec()
	log.Printf("[SUCCESS] LessonDialogsModule dialog closed with result: %d", result)

	if result == int(qt.QDialog__Accepted) {
		data := mod.getNewLessonData()
		log.Printf("[SUCCESS] New lesson dialog returned data: %v", data)
		return data
	} else {
		log.Printf("[INFO] New lesson dialog was cancelled")
	}

	return nil
}

// ShowPropertiesDialog displays the lesson properties dialog
func (mod *LessonDialogsModule) ShowPropertiesDialog(parent *qt.QWidget, lessonData map[string]interface{}) map[string]interface{} {
	if mod.propertiesDialog == nil {
		mod.createPropertiesDialog(parent)
	}

	// Load current lesson data
	mod.loadPropertiesData(lessonData)

	if mod.propertiesDialog.Exec() == int(qt.QDialog__Accepted) {
		return mod.getPropertiesData()
	}

	return nil
}

// ShowImportDialog displays the lesson import dialog
func (mod *LessonDialogsModule) ShowImportDialog(parent *qt.QWidget) map[string]interface{} {
	if mod.importDialog == nil {
		mod.createImportDialog(parent)
	}

	if mod.importDialog.Exec() == int(qt.QDialog__Accepted) {
		return mod.getImportData()
	}

	return nil
}

// createNewLessonDialog creates the new lesson dialog
func (mod *LessonDialogsModule) createNewLessonDialog(parent *qt.QWidget) {
	mod.newLessonDialog = qt.NewQDialog(parent, 0)
	mod.newLessonDialog.SetWindowTitle("Create New Lesson")
	mod.newLessonDialog.SetFixedSize2(400, 350)
	mod.newLessonDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout()
	mod.newLessonDialog.SetLayout(layout)

	// Lesson name
	nameGroup := qt.NewQGroupBox2("Lesson Information", nil)
	nameLayout := qt.NewQFormLayout(nil)
	nameGroup.SetLayout(nameLayout)

	nameEdit := qt.NewQLineEdit(nil)
	nameEdit.SetObjectName("lessonName")
	nameEdit.SetPlaceholderText("Enter lesson name...")
	nameLayout.AddRow3("Name:", nameEdit)

	descEdit := qt.NewQTextEdit(nil)
	descEdit.SetObjectName("lessonDescription")
	descEdit.SetPlaceholderText("Enter lesson description...")
	descEdit.SetMaximumHeight(80)
	nameLayout.AddRow3("Description:", descEdit)

	layout.AddWidget(nameGroup, 0, 0)

	// Lesson type
	typeGroup := qt.NewQGroupBox2("Lesson Type", nil)
	typeLayout := qt.NewQVBoxLayout()
	typeGroup.SetLayout(typeLayout)

	wordsRadio := qt.NewQRadioButton2("Words List", nil)
	wordsRadio.SetObjectName("wordsRadio")
	wordsRadio.SetChecked(true)
	typeLayout.AddWidget(wordsRadio, 0, 0)

	topoRadio := qt.NewQRadioButton2("Topology", nil)
	topoRadio.SetObjectName("topoRadio")
	typeLayout.AddWidget(topoRadio, 0, 0)

	mediaRadio := qt.NewQRadioButton2("Media", nil)
	mediaRadio.SetObjectName("mediaRadio")
	typeLayout.AddWidget(mediaRadio, 0, 0)

	layout.AddWidget(typeGroup, 0, 0)

	// Language settings
	langGroup := qt.NewQGroupBox2("Languages", nil)
	langLayout := qt.NewQFormLayout(nil)
	langGroup.SetLayout(langLayout)

	questionLangCombo := qt.NewQComboBox(nil)
	questionLangCombo.SetObjectName("questionLanguage")
	questionLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	langLayout.AddRow3("Question language:", questionLangCombo)

	answerLangCombo := qt.NewQComboBox(nil)
	answerLangCombo.SetObjectName("answerLanguage")
	answerLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	answerLangCombo.SetCurrentIndex(1) // Default to Dutch
	langLayout.AddRow3("Answer language:", answerLangCombo)

	layout.AddWidget(langGroup, 0, 0)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox3(
		qt.QDialogButtonBox__Ok|qt.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.newLessonDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.newLessonDialog.Reject()
	})
}

// createPropertiesDialog creates the lesson properties dialog
func (mod *LessonDialogsModule) createPropertiesDialog(parent *qt.QWidget) {
	mod.propertiesDialog = qt.NewQDialog(parent, 0)
	mod.propertiesDialog.SetWindowTitle("Lesson Properties")
	mod.propertiesDialog.SetFixedSize2(450, 400)
	mod.propertiesDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout()
	mod.propertiesDialog.SetLayout(layout)

	// Create tab widget
	tabWidget := qt.NewQTabWidget(nil)
	layout.AddWidget(tabWidget, 1, 0)

	// General tab
	generalTab := qt.NewQWidget(nil, 0)
	generalLayout := qt.NewQFormLayout(nil)
	generalTab.SetLayout(generalLayout)

	nameEdit := qt.NewQLineEdit(nil)
	nameEdit.SetObjectName("propLessonName")
	generalLayout.AddRow3("Name:", nameEdit)

	descEdit := qt.NewQTextEdit(nil)
	descEdit.SetObjectName("propLessonDescription")
	descEdit.SetMaximumHeight(100)
	generalLayout.AddRow3("Description:", descEdit)

	authorEdit := qt.NewQLineEdit(nil)
	authorEdit.SetObjectName("propAuthor")
	generalLayout.AddRow3("Author:", authorEdit)

	versionEdit := qt.NewQLineEdit(nil)
	versionEdit.SetObjectName("propVersion")
	generalLayout.AddRow3("Version:", versionEdit)

	tabWidget.AddTab(generalTab, "General")

	// Statistics tab
	statsTab := qt.NewQWidget(nil, 0)
	statsLayout := qt.NewQFormLayout(nil)
	statsTab.SetLayout(statsLayout)

	itemCountLabel := qt.NewQLabel2("0", nil, 0)
	itemCountLabel.SetObjectName("itemCount")
	statsLayout.AddRow3("Number of items:", itemCountLabel)

	createdLabel := qt.NewQLabel2("Unknown", nil, 0)
	createdLabel.SetObjectName("createdDate")
	statsLayout.AddRow3("Created:", createdLabel)

	modifiedLabel := qt.NewQLabel2("Unknown", nil, 0)
	modifiedLabel.SetObjectName("modifiedDate")
	statsLayout.AddRow3("Last modified:", modifiedLabel)

	fileSizeLabel := qt.NewQLabel2("0 KB", nil, 0)
	fileSizeLabel.SetObjectName("fileSize")
	statsLayout.AddRow3("File size:", fileSizeLabel)

	tabWidget.AddTab(statsTab, "Statistics")

	// Buttons
	buttonBox := qt.NewQDialogButtonBox3(
		qt.QDialogButtonBox__Ok|qt.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.propertiesDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.propertiesDialog.Reject()
	})
}

// createImportDialog creates the lesson import dialog
func (mod *LessonDialogsModule) createImportDialog(parent *qt.QWidget) {
	mod.importDialog = qt.NewQDialog(parent, 0)
	mod.importDialog.SetWindowTitle("Import Lesson")
	mod.importDialog.SetFixedSize2(500, 300)
	mod.importDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout()
	mod.importDialog.SetLayout(layout)

	// File selection
	fileGroup := qt.NewQGroupBox2("Source File", nil)
	fileLayout := qt.NewQHBoxLayout()
	fileGroup.SetLayout(fileLayout)

	fileEdit := qt.NewQLineEdit(nil)
	fileEdit.SetObjectName("importFile")
	fileEdit.SetPlaceholderText("Select file to import...")
	fileLayout.AddWidget(fileEdit, 1, 0)

	browseButton := qt.NewQPushButton2("Browse...", nil)
	browseButton.ConnectClicked(func(checked bool) {
		fileName := qt.QFileDialog_GetOpenFileName(mod.importDialog,
			"Select lesson file",
			"",
			"All supported files (*.ot *.txt *.csv);;OpenTeacher files (*.ot);;Text files (*.txt);;CSV files (*.csv);;All files (*.*)",
			"",
			0)
		if fileName != "" {
			fileEdit.SetText(fileName)
		}
	})
	fileLayout.AddWidget(browseButton, 0, 0)

	layout.AddWidget(fileGroup, 0, 0)

	// Import options
	optionsGroup := qt.NewQGroupBox2("Import Options", nil)
	optionsLayout := qt.NewQVBoxLayout()
	optionsGroup.SetLayout(optionsLayout)

	encodingLayout := qt.NewQHBoxLayout()
	encodingLabel := qt.NewQLabel2("File encoding:", nil, 0)
	encodingLayout.AddWidget(encodingLabel, 0, 0)

	encodingCombo := qt.NewQComboBox(nil)
	encodingCombo.SetObjectName("encoding")
	encodingCombo.AddItems([]string{"UTF-8", "UTF-16", "ISO-8859-1", "ASCII"})
	encodingLayout.AddWidget(encodingCombo, 1, 0)
	optionsLayout.AddLayout(encodingLayout, 0)

	separatorLayout := qt.NewQHBoxLayout()
	separatorLabel := qt.NewQLabel2("Field separator:", nil, 0)
	separatorLayout.AddWidget(separatorLabel, 0, 0)

	separatorCombo := qt.NewQComboBox(nil)
	separatorCombo.SetObjectName("separator")
	separatorCombo.AddItems([]string{"Tab", "Comma", "Semicolon", "Space"})
	separatorLayout.AddWidget(separatorCombo, 1, 0)
	optionsLayout.AddLayout(separatorLayout, 0)

	firstRowCheck := qt.NewQCheckBox2("First row contains headers", nil)
	firstRowCheck.SetObjectName("firstRowHeaders")
	firstRowCheck.SetChecked(true)
	optionsLayout.AddWidget(firstRowCheck, 0, 0)

	layout.AddWidget(optionsGroup, 0, 0)

	// Preview area
	previewGroup := qt.NewQGroupBox2("Preview", nil)
	previewLayout := qt.NewQVBoxLayout()
	previewGroup.SetLayout(previewLayout)

	previewText := qt.NewQTextEdit(nil)
	previewText.SetObjectName("preview")
	previewText.SetReadOnly(true)
	previewText.SetPlainText("Select a file to see preview...")
	previewLayout.AddWidget(previewText, 1, 0)

	layout.AddWidget(previewGroup, 1, 0)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox3(
		qt.QDialogButtonBox__Ok|qt.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.importDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.importDialog.Reject()
	})
}

// resetNewLessonForm resets the new lesson dialog form
func (mod *LessonDialogsModule) resetNewLessonForm() {
	if mod.newLessonDialog == nil {
		return
	}

	nameEdit := qt.NewQLineEditFromPointer(mod.newLessonDialog.FindChild("lessonName", qt.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		nameEdit.Clear()
	}

	descEdit := qt.NewQTextEditFromPointer(mod.newLessonDialog.FindChild("lessonDescription", qt.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		descEdit.Clear()
	}

	wordsRadio := qt.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("wordsRadio", qt.Qt__FindChildrenRecursively).Pointer())
	if wordsRadio != nil {
		wordsRadio.SetChecked(true)
	}
}

// getNewLessonData extracts data from the new lesson dialog
func (mod *LessonDialogsModule) getNewLessonData() map[string]interface{} {
	if mod.newLessonDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	nameEdit := qt.NewQLineEditFromPointer(mod.newLessonDialog.FindChild("lessonName", qt.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		data["name"] = strings.TrimSpace(nameEdit.Text())
	}

	descEdit := qt.NewQTextEditFromPointer(mod.newLessonDialog.FindChild("lessonDescription", qt.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		data["description"] = strings.TrimSpace(descEdit.ToPlainText())
	}

	// Determine lesson type
	wordsRadio := qt.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("wordsRadio", qt.Qt__FindChildrenRecursively).Pointer())
	topoRadio := qt.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("topoRadio", qt.Qt__FindChildrenRecursively).Pointer())
	mediaRadio := qt.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("mediaRadio", qt.Qt__FindChildrenRecursively).Pointer())

	if wordsRadio != nil && wordsRadio.IsChecked() {
		data["type"] = "words"
	} else if topoRadio != nil && topoRadio.IsChecked() {
		data["type"] = "topology"
	} else if mediaRadio != nil && mediaRadio.IsChecked() {
		data["type"] = "media"
	} else {
		data["type"] = "words" // default
	}

	// Get languages
	questionLang := qt.NewQComboBoxFromPointer(mod.newLessonDialog.FindChild("questionLanguage", qt.Qt__FindChildrenRecursively).Pointer())
	if questionLang != nil {
		data["questionLanguage"] = questionLang.CurrentText()
	}

	answerLang := qt.NewQComboBoxFromPointer(mod.newLessonDialog.FindChild("answerLanguage", qt.Qt__FindChildrenRecursively).Pointer())
	if answerLang != nil {
		data["answerLanguage"] = answerLang.CurrentText()
	}

	return data
}

// loadPropertiesData loads lesson data into the properties dialog
func (mod *LessonDialogsModule) loadPropertiesData(lessonData map[string]interface{}) {
	if mod.propertiesDialog == nil || lessonData == nil {
		return
	}

	if name, ok := lessonData["name"].(string); ok {
		nameEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propLessonName", qt.Qt__FindChildrenRecursively).Pointer())
		if nameEdit != nil {
			nameEdit.SetText(name)
		}
	}

	if desc, ok := lessonData["description"].(string); ok {
		descEdit := qt.NewQTextEditFromPointer(mod.propertiesDialog.FindChild("propLessonDescription", qt.Qt__FindChildrenRecursively).Pointer())
		if descEdit != nil {
			descEdit.SetPlainText(desc)
		}
	}

	if author, ok := lessonData["author"].(string); ok {
		authorEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propAuthor", qt.Qt__FindChildrenRecursively).Pointer())
		if authorEdit != nil {
			authorEdit.SetText(author)
		}
	}

	if version, ok := lessonData["version"].(string); ok {
		versionEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propVersion", qt.Qt__FindChildrenRecursively).Pointer())
		if versionEdit != nil {
			versionEdit.SetText(version)
		}
	}

	// Update statistics
	if itemCount, ok := lessonData["itemCount"].(int); ok {
		itemLabel := qt.NewQLabelFromPointer(mod.propertiesDialog.FindChild("itemCount", qt.Qt__FindChildrenRecursively).Pointer())
		if itemLabel != nil {
			itemLabel.SetText(fmt.Sprintf("%d", itemCount))
		}
	}
}

// getPropertiesData extracts data from the properties dialog
func (mod *LessonDialogsModule) getPropertiesData() map[string]interface{} {
	if mod.propertiesDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	nameEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propLessonName", qt.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		data["name"] = strings.TrimSpace(nameEdit.Text())
	}

	descEdit := qt.NewQTextEditFromPointer(mod.propertiesDialog.FindChild("propLessonDescription", qt.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		data["description"] = strings.TrimSpace(descEdit.ToPlainText())
	}

	authorEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propAuthor", qt.Qt__FindChildrenRecursively).Pointer())
	if authorEdit != nil {
		data["author"] = strings.TrimSpace(authorEdit.Text())
	}

	versionEdit := qt.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propVersion", qt.Qt__FindChildrenRecursively).Pointer())
	if versionEdit != nil {
		data["version"] = strings.TrimSpace(versionEdit.Text())
	}

	return data
}

// getImportData extracts data from the import dialog
func (mod *LessonDialogsModule) getImportData() map[string]interface{} {
	if mod.importDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	fileEdit := qt.NewQLineEditFromPointer(mod.importDialog.FindChild("importFile", qt.Qt__FindChildrenRecursively).Pointer())
	if fileEdit != nil {
		data["file"] = strings.TrimSpace(fileEdit.Text())
	}

	encodingCombo := qt.NewQComboBoxFromPointer(mod.importDialog.FindChild("encoding", qt.Qt__FindChildrenRecursively).Pointer())
	if encodingCombo != nil {
		data["encoding"] = encodingCombo.CurrentText()
	}

	separatorCombo := qt.NewQComboBoxFromPointer(mod.importDialog.FindChild("separator", qt.Qt__FindChildrenRecursively).Pointer())
	if separatorCombo != nil {
		data["separator"] = separatorCombo.CurrentText()
	}

	firstRowCheck := qt.NewQCheckBoxFromPointer(mod.importDialog.FindChild("firstRowHeaders", qt.Qt__FindChildrenRecursively).Pointer())
	if firstRowCheck != nil {
		data["firstRowHeaders"] = firstRowCheck.IsChecked()
	}

	return data
}

// Enable activates the module
func (mod *LessonDialogsModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("LessonDialogsModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *LessonDialogsModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up dialogs
	if mod.newLessonDialog != nil {
		mod.newLessonDialog.Close()
		mod.newLessonDialog = nil
	}

	if mod.propertiesDialog != nil {
		mod.propertiesDialog.Close()
		mod.propertiesDialog = nil
	}

	if mod.importDialog != nil {
		mod.importDialog.Close()
		mod.importDialog = nil
	}

	fmt.Println("LessonDialogsModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *LessonDialogsModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitLessonDialogsModule creates and returns a new LessonDialogsModule instance
func InitLessonDialogsModule() core.Module {
	return NewLessonDialogsModule()
}
