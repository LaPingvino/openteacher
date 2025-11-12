package lesson

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileLoader provides file loading functionality for various lesson formats
type FileLoader struct{}

// NewFileLoader creates a new file loader instance
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// LoadFile loads a lesson file and returns lesson data
func (fl *FileLoader) LoadFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.LoadFile() - loading file: %s", filePath)

	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv", ".tsv":
		return fl.loadCSV(filePath)
	case ".txt":
		return fl.loadTextFile(filePath)
	case ".ot":
		return fl.loadOpenTeacherFile(filePath)
	case ".json":
		return fl.loadJSONFile(filePath)
	default:
		// Try to auto-detect format by content
		return fl.loadAutoDetect(filePath)
	}
}

// GetFileType determines the lesson type for a given file path
func (fl *FileLoader) GetFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv", ".tsv", ".txt", ".ot", ".json":
		return "words"
	case ".kgm":
		return "topo"
	default:
		return "words" // Default to words for unknown formats
	}
}

// loadCSV loads CSV or TSV files
func (fl *FileLoader) loadCSV(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadCSV() - parsing CSV file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open CSV file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Determine delimiter
	delimiter := ','
	if strings.HasSuffix(strings.ToLower(filePath), ".tsv") {
		delimiter = '\t'
	}

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	itemID := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[WARNING] Error reading CSV line: %v", err)
			continue
		}

		if len(record) < 2 {
			continue // Skip lines with insufficient data
		}

		// Parse questions and answers (may be comma-separated within cells)
		questions := fl.parseWordString(strings.TrimSpace(record[0]))
		answers := fl.parseWordString(strings.TrimSpace(record[1]))

		comment := ""
		if len(record) > 2 {
			comment = strings.TrimSpace(record[2])
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        itemID,
				Questions: questions,
				Answers:   answers,
				Comment:   comment,
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadCSV() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadTextFile loads simple text files with word pairs
func (fl *FileLoader) loadTextFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadTextFile() - parsing text file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open text file: %v", err)
		return nil, err
	}
	defer file.Close()

	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	itemID := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Try different separators
		var questions, answers []string
		var comment string

		if strings.Contains(line, "\t") {
			// Tab-separated
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(parts[1])
				if len(parts) > 2 {
					comment = strings.Join(parts[2:], " ")
				}
			}
		} else if strings.Contains(line, "|") {
			// Pipe-separated
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(parts[1])
				if len(parts) > 2 {
					comment = strings.Join(parts[2:], " ")
				}
			}
		} else if strings.Contains(line, "=") {
			// Equals-separated
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				questions = fl.parseWordString(parts[0])
				answers = fl.parseWordString(strings.Join(parts[1:], "="))
			}
		} else {
			// Try to find other patterns or skip
			continue
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        itemID,
				Questions: questions,
				Answers:   answers,
				Comment:   comment,
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
			itemID++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Error reading text file: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] FileLoader.loadTextFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadOpenTeacherFile loads OpenTeacher .ot XML files
func (fl *FileLoader) loadOpenTeacherFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadOpenTeacherFile() - parsing .ot XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open .ot file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Simple XML structure for .ot files
	type OTWord struct {
		Known   string `xml:"known"`
		Foreign string `xml:"foreign"`
		Second  string `xml:"second"`
		Results string `xml:"results"`
	}

	type OTRoot struct {
		XMLName          xml.Name `xml:"openteaching-words-file"`
		Title            string   `xml:"title"`
		QuestionLanguage string   `xml:"question_language"`
		AnswerLanguage   string   `xml:"answer_language"`
		Words            []OTWord `xml:"word"`
	}

	var root OTRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse .ot XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Title
	lessonData.List.QuestionLanguage = root.QuestionLanguage
	lessonData.List.AnswerLanguage = root.AnswerLanguage

	// Create a test for storing results
	test := Test{Results: []TestResult{}}

	for i, word := range root.Words {
		questions := fl.parseWordString(word.Known)

		// Handle foreign and second fields
		foreignText := word.Foreign
		if word.Second != "" {
			foreignText += ", " + word.Second
		}
		answers := fl.parseWordString(foreignText)

		item := WordItem{
			ID:        i,
			Questions: questions,
			Answers:   answers,
			Comment:   "",
		}
		lessonData.List.Items = append(lessonData.List.Items, item)

		// Parse results (format: "wrong/total")
		if word.Results != "" {
			parts := strings.Split(word.Results, "/")
			if len(parts) == 2 {
				if wrong, err := strconv.Atoi(parts[0]); err == nil {
					if total, err := strconv.Atoi(parts[1]); err == nil {
						right := total - wrong

						// Add right results
						for j := 0; j < right; j++ {
							test.Results = append(test.Results, TestResult{
								Result: "right",
								ItemID: i,
							})
						}

						// Add wrong results
						for j := 0; j < wrong; j++ {
							test.Results = append(test.Results, TestResult{
								Result: "wrong",
								ItemID: i,
							})
						}
					}
				}
			}
		}
	}

	// Only add test if it has results
	if len(test.Results) > 0 {
		lessonData.List.Tests = append(lessonData.List.Tests, test)
	}

	log.Printf("[SUCCESS] FileLoader.loadOpenTeacherFile() - loaded %d word pairs with %d test results",
		len(lessonData.List.Items), len(test.Results))
	return lessonData, nil
}

// loadJSONFile loads JSON-formatted lesson files
func (fl *FileLoader) loadJSONFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadJSONFile() - parsing JSON file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open JSON file: %v", err)
		return nil, err
	}
	defer file.Close()

	var lessonData LessonData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&lessonData); err != nil {
		log.Printf("[ERROR] Failed to parse JSON: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] FileLoader.loadJSONFile() - loaded %d word pairs", len(lessonData.List.Items))
	return &lessonData, nil
}

// loadAutoDetect attempts to auto-detect file format and load accordingly
func (fl *FileLoader) loadAutoDetect(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadAutoDetect() - attempting to auto-detect format")

	// Try different loaders in order of likelihood
	loaders := []func(string) (*LessonData, error){
		fl.loadTextFile,
		fl.loadCSV,
		fl.loadJSONFile,
	}

	var lastErr error
	for _, loader := range loaders {
		if data, err := loader(filePath); err == nil {
			return data, nil
		} else {
			lastErr = err
		}
	}

	log.Printf("[ERROR] FileLoader.loadAutoDetect() - all format detection failed")
	return nil, fmt.Errorf("unable to detect file format: %v", lastErr)
}

// parseWordString parses a string containing potentially multiple words/phrases
// This mimics the Python wordsStringParser functionality
func (fl *FileLoader) parseWordString(input string) []string {
	if input == "" {
		return []string{}
	}

	input = strings.TrimSpace(input)

	// Handle multiple separators
	var words []string

	// First try semicolon separation (most explicit)
	if strings.Contains(input, ";") {
		parts := strings.Split(input, ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				words = append(words, trimmed)
			}
		}
	} else if strings.Contains(input, ",") {
		// Try comma separation
		parts := strings.Split(input, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				words = append(words, trimmed)
			}
		}
	} else {
		// Single word/phrase
		words = []string{input}
	}

	return words
}

// GetSupportedExtensions returns a list of supported file extensions
func (fl *FileLoader) GetSupportedExtensions() []string {
	return []string{".csv", ".tsv", ".txt", ".ot", ".json"}
}

// GetFormatName returns a human-readable name for a file extension
func (fl *FileLoader) GetFormatName(ext string) string {
	switch strings.ToLower(ext) {
	case ".csv":
		return "Spreadsheet (CSV)"
	case ".tsv":
		return "Tab-Separated Values"
	case ".txt":
		return "Text File"
	case ".ot":
		return "OpenTeacher 2.x"
	case ".json":
		return "JSON Lesson File"
	default:
		return "Unknown Format"
	}
}
