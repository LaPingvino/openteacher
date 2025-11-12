package lesson

import (
	"time"
)

// WordItem represents a single word/question-answer pair in a lesson
type WordItem struct {
	ID        int      `json:"id"`
	Questions []string `json:"questions"`
	Answers   []string `json:"answers"`
	Comment   string   `json:"comment,omitempty"`
}

// TestResult represents a single test result for an item
type TestResult struct {
	Result string     `json:"result"` // "right" or "wrong"
	ItemID int        `json:"itemId"`
	Time   *time.Time `json:"time,omitempty"`
}

// Test represents a collection of test results
type Test struct {
	Results []TestResult `json:"results"`
	Date    *time.Time   `json:"date,omitempty"`
}

// WordList represents the core lesson data structure
type WordList struct {
	Title            string     `json:"title,omitempty"`
	QuestionLanguage string     `json:"questionLanguage,omitempty"`
	AnswerLanguage   string     `json:"answerLanguage,omitempty"`
	Items            []WordItem `json:"items"`
	Tests            []Test     `json:"tests"`
}

// LessonData represents the complete lesson data as returned by loaders
type LessonData struct {
	List      WordList               `json:"list"`
	Resources map[string]interface{} `json:"resources"`
	Changed   bool                   `json:"changed,omitempty"`
}

// Lesson represents a lesson instance in the application
type Lesson struct {
	Data     LessonData
	Path     string
	DataType string // "words", "media", "topo", etc.
}

// NewWordList creates a new empty word list
func NewWordList() *WordList {
	return &WordList{
		Items: make([]WordItem, 0),
		Tests: make([]Test, 0),
	}
}

// NewLessonData creates a new lesson data structure
func NewLessonData() *LessonData {
	return &LessonData{
		List:      *NewWordList(),
		Resources: make(map[string]interface{}),
		Changed:   false,
	}
}

// NewLesson creates a new lesson instance
func NewLesson(dataType string) *Lesson {
	return &Lesson{
		Data:     *NewLessonData(),
		DataType: dataType,
	}
}

// AddWordItem adds a word item to the lesson
func (wl *WordList) AddWordItem(questions, answers []string, comment string) {
	item := WordItem{
		ID:        len(wl.Items),
		Questions: questions,
		Answers:   answers,
		Comment:   comment,
	}
	wl.Items = append(wl.Items, item)
}

// GetWordCount returns the number of word items in the lesson
func (wl *WordList) GetWordCount() int {
	return len(wl.Items)
}

// GetTestCount returns the number of tests in the lesson
func (wl *WordList) GetTestCount() int {
	return len(wl.Tests)
}

// AddTestResult adds a test result to the lesson
func (wl *WordList) AddTestResult(itemID int, result string) {
	testResult := TestResult{
		Result: result,
		ItemID: itemID,
		Time:   &time.Time{},
	}

	// Add to the most recent test, or create a new one
	if len(wl.Tests) == 0 {
		wl.Tests = append(wl.Tests, Test{
			Results: []TestResult{testResult},
			Date:    &time.Time{},
		})
	} else {
		lastTest := &wl.Tests[len(wl.Tests)-1]
		lastTest.Results = append(lastTest.Results, testResult)
	}
}

// GetRightAnswersCount returns the number of correct answers for an item
func (wl *WordList) GetRightAnswersCount(itemID int) int {
	count := 0
	for _, test := range wl.Tests {
		for _, result := range test.Results {
			if result.ItemID == itemID && result.Result == "right" {
				count++
			}
		}
	}
	return count
}

// GetWrongAnswersCount returns the number of incorrect answers for an item
func (wl *WordList) GetWrongAnswersCount(itemID int) int {
	count := 0
	for _, test := range wl.Tests {
		for _, result := range test.Results {
			if result.ItemID == itemID && result.Result == "wrong" {
				count++
			}
		}
	}
	return count
}
