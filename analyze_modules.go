package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ModuleInfo holds information about a module's implementation status
type ModuleInfo struct {
	Name              string
	Path              string
	Type              string
	HasNewFunction    bool
	HasEnable         bool
	HasDisable        bool
	HasInterface      bool
	FunctionCount     int
	StructCount       int
	TODOCount         int
	CommentLines      int
	CodeLines         int
	HasTests          bool
	Dependencies      []string
	CompletionPercent int
	Priority          int
	Category          string
	Issues            []string
	Strengths         []string
}

// Priority categories
const (
	CRITICAL = 1 // Core system functionality
	HIGH     = 2 // Important user-facing features
	MEDIUM   = 3 // Supporting functionality
	LOW      = 4 // Optional/enhancement features
)

var categoryPriority = map[string]int{
	"core":           CRITICAL,
	"system":         CRITICAL,
	"logic":          CRITICAL,
	"interfaces":     HIGH,
	"gui":            HIGH,
	"qt":             HIGH,
	"data":           HIGH,
	"misc":           MEDIUM,
	"profileRunners": MEDIUM,
	"profilerunners": MEDIUM,
	"jseval":         LOW,
	"pyttsx":         LOW,
	"latexcodec":     LOW,
}

func main() {
	fmt.Println("🔍 Recuerdo Module Implementation Analyzer")
	fmt.Println("==========================================")

	modulesDir := "internal/modules"
	if len(os.Args) > 1 {
		modulesDir = os.Args[1]
	}

	modules, err := analyzeModules(modulesDir)
	if err != nil {
		fmt.Printf("Error analyzing modules: %v\n", err)
		os.Exit(1)
	}

	// Sort by priority and completion
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Priority != modules[j].Priority {
			return modules[i].Priority < modules[j].Priority
		}
		return modules[i].CompletionPercent < modules[j].CompletionPercent
	})

	generateReport(modules)
}

func analyzeModules(rootDir string) ([]*ModuleInfo, error) {
	var modules []*ModuleInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") && !strings.Contains(path, "_test.go") {
			module, err := analyzeGoFile(path)
			if err != nil {
				fmt.Printf("Warning: Failed to analyze %s: %v\n", path, err)
				return nil
			}
			if module != nil {
				modules = append(modules, module)
			}
		}
		return nil
	})

	return modules, err
}

func analyzeGoFile(filePath string) (*ModuleInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Check if this looks like a module file
	if !looksLikeModule(node) {
		return nil, nil // Not a module file
	}

	module := &ModuleInfo{
		Path:         filePath,
		Name:         extractModuleName(filePath),
		Category:     extractCategory(filePath),
		Dependencies: extractDependencies(node),
		Issues:       []string{},
		Strengths:    []string{},
	}

	// Read file content for additional analysis
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	contentStr := string(content)

	// Analyze AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			module.FunctionCount++
			if x.Name != nil {
				funcName := x.Name.Name
				if strings.HasPrefix(funcName, "New") && strings.HasSuffix(funcName, "Module") {
					module.HasNewFunction = true
					module.Strengths = append(module.Strengths, "Has proper constructor")
				}
				if funcName == "Enable" {
					module.HasEnable = true
					module.Strengths = append(module.Strengths, "Implements Enable")
				}
				if funcName == "Disable" {
					module.HasDisable = true
					module.Strengths = append(module.Strengths, "Implements Disable")
				}
			}
		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				module.StructCount++
				if strings.HasSuffix(x.Name.Name, "Module") {
					module.HasInterface = true
					module.Strengths = append(module.Strengths, "Has Module struct")
				}
			}
		}
		return true
	})

	// Analyze content
	lines := strings.Split(contentStr, "\n")
	module.CodeLines = len(lines)

	todoPattern := regexp.MustCompile(`(?i)todo|fixme|hack|xxx`)
	commentPattern := regexp.MustCompile(`^\s*//`)

	for _, line := range lines {
		if commentPattern.MatchString(line) {
			module.CommentLines++
		}
		if todoPattern.MatchString(line) {
			module.TODOCount++
		}
	}

	// Check for test file
	testFile := strings.Replace(filePath, ".go", "_test.go", 1)
	if _, err := os.Stat(testFile); err == nil {
		module.HasTests = true
		module.Strengths = append(module.Strengths, "Has tests")
	}

	// Analyze content for implementation quality
	analyzeImplementationQuality(module, contentStr)

	// Calculate completion percentage
	module.CompletionPercent = calculateCompletion(module)

	// Assign priority
	module.Priority = assignPriority(module)

	// Extract module type from Go code
	module.Type = extractModuleType(contentStr)

	return module, nil
}

func looksLikeModule(node *ast.File) bool {
	// Check if file contains module-related patterns
	hasModuleStruct := false
	hasNewFunction := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if strings.HasSuffix(x.Name.Name, "Module") {
				hasModuleStruct = true
			}
		case *ast.FuncDecl:
			if x.Name != nil && strings.HasPrefix(x.Name.Name, "New") && strings.Contains(x.Name.Name, "Module") {
				hasNewFunction = true
			}
		}
		return true
	})

	return hasModuleStruct || hasNewFunction
}

func extractModuleName(path string) string {
	dir := filepath.Dir(path)
	name := filepath.Base(dir)

	// If the filename is different from directory, use filename
	filename := filepath.Base(path)
	if !strings.HasPrefix(filename, name) && filename != "main.go" {
		name = strings.TrimSuffix(filename, ".go")
	}

	return name
}

func extractCategory(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "misc"
}

func extractDependencies(node *ast.File) []string {
	var deps []string
	for _, imp := range node.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(path, "recuerdo/internal") {
				deps = append(deps, path)
			}
		}
	}
	return deps
}

func extractModuleType(content string) string {
	typePattern := regexp.MustCompile(`type\s*=\s*"([^"]+)"`)
	matches := typePattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: look for common type assignments
	patterns := []string{
		`\.type\s*=\s*"([^"]+)"`,
		`Type:\s*"([^"]+)"`,
		`type.*"([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return "unknown"
}

func analyzeImplementationQuality(module *ModuleInfo, content string) {
	// Check for common implementation issues
	if strings.Contains(content, "panic(\"not implemented\")") {
		module.Issues = append(module.Issues, "Contains unimplemented panics")
	}

	if strings.Contains(content, "// TODO") && module.TODOCount > 5 {
		module.Issues = append(module.Issues, fmt.Sprintf("Many TODOs (%d)", module.TODOCount))
	}

	if module.FunctionCount == 0 {
		module.Issues = append(module.Issues, "No functions implemented")
	}

	if !module.HasNewFunction {
		module.Issues = append(module.Issues, "Missing constructor function")
	}

	if !module.HasEnable && !module.HasDisable {
		module.Issues = append(module.Issues, "Missing Enable/Disable methods")
	}

	// Check for positive indicators
	if strings.Contains(content, "context.Context") {
		module.Strengths = append(module.Strengths, "Uses context for lifecycle")
	}

	if strings.Contains(content, "logging.Logger") {
		module.Strengths = append(module.Strengths, "Has logging integration")
	}

	if strings.Contains(content, "BaseModule") {
		module.Strengths = append(module.Strengths, "Uses BaseModule pattern")
	}

	if module.CommentLines > module.CodeLines/4 {
		module.Strengths = append(module.Strengths, "Well documented")
	}

	// Check for Qt integration quality
	if strings.Contains(content, "widgets.") && strings.Contains(content, "ConnectClicked") {
		module.Strengths = append(module.Strengths, "Has interactive Qt widgets")
	}
}

func calculateCompletion(module *ModuleInfo) int {
	score := 0
	maxScore := 100

	// Basic structure (30 points)
	if module.HasNewFunction {
		score += 10
	}
	if module.HasInterface {
		score += 10
	}
	if module.HasEnable || module.HasDisable {
		score += 10
	}

	// Implementation depth (40 points)
	if module.FunctionCount > 0 {
		score += min(module.FunctionCount*3, 20) // Up to 20 points for functions
	}
	if len(module.Issues) == 0 {
		score += 10
	}
	if len(module.Strengths) > 2 {
		score += 10
	}

	// Code quality (20 points)
	if module.HasTests {
		score += 10
	}
	if module.TODOCount < 3 {
		score += 5
	}
	if module.CommentLines > 5 {
		score += 5
	}

	// Integration (10 points)
	if len(module.Dependencies) > 0 {
		score += 5
	}
	if module.Type != "unknown" {
		score += 5
	}

	return min(score, maxScore)
}

func assignPriority(module *ModuleInfo) int {
	basePriority, exists := categoryPriority[module.Category]
	if !exists {
		basePriority = MEDIUM
	}

	// Adjust based on current state
	if module.CompletionPercent < 30 && basePriority <= HIGH {
		// Critical/High priority modules that are barely implemented need attention
		return CRITICAL
	}

	if module.CompletionPercent > 80 {
		// Well-implemented modules can be lower priority
		return basePriority + 1
	}

	// Special cases
	if strings.Contains(strings.ToLower(module.Name), "gui") ||
		strings.Contains(strings.ToLower(module.Name), "dialog") {
		return HIGH // UI is important for user experience
	}

	if strings.Contains(strings.ToLower(module.Name), "system") ||
		strings.Contains(strings.ToLower(module.Name), "core") {
		return CRITICAL // System modules are critical
	}

	return basePriority
}

func generateReport(modules []*ModuleInfo) {
	fmt.Printf("\n📊 ANALYSIS RESULTS: %d modules analyzed\n\n", len(modules))

	// Summary statistics
	var totalCompletion, criticalCount, highCount, mediumCount, lowCount int
	var implementedCount, partialCount, stubCount int

	for _, module := range modules {
		totalCompletion += module.CompletionPercent

		switch module.Priority {
		case CRITICAL:
			criticalCount++
		case HIGH:
			highCount++
		case MEDIUM:
			mediumCount++
		case LOW:
			lowCount++
		}

		if module.CompletionPercent >= 70 {
			implementedCount++
		} else if module.CompletionPercent >= 30 {
			partialCount++
		} else {
			stubCount++
		}
	}

	avgCompletion := totalCompletion / len(modules)

	fmt.Printf("📈 SUMMARY STATISTICS\n")
	fmt.Printf("Average Completion: %d%%\n", avgCompletion)
	fmt.Printf("Implementation Status:\n")
	fmt.Printf("  ✅ Fully Implemented (70%+): %d modules\n", implementedCount)
	fmt.Printf("  🔄 Partially Implemented (30-70%%): %d modules\n", partialCount)
	fmt.Printf("  📝 Stub/Skeleton (<30%%): %d modules\n", stubCount)
	fmt.Printf("\nPriority Distribution:\n")
	fmt.Printf("  🚨 Critical: %d modules\n", criticalCount)
	fmt.Printf("  ⚡ High: %d modules\n", highCount)
	fmt.Printf("  📊 Medium: %d modules\n", mediumCount)
	fmt.Printf("  💡 Low: %d modules\n", lowCount)

	// Priority-based recommendations
	fmt.Printf("\n🎯 DEVELOPMENT RECOMMENDATIONS\n")
	fmt.Printf("===============================\n")

	currentPriority := 0
	for _, module := range modules {
		if module.Priority != currentPriority {
			currentPriority = module.Priority
			priorityName := []string{"", "🚨 CRITICAL", "⚡ HIGH", "📊 MEDIUM", "💡 LOW"}[currentPriority]
			fmt.Printf("\n%s PRIORITY:\n", priorityName)
			fmt.Println(strings.Repeat("─", 50))
		}

		statusIcon := getStatusIcon(module.CompletionPercent)
		fmt.Printf("%s %s (%s) - %d%% complete\n",
			statusIcon, module.Name, module.Category, module.CompletionPercent)
		fmt.Printf("    Type: %s | Functions: %d | TODOs: %d\n",
			module.Type, module.FunctionCount, module.TODOCount)

		if len(module.Issues) > 0 {
			fmt.Printf("    ❌ Issues: %s\n", strings.Join(module.Issues, ", "))
		}
		if len(module.Strengths) > 0 {
			fmt.Printf("    ✅ Strengths: %s\n", strings.Join(module.Strengths[:min(len(module.Strengths), 3)], ", "))
		}
		fmt.Println()
	}

	// Generate action items
	fmt.Printf("\n🚀 IMMEDIATE ACTION ITEMS\n")
	fmt.Printf("=========================\n")

	actionCount := 0
	for _, module := range modules {
		if actionCount >= 10 { // Limit to top 10 actions
			break
		}

		if module.Priority <= HIGH && module.CompletionPercent < 50 {
			actionCount++
			fmt.Printf("%d. Complete %s module (%s category)\n", actionCount, module.Name, module.Category)
			fmt.Printf("   Current: %d%% | Priority: %s | Path: %s\n",
				module.CompletionPercent, getPriorityName(module.Priority), module.Path)

			if len(module.Issues) > 0 {
				fmt.Printf("   Focus: %s\n", module.Issues[0])
			}
			fmt.Println()
		}
	}

	fmt.Printf("\n💡 Use this analysis to prioritize development efforts!\n")
	fmt.Printf("   Focus on Critical and High priority modules with low completion rates.\n")
}

func getStatusIcon(completion int) string {
	if completion >= 80 {
		return "✅"
	} else if completion >= 50 {
		return "🔄"
	} else if completion >= 20 {
		return "📝"
	} else {
		return "❌"
	}
}

func getPriorityName(priority int) string {
	names := []string{"", "Critical", "High", "Medium", "Low"}
	if priority < len(names) {
		return names[priority]
	}
	return "Unknown"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
