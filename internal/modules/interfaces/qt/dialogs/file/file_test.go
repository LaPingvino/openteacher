package file

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFileDialogModule(t *testing.T) {
	module := NewFileDialogModule()

	if module == nil {
		t.Fatal("NewFileDialogModule returned nil")
	}

	if module.GetType() != "fileDialog" {
		t.Errorf("Expected module type 'fileDialog', got '%s'", module.GetType())
	}

	if module.GetName() != "file-dialog-module" {
		t.Errorf("Expected module name 'file-dialog-module', got '%s'", module.GetName())
	}
}

func TestGetSupportedFormats(t *testing.T) {
	module := NewFileDialogModule()
	formats := module.GetSupportedFormats()

	expectedFormats := map[string]string{
		"All Lesson Files":       "*.ot *.otwd *.csv *.tsv *.txt *.json *.kvtml *.anki *.anki2 *.apkg *.xml *.kgm *.ottp *.otmd",
		"OpenTeacher Files":      "*.ot *.otwd *.ottp *.otmd",
		"Text & CSV Files":       "*.txt *.csv *.tsv *.json",
		"Anki Files":             "*.anki *.anki2 *.apkg",
		"KDE/Educational Files":  "*.kvtml *.kgm",
		"Vocabulary Trainers":    "*.voc *.fq *.fmd *.dkf *.jml *.jvlt *.stp *.db",
		"Language Learning Apps": "*.oh *.ohw *.oh4 *.ovr *.pau *.t2k *.vok2 *.wdl *.vtl3 *.wrts",
		"Other Formats":          "*.backpack *.wcu *.xml",
		"All Files":              "*.*",
	}

	if !reflect.DeepEqual(formats, expectedFormats) {
		t.Errorf("GetSupportedFormats() returned unexpected formats")
		for key, expected := range expectedFormats {
			if actual, exists := formats[key]; !exists {
				t.Errorf("Missing format category: %s", key)
			} else if actual != expected {
				t.Errorf("Format category %s: expected %s, got %s", key, expected, actual)
			}
		}
		for key := range formats {
			if _, exists := expectedFormats[key]; !exists {
				t.Errorf("Unexpected format category: %s", key)
			}
		}
	}
}

func TestBuildFilterString(t *testing.T) {
	module := NewFileDialogModule()

	testCases := []struct {
		name     string
		formats  map[string]string
		expected string
	}{
		{
			name:     "Empty formats",
			formats:  map[string]string{},
			expected: module.fileFilter, // Should return default filter
		},
		{
			name: "Single format",
			formats: map[string]string{
				"Text Files": "*.txt",
			},
			expected: "Text Files (*.txt)",
		},
		{
			name: "Multiple formats",
			formats: map[string]string{
				"Text Files": "*.txt",
				"CSV Files":  "*.csv",
			},
			expected: "CSV Files (*.csv);;Text Files (*.txt)", // Note: map iteration order may vary
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := module.BuildFilterString(tc.formats)

			if tc.name == "Empty formats" {
				if result != tc.expected {
					t.Errorf("BuildFilterString with empty formats: expected default filter, got %s", result)
				}
			} else if tc.name == "Single format" {
				if result != tc.expected {
					t.Errorf("BuildFilterString: expected %s, got %s", tc.expected, result)
				}
			} else if tc.name == "Multiple formats" {
				// For multiple formats, just check that both are present
				if !strings.Contains(result, "Text Files (*.txt)") || !strings.Contains(result, "CSV Files (*.csv)") {
					t.Errorf("BuildFilterString: expected both formats to be present in %s", result)
				}
			}
		})
	}
}

func TestDefaultFilter(t *testing.T) {
	module := NewFileDialogModule()

	// Verify that the default filter contains all major file types
	expectedExtensions := []string{
		"*.ot", "*.otwd", "*.csv", "*.tsv", "*.txt", "*.json", "*.kvtml",
		"*.anki", "*.anki2", "*.apkg", "*.xml", "*.kgm", "*.ottp", "*.otmd",
		"*.voc", "*.fq", "*.fmd", "*.dkf", "*.jml", "*.jvlt", "*.stp", "*.db",
		"*.oh", "*.ohw", "*.oh4", "*.ovr", "*.pau", "*.t2k", "*.vok2",
		"*.wdl", "*.vtl3", "*.wrts", "*.backpack", "*.wcu",
	}

	for _, ext := range expectedExtensions {
		if !strings.Contains(module.fileFilter, ext) {
			t.Errorf("Default filter missing extension: %s", ext)
		}
	}

	// Verify filter format is correct (contains ;; separators)
	if !strings.Contains(module.fileFilter, ";;") {
		t.Error("Default filter should contain ;; separators for multiple filters")
	}

	// Verify "All Files" is present
	if !strings.Contains(module.fileFilter, "All Files (*.*)") {
		t.Error("Default filter should contain 'All Files (*.*)'")
	}
}

func TestSetGetDefaultDirectory(t *testing.T) {
	module := NewFileDialogModule()

	testDir := "/tmp/test"
	module.SetDefaultDirectory(testDir)

	result := module.GetDefaultDirectory()
	if result != testDir {
		t.Errorf("SetDefaultDirectory/GetDefaultDirectory: expected %s, got %s", testDir, result)
	}
}

func TestSetDefaultFilter(t *testing.T) {
	module := NewFileDialogModule()

	originalFilter := module.fileFilter
	testFilter := "Test Files (*.test)"

	module.SetDefaultFilter(testFilter)

	// The filter should be updated
	if module.fileFilter != testFilter {
		t.Errorf("SetDefaultFilter: expected filter to be updated to %s, got %s", testFilter, module.fileFilter)
	}

	// Restore original filter
	module.SetDefaultFilter(originalFilter)
}

func TestFileFilterCoverage(t *testing.T) {
	module := NewFileDialogModule()

	// Test that major file format categories are covered
	expectedCategories := []string{
		"All Lesson Files",
		"OpenTeacher Files",
		"Text & CSV Files",
		"Anki Files",
		"KDE/Educational Files",
		"Vocabulary Trainers",
		"Language Learning Apps",
		"Other Formats",
		"All Files",
	}

	formats := module.GetSupportedFormats()

	for _, category := range expectedCategories {
		if _, exists := formats[category]; !exists {
			t.Errorf("Missing expected format category: %s", category)
		}
	}

	// Test that each category has meaningful content
	for category, pattern := range formats {
		if category == "All Files" {
			if pattern != "*.*" {
				t.Errorf("Category '%s' should be '*.*', got '%s'", category, pattern)
			}
		} else {
			if !strings.Contains(pattern, "*") {
				t.Errorf("Category '%s' should contain file patterns, got '%s'", category, pattern)
			}
		}
	}
}

func TestComprehensiveFormatSupport(t *testing.T) {
	module := NewFileDialogModule()

	// List of all file extensions that should be supported based on legacy OpenTeacher
	allExpectedExtensions := []string{
		// Basic formats
		"*.ot", "*.otwd", "*.ottp", "*.otmd",
		"*.csv", "*.tsv", "*.txt", "*.json", "*.xml",

		// Anki formats
		"*.anki", "*.anki2", "*.apkg",

		// KDE/Educational formats
		"*.kvtml", "*.kgm",

		// Vocabulary trainer formats
		"*.voc", "*.fq", "*.fmd", "*.dkf", "*.jml", "*.jvlt", "*.stp", "*.db",

		// Language learning app formats
		"*.oh", "*.ohw", "*.oh4", "*.ovr", "*.pau", "*.t2k", "*.vok2",
		"*.wdl", "*.vtl3", "*.wrts",

		// Other formats
		"*.backpack", "*.wcu",
	}

	formats := module.GetSupportedFormats()

	// Build a single string containing all patterns
	allPatterns := ""
	for _, pattern := range formats {
		allPatterns += " " + pattern
	}

	// Check that each expected extension appears somewhere in the formats
	for _, ext := range allExpectedExtensions {
		if !strings.Contains(allPatterns, ext) {
			t.Errorf("Expected extension %s not found in any format category", ext)
		}
	}

	// Check specific category contents
	testCases := []struct {
		category    string
		mustContain []string
	}{
		{
			category:    "All Lesson Files",
			mustContain: []string{"*.ot", "*.csv", "*.kvtml", "*.anki"},
		},
		{
			category:    "Anki Files",
			mustContain: []string{"*.anki", "*.anki2", "*.apkg"},
		},
		{
			category:    "KDE/Educational Files",
			mustContain: []string{"*.kvtml", "*.kgm"},
		},
		{
			category:    "Vocabulary Trainers",
			mustContain: []string{"*.voc", "*.jml", "*.db"},
		},
	}

	for _, tc := range testCases {
		pattern, exists := formats[tc.category]
		if !exists {
			t.Errorf("Category %s not found", tc.category)
			continue
		}

		for _, ext := range tc.mustContain {
			if !strings.Contains(pattern, ext) {
				t.Errorf("Category %s missing expected extension %s", tc.category, ext)
			}
		}
	}
}

func TestModuleRequirements(t *testing.T) {
	module := NewFileDialogModule()

	// Check that the module requires qtApp
	requirements := module.GetRequires()
	found := false
	for _, req := range requirements {
		if req == "qtApp" {
			found = true
			break
		}
	}

	if !found {
		t.Error("FileDialogModule should require 'qtApp'")
	}
}

// Integration test to verify the filter string works as expected
func TestFilterStringIntegration(t *testing.T) {
	module := NewFileDialogModule()

	// The default filter should be properly formatted for Qt
	filter := module.fileFilter

	// Should start with a filter name
	if !strings.Contains(filter, "All Lesson Files") {
		t.Error("Filter should start with 'All Lesson Files' category")
	}

	// Should contain proper Qt filter format: "Name (*.ext)"
	if !strings.Contains(filter, "(*.") {
		t.Error("Filter should contain Qt format '(*.ext)'")
	}

	// Should have multiple categories separated by ;;
	parts := strings.Split(filter, ";;")
	if len(parts) < 5 {
		t.Errorf("Filter should have multiple categories separated by ;;, got %d parts", len(parts))
	}

	// Each part should be properly formatted
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, "(") || !strings.Contains(part, ")") {
			t.Errorf("Filter part %d should contain parentheses: %s", i, part)
		}
	}
}
