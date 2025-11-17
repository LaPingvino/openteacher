// Package lessonDialogs provides enhanced lesson dialog functionality
//
// This enhanced version provides comprehensive lesson management dialogs
// with improved validation, better user experience, and extended functionality.
package lessonDialogs

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// LessonType represents different types of lessons
type LessonType int

const (
	LessonTypeWords LessonType = iota
	LessonTypeTopo
	LessonTypeMedia
	LessonTypeCustom
)

// String returns string representation of lesson type
func (lt LessonType) String() string {
	switch lt {
	case LessonTypeWords:
		return "Words"
	case LessonTypeTopo:
		return "Topography"
	case LessonTypeMedia:
		return "Media"
	case LessonTypeCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// LessonDialogResult represents the result of a lesson dialog
type LessonDialogResult struct {
	Success         bool
	Title           string
	Description     string
	Author          string
	Version         string
	LessonType      LessonType
	QuestionLang    string
	AnswerLang      string
	FilePath        string
	ImportSettings  *ImportSettings
	ExportSettings  *ExportSettings
	ValidationError string
}

// ImportSettings represents import configuration
type ImportSettings struct {
	FilePath       string
	FileType       string
	Encoding       string
	Separator      string
	FirstRowHeader bool
	QuestionColumn int
	AnswerColumn   int
	CommentColumn  int
	SkipEmptyRows  bool
}

// ExportSettings represents export configuration
type ExportSettings struct {
	FilePath     string
	FileType     string
	IncludeStats bool
	IncludeTests bool
	Format       string
	Compression  bool
}

// EnhancedLessonDialogsModule provides comprehensive lesson dialog functionality
type EnhancedLessonDialogsModule struct {
	*core.BaseModule
	manager          *core.Manager
	supportedFormats map[string][]string
	recentFiles      []string
	lastDirectory    string
	languages        []Language
}

// Language represents a supported language
type Language struct {
	Code string
	Name string
}

// NewEnhancedLessonDialogsModule creates a new enhanced lesson dialogs module
func NewEnhancedLessonDialogsModule() *EnhancedLessonDialogsModule {
	return &EnhancedLessonDialogsModule{
		BaseModule: core.NewBaseModule("enhancedLessonDialogs", "Enhanced Lesson Dialogs"),
		supportedFormats: map[string][]string{
			"import": {".otwd", ".csv", ".txt", ".tsv", ".json", ".xml", ".kvtml"},
			"export": {".otwd", ".csv", ".txt", ".html", ".json", ".pdf"},
		},
		recentFiles: make([]string, 0, 10),
		languages:   getDefaultLanguages(),
	}
}

// Priority returns module loading priority
func (m *EnhancedLessonDialogsModule) Priority() int {
	return 180 // Load after core, before UI modules
}

// Requires returns module dependencies
func (m *EnhancedLessonDialogsModule) Requires() []string {
	return []string{"core", "dialogs"}
}

// Enable activates the module
func (m *EnhancedLessonDialogsModule) Enable(ctx context.Context) error {
	if err := m.BaseModule.Enable(ctx); err != nil {
		return err
	}

	log.Println("EnhancedLessonDialogsModule enabled")
	return nil
}

// Disable deactivates the module
func (m *EnhancedLessonDialogsModule) Disable(ctx context.Context) error {
	if err := m.BaseModule.Disable(ctx); err != nil {
		return err
	}

	log.Println("EnhancedLessonDialogsModule disabled")
	return nil
}

// SetManager sets the module manager
func (m *EnhancedLessonDialogsModule) SetManager(manager *core.Manager) {
	m.manager = manager
}

// ShowNewLessonDialog displays the new lesson creation dialog
func (m *EnhancedLessonDialogsModule) ShowNewLessonDialog() *LessonDialogResult {
	log.Printf("Showing new lesson dialog")

	// TODO: In full implementation, create and show Qt dialog
	// For now, return a stub result

	result := &LessonDialogResult{
		Success:      true,
		Title:        "New Lesson",
		Description:  "A new lesson created with enhanced dialog system",
		Author:       "User",
		Version:      "1.0",
		LessonType:   LessonTypeWords,
		QuestionLang: "en",
		AnswerLang:   "es",
	}

	if !m.validateNewLessonData(result) {
		result.Success = false
	}

	return result
}

// ShowEditPropertiesDialog displays the lesson properties editing dialog
func (m *EnhancedLessonDialogsModule) ShowEditPropertiesDialog(currentLesson *lesson.Lesson) *LessonDialogResult {
	log.Printf("Showing edit properties dialog for lesson type: %s", currentLesson.DataType)

	// TODO: In full implementation, create and show Qt dialog with current values
	// For now, return current properties with stub modifications

	result := &LessonDialogResult{
		Success:     true,
		Title:       currentLesson.Data.List.Title,
		Description: "Updated lesson description",
		Author:      "User",
		Version:     "1.1",
		LessonType:  m.stringToLessonType(currentLesson.DataType),
	}

	return result
}

// ShowImportDialog displays the lesson import dialog
func (m *EnhancedLessonDialogsModule) ShowImportDialog() *LessonDialogResult {
	log.Printf("Showing import dialog")

	// TODO: In full implementation, create and show Qt file dialog and import options

	result := &LessonDialogResult{
		Success:  true,
		FilePath: "/tmp/example.csv", // Stub file path
		ImportSettings: &ImportSettings{
			FilePath:       "/tmp/example.csv",
			FileType:       "csv",
			Encoding:       "UTF-8",
			Separator:      ",",
			FirstRowHeader: true,
			QuestionColumn: 0,
			AnswerColumn:   1,
			CommentColumn:  -1,
			SkipEmptyRows:  true,
		},
	}

	return result
}

// ShowExportDialog displays the lesson export dialog
func (m *EnhancedLessonDialogsModule) ShowExportDialog(currentLesson *lesson.Lesson) *LessonDialogResult {
	log.Printf("Showing export dialog for lesson: %s", currentLesson.Data.List.Title)

	// TODO: In full implementation, create and show Qt file dialog and export options

	result := &LessonDialogResult{
		Success:  true,
		FilePath: fmt.Sprintf("/tmp/%s_exported.csv", strings.ReplaceAll(currentLesson.Data.List.Title, " ", "_")),
		ExportSettings: &ExportSettings{
			FilePath:     fmt.Sprintf("/tmp/%s_exported.csv", strings.ReplaceAll(currentLesson.Data.List.Title, " ", "_")),
			FileType:     "csv",
			IncludeStats: true,
			IncludeTests: false,
			Format:       "standard",
			Compression:  false,
		},
	}

	return result
}

// ShowSaveAsDialog displays the save as dialog
func (m *EnhancedLessonDialogsModule) ShowSaveAsDialog(currentLesson *lesson.Lesson) *LessonDialogResult {
	log.Printf("Showing save as dialog for lesson: %s", currentLesson.Data.List.Title)

	// TODO: In full implementation, create and show Qt file dialog

	suggestedName := m.sanitizeFilename(currentLesson.Data.List.Title)
	if suggestedName == "" {
		suggestedName = "untitled_lesson"
	}

	result := &LessonDialogResult{
		Success:  true,
		FilePath: fmt.Sprintf("/tmp/%s.otwd", suggestedName),
		Title:    currentLesson.Data.List.Title,
	}

	return result
}

// OkToClose asks if it's okay to close a lesson (handles unsaved changes)
func (m *EnhancedLessonDialogsModule) OkToClose(hasUnsavedChanges bool, lessonTitle string) bool {
	if !hasUnsavedChanges {
		return true
	}

	log.Printf("Asking user about unsaved changes in lesson: %s", lessonTitle)

	// TODO: In full implementation, show Qt message box asking user
	// For now, always return true (assume user wants to continue)

	return true
}

// ShowAboutLessonDialog shows information about a lesson
func (m *EnhancedLessonDialogsModule) ShowAboutLessonDialog(currentLesson *lesson.Lesson) {
	stats := m.calculateLessonStatistics(currentLesson)

	log.Printf("Lesson Information:")
	log.Printf("  Title: %s", currentLesson.Data.List.Title)
	log.Printf("  Type: %s", currentLesson.DataType)
	log.Printf("  Items: %d", stats.ItemCount)
	log.Printf("  Tests: %d", stats.TestCount)
	if stats.LastModified != nil {
		log.Printf("  Last Modified: %s", stats.LastModified.Format("2006-01-02 15:04:05"))
	}

	// TODO: In full implementation, show Qt dialog with this information
}

// ShowErrorDialog displays an error dialog
func (m *EnhancedLessonDialogsModule) ShowErrorDialog(title, message string) {
	log.Printf("ERROR DIALOG - %s: %s", title, message)
	// TODO: In full implementation, show Qt error message box
}

// ShowWarningDialog displays a warning dialog
func (m *EnhancedLessonDialogsModule) ShowWarningDialog(title, message string) {
	log.Printf("WARNING DIALOG - %s: %s", title, message)
	// TODO: In full implementation, show Qt warning message box
}

// ShowInfoDialog displays an information dialog
func (m *EnhancedLessonDialogsModule) ShowInfoDialog(title, message string) {
	log.Printf("INFO DIALOG - %s: %s", title, message)
	// TODO: In full implementation, show Qt information message box
}

// GetSupportedImportFormats returns supported import file formats
func (m *EnhancedLessonDialogsModule) GetSupportedImportFormats() []string {
	return append([]string(nil), m.supportedFormats["import"]...)
}

// GetSupportedExportFormats returns supported export file formats
func (m *EnhancedLessonDialogsModule) GetSupportedExportFormats() []string {
	return append([]string(nil), m.supportedFormats["export"]...)
}

// GetRecentFiles returns recently used files
func (m *EnhancedLessonDialogsModule) GetRecentFiles() []string {
	return append([]string(nil), m.recentFiles...)
}

// AddToRecentFiles adds a file to the recent files list
func (m *EnhancedLessonDialogsModule) AddToRecentFiles(filePath string) {
	// Remove if already exists
	for i, path := range m.recentFiles {
		if path == filePath {
			m.recentFiles = append(m.recentFiles[:i], m.recentFiles[i+1:]...)
			break
		}
	}

	// Add to beginning
	m.recentFiles = append([]string{filePath}, m.recentFiles...)

	// Keep only last 10
	if len(m.recentFiles) > 10 {
		m.recentFiles = m.recentFiles[:10]
	}
}

// GetLanguages returns available languages
func (m *EnhancedLessonDialogsModule) GetLanguages() []Language {
	return append([]Language(nil), m.languages...)
}

// validateNewLessonData validates new lesson dialog data
func (m *EnhancedLessonDialogsModule) validateNewLessonData(result *LessonDialogResult) bool {
	if strings.TrimSpace(result.Title) == "" {
		result.ValidationError = "Lesson title cannot be empty"
		return false
	}

	if len(result.Title) > 100 {
		result.ValidationError = "Lesson title is too long (max 100 characters)"
		return false
	}

	if result.QuestionLang == result.AnswerLang {
		result.ValidationError = "Question and answer languages should be different"
		return false
	}

	return true
}

// stringToLessonType converts string to LessonType
func (m *EnhancedLessonDialogsModule) stringToLessonType(dataType string) LessonType {
	switch strings.ToLower(dataType) {
	case "words":
		return LessonTypeWords
	case "topo", "topography":
		return LessonTypeTopo
	case "media":
		return LessonTypeMedia
	default:
		return LessonTypeCustom
	}
}

// sanitizeFilename removes invalid characters from filename
func (m *EnhancedLessonDialogsModule) sanitizeFilename(name string) string {
	// Remove or replace invalid filename characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	return strings.TrimSpace(name)
}

// LessonStatistics represents lesson statistics
type LessonStatistics struct {
	ItemCount    int
	TestCount    int
	LastModified *time.Time
	Size         int64
}

// calculateLessonStatistics calculates various statistics for a lesson
func (m *EnhancedLessonDialogsModule) calculateLessonStatistics(lesson *lesson.Lesson) *LessonStatistics {
	stats := &LessonStatistics{
		ItemCount: len(lesson.Data.List.Items),
		TestCount: len(lesson.Data.List.Tests),
	}

	// Calculate size (approximate)
	for _, item := range lesson.Data.List.Items {
		for _, q := range item.Questions {
			stats.Size += int64(len(q))
		}
		for _, a := range item.Answers {
			stats.Size += int64(len(a))
		}
		stats.Size += int64(len(item.Comment))
	}

	return stats
}

// getDefaultLanguages returns a list of commonly supported languages
func getDefaultLanguages() []Language {
	return []Language{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
		{"it", "Italian"},
		{"pt", "Portuguese"},
		{"ru", "Russian"},
		{"ja", "Japanese"},
		{"ko", "Korean"},
		{"zh", "Chinese"},
		{"ar", "Arabic"},
		{"hi", "Hindi"},
		{"nl", "Dutch"},
		{"sv", "Swedish"},
		{"no", "Norwegian"},
		{"da", "Danish"},
		{"fi", "Finnish"},
		{"pl", "Polish"},
		{"cs", "Czech"},
		{"hu", "Hungarian"},
	}
}

// GetInfo returns module information
func (m *EnhancedLessonDialogsModule) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":        "enhancedLessonDialogs",
		"name":        "Enhanced Lesson Dialogs",
		"description": "Comprehensive lesson dialog system with improved validation and user experience",
		"features": []string{
			"New lesson creation with validation",
			"Lesson properties editing",
			"Import/Export with format options",
			"Recent files tracking",
			"Multi-language support",
			"Save/Save As functionality",
			"Unsaved changes handling",
			"Lesson statistics display",
		},
		"status":              "stub",
		"supported_languages": len(m.languages),
		"import_formats":      m.supportedFormats["import"],
		"export_formats":      m.supportedFormats["export"],
		"recent_files_count":  len(m.recentFiles),
	}
}

// IsFileFormatSupported checks if a file format is supported for import/export
func (m *EnhancedLessonDialogsModule) IsFileFormatSupported(filePath, operation string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	formats, exists := m.supportedFormats[operation]
	if !exists {
		return false
	}

	for _, format := range formats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetFileTypeDescription returns a user-friendly description of a file type
func (m *EnhancedLessonDialogsModule) GetFileTypeDescription(ext string) string {
	switch strings.ToLower(ext) {
	case ".otwd":
		return "OpenTeacher Word Lesson"
	case ".csv":
		return "Comma Separated Values"
	case ".txt":
		return "Plain Text"
	case ".tsv":
		return "Tab Separated Values"
	case ".json":
		return "JSON Format"
	case ".xml":
		return "XML Format"
	case ".kvtml":
		return "KDE Vocabulary Trainer"
	case ".html":
		return "HTML Document"
	case ".pdf":
		return "PDF Document"
	default:
		return "Unknown Format"
	}
}
