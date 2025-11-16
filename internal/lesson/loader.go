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
	case ".ot", ".otwd":
		return fl.loadOpenTeacherFile(filePath)
	case ".json":
		return fl.loadJSONFile(filePath)
	case ".kvtml":
		return fl.loadKVTMLFile(filePath)
	case ".anki":
		return fl.loadAnkiFile(filePath)
	case ".xml":
		return fl.loadXMLFile(filePath)
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
	case ".anki", ".anki2", ".apkg", ".backpack", ".wcu", ".voc", ".fq", ".fmd":
		return "words"
	case ".dkf", ".jml", ".jvlt", ".kvtml", ".stp", ".db", ".oh", ".ohw", ".oh4":
		return "words"
	case ".ovr", ".pau", ".vok2", ".wdl", ".vtl3", ".wrts", ".xml", ".otwd":
		return "words"
	case ".kgm", ".ottp":
		return "topo"
	case ".t2k":
		return "words" // Can be both words and topo, defaulting to words
	case ".otmd":
		return "media"
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

// loadKVTMLFile loads KVTML (KDE Vocabulary Document) files
func (fl *FileLoader) loadKVTMLFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadKVTMLFile() - parsing KVTML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open KVTML file: %v", err)
		return nil, err
	}
	defer file.Close()

	// KVTML XML structure
	type KVTMLTranslation struct {
		ID   string `xml:"id,attr"`
		Text string `xml:"text"`
	}

	type KVTMLEntry struct {
		ID           string             `xml:"id,attr"`
		Translations []KVTMLTranslation `xml:"translation"`
	}

	type KVTMLIdentifier struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name"`
	}

	type KVTMLInformation struct {
		Title   string `xml:"title"`
		Author  string `xml:"author"`
		Comment string `xml:"comment"`
	}

	type KVTMLRoot struct {
		XMLName     xml.Name          `xml:"kvtml"`
		Version     string            `xml:"version,attr"`
		Information KVTMLInformation  `xml:"information"`
		Identifiers []KVTMLIdentifier `xml:"identifiers>identifier"`
		Entries     []KVTMLEntry      `xml:"entries>entry"`
	}

	var root KVTMLRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		log.Printf("[ERROR] Failed to parse KVTML XML: %v", err)
		return nil, err
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Information.Title
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	// Set language names if available
	if len(root.Identifiers) >= 2 {
		lessonData.List.QuestionLanguage = root.Identifiers[0].Name
		lessonData.List.AnswerLanguage = root.Identifiers[1].Name
	}

	// Process entries
	for i, entry := range root.Entries {
		var questions, answers []string

		// Find question and answer translations
		for _, translation := range entry.Translations {
			if translation.ID == "0" && translation.Text != "" {
				questions = fl.parseWordString(translation.Text)
			} else if translation.ID == "1" && translation.Text != "" {
				answers = fl.parseWordString(translation.Text)
			}
		}

		if len(questions) > 0 && len(answers) > 0 {
			item := WordItem{
				ID:        i,
				Questions: questions,
				Answers:   answers,
				Comment:   "",
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadKVTMLFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// loadAnkiFile loads Anki database files
func (fl *FileLoader) loadAnkiFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadAnkiFile() - parsing Anki file")

	// Placeholder implementation - would need SQLite support for full functionality
	lessonData := NewLessonData()
	lessonData.List.Title = filepath.Base(filePath)

	log.Printf("[WARNING] FileLoader.loadAnkiFile() - Anki format not fully implemented yet")
	return lessonData, nil
}

// loadXMLFile loads XML files (including ABBYY format)
func (fl *FileLoader) loadXMLFile(filePath string) (*LessonData, error) {
	log.Printf("[ACTION] FileLoader.loadXMLFile() - parsing XML file")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open XML file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Basic XML structure for simple word lists
	type XMLWord struct {
		XMLName xml.Name `xml:"word"`
		Known   string   `xml:"known"`
		Foreign string   `xml:"foreign"`
		Text    string   `xml:",chardata"`
	}

	type XMLRoot struct {
		XMLName xml.Name  `xml:"root"`
		Title   string    `xml:"title"`
		Words   []XMLWord `xml:"word"`
	}

	// Try to parse as simple XML word list
	var root XMLRoot
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		// If XML parsing fails, try as text file
		log.Printf("[WARNING] XML parsing failed, trying as text file: %v", err)
		return fl.loadTextFile(filePath)
	}

	lessonData := NewLessonData()
	lessonData.List.Title = root.Title
	if lessonData.List.Title == "" {
		lessonData.List.Title = filepath.Base(filePath)
	}

	for i, word := range root.Words {
		var questions, answers []string

		if word.Known != "" {
			questions = fl.parseWordString(word.Known)
		}
		if word.Foreign != "" {
			answers = fl.parseWordString(word.Foreign)
		}
		if word.Text != "" && len(questions) == 0 {
			questions = fl.parseWordString(word.Text)
		}

		if len(questions) > 0 || len(answers) > 0 {
			item := WordItem{
				ID:        i,
				Questions: questions,
				Answers:   answers,
				Comment:   "",
			}
			lessonData.List.Items = append(lessonData.List.Items, item)
		}
	}

	log.Printf("[SUCCESS] FileLoader.loadXMLFile() - loaded %d word pairs", len(lessonData.List.Items))
	return lessonData, nil
}

// GetSupportedExtensions returns a list of supported file extensions
func (fl *FileLoader) GetSupportedExtensions() []string {
	return []string{
		".csv", ".tsv", ".txt", ".ot", ".json", ".kvtml", ".anki", ".anki2",
		".apkg", ".backpack", ".wcu", ".voc", ".fq", ".fmd", ".dkf", ".jml",
		".jvlt", ".stp", ".db", ".oh", ".ohw", ".oh4", ".ovr", ".pau",
		".t2k", ".vok2", ".wdl", ".vtl3", ".wrts", ".xml", ".kgm", ".ottp",
		".otmd", ".otwd",
	}
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
	case ".kvtml":
		return "KDE Vocabulary Document"
	case ".anki":
		return "Anki Database"
	case ".anki2":
		return "Anki 2.0 Database"
	case ".apkg":
		return "Anki Package"
	case ".backpack":
		return "Backpack File"
	case ".wcu":
		return "CueCard File"
	case ".voc":
		return "Vocabulary File"
	case ".fq":
		return "FlashQard File"
	case ".fmd":
		return "FM Dictionary"
	case ".dkf":
		return "Granule Deck"
	case ".jml":
		return "JMemorize Lesson"
	case ".jvlt":
		return "JVLT File"
	case ".stp":
		return "Ludem File"
	case ".db":
		return "Database File"
	case ".oh", ".ohw", ".oh4":
		return "Overhoor File"
	case ".ovr":
		return "Overhoringsprogramma Talen"
	case ".pau":
		return "Pauker File"
	case ".t2k":
		return "Teach2000 File"
	case ".vok2":
		return "Teachmaster File"
	case ".wdl":
		return "Oriente Voca File"
	case ".vtl3":
		return "VokabelTrainer File"
	case ".wrts":
		return "WRTS File"
	case ".xml":
		return "XML File"
	case ".kgm":
		return "KGeography Map"
	case ".ottp":
		return "OpenTeaching Topography"
	case ".otmd":
		return "OpenTeaching Media"
	case ".otwd":
		return "OpenTeaching Words"
	default:
		return "Unknown Format"
	}
}
