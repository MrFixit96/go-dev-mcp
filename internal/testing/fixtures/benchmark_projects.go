// filepath: c:\Users\James\Documents\go-dev-mcp\internal\testing\fixtures\benchmark_projects.go
package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BenchmarkProject represents a project fixture for benchmark testing
type BenchmarkProject struct {
	Path          string // Root path of the project
	Name          string // Project name
	CodeSize      int    // Total code size in bytes
	FileCount     int    // Number of files to create
	TempWorkspace string // Temp workspace for the project
}

// BenchmarkProjectFixture creates a new benchmark project with specified size
func BenchmarkProjectFixture(rootDir, name string, codeSize, fileCount int) *BenchmarkProject {
	projPath := filepath.Join(rootDir, fmt.Sprintf("benchmark-%s-project", name))
	return &BenchmarkProject{
		Path:      projPath,
		Name:      name,
		CodeSize:  codeSize,
		FileCount: fileCount,
	}
}

// Setup initializes the benchmark project
func (p *BenchmarkProject) Setup() error {
	// Create project directory
	if err := os.MkdirAll(p.Path, 0755); err != nil {
		return fmt.Errorf("failed to create benchmark project directory: %w", err)
	}

	// Create go.mod file
	goModContent := fmt.Sprintf("module example.com/benchmark-%s\n\ngo 1.18\n", p.Name)
	if err := os.WriteFile(filepath.Join(p.Path, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create main.go
	mainGoContent := generateMainFile(p.Name)
	if err := os.WriteFile(filepath.Join(p.Path, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Calculate size per additional file (excluding main.go)
	remainingFiles := p.FileCount - 1
	if remainingFiles <= 0 {
		return nil // No additional files needed
	}

	// Size per additional file
	mainSize := len(mainGoContent)
	remainingSize := p.CodeSize - mainSize
	sizePerFile := remainingSize / remainingFiles

	// Create additional files
	for i := 0; i < remainingFiles; i++ {
		fileName := fmt.Sprintf("file%d.go", i+1)
		fileContent := generateGoFile(p.Name, i+1, sizePerFile)
		if err := os.WriteFile(filepath.Join(p.Path, fileName), []byte(fileContent), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", fileName, err)
		}
	}

	return nil
}

// Clean removes the benchmark project directory
func (p *BenchmarkProject) Clean() error {
	return os.RemoveAll(p.Path)
}

// Helper function to generate a main file
func generateMainFile(projName string) string {
	return fmt.Sprintf(`package main

import (
	"fmt"
)

func main() {
	fmt.Println("Benchmark project: %s")
	
	for i := 1; i <= 10; i++ {
		fmt.Printf("Processing item %%d\n", i)
	}
	
	message := getMessage()
	fmt.Println(message)
}

func getMessage() string {
	return "Benchmark completed successfully"
}
`, projName)
}

// Helper function to generate a Go file with specified size
func generateGoFile(projName string, fileNum, targetSize int) string {
	// Create a standard header
	header := fmt.Sprintf(`package main

// File%d contains code for the %s benchmark project
`, fileNum, projName)

	// Create function signature
	funcSig := fmt.Sprintf(`
func ProcessData%d(data []string) []string {
	result := make([]string, 0, len(data))
`, fileNum)

	// Create function footer
	footer := `
	return result
}
`

	// Estimate statement size
	stmtTemplate := `
	// Process item at index %d
	if %d > 0 && len(data) > %d {
		item := data[%d]
		processed := fmt.Sprintf("Processed: %%s", item)
		result = append(result, processed)
	}
`
	stmtSize := len(fmt.Sprintf(stmtTemplate, 0, 0, 0, 0))

	// Calculate how many statements to generate
	baseSize := len(header) + len(funcSig) + len(footer)
	remainingSize := targetSize - baseSize
	if remainingSize <= 0 {
		// File is already large enough with just the boilerplate
		return header + funcSig + footer
	}

	numStatements := remainingSize / stmtSize

	// Build the function body
	builder := strings.Builder{}
	builder.WriteString(header)

	// Add import if needed
	if numStatements > 0 {
		builder.WriteString("\nimport \"fmt\"\n")
	}

	builder.WriteString(funcSig)

	// Add statements
	for i := 0; i < numStatements; i++ {
		builder.WriteString(fmt.Sprintf(stmtTemplate, i, i, i, i))
	}

	builder.WriteString(footer)

	// Add additional helper functions if needed to meet the size target
	currentSize := builder.Len()
	if currentSize < targetSize {
		extraFunc := fmt.Sprintf(`
// Helper%d is an additional function to meet size requirements
func Helper%d(value string) string {
	return fmt.Sprintf("Helper: %%s", value)
}
`, fileNum, fileNum)

		// Add as many helper functions as needed
		helpersNeeded := (targetSize-currentSize)/len(extraFunc) + 1

		for i := 0; i < helpersNeeded; i++ {
			builder.WriteString(fmt.Sprintf(`
// Helper%d_%d is an additional function to meet size requirements
func Helper%d_%d(value string) string {
	return fmt.Sprintf("Helper%d_%d: %%s", value)
}
`, fileNum, i, fileNum, i, fileNum, i))
		}
	}

	return builder.String()
}
