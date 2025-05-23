package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
)

// TestProject represents a Go project fixture for testing
type TestProject struct {
	Path          string
	Name          string
	Type          string
	Files         map[string]string
	IsSetup       bool
	TempWorkspace string // Temporary workspace directory for strategy testing
}

// SimpleProjectFixture returns a simple project fixture
func SimpleProjectFixture(basePath, name string) *TestProject {
	projectPath := filepath.Join(basePath, name)
	return &TestProject{
		Path: projectPath,
		Name: name,
		Type: "simple",
		Files: map[string]string{
			"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		},
	}
}

// MultiFileProjectFixture returns a multi-file project fixture
func MultiFileProjectFixture(basePath, name string) *TestProject {
	projectPath := filepath.Join(basePath, name)
	return &TestProject{
		Path: projectPath,
		Name: name,
		Type: "multi-file",
		Files: map[string]string{
			"main.go": `package main

import "fmt"

func main() {
	greeting := GetGreeting()
	name := GetName()
	fmt.Printf("%s, %s!\n", greeting, name)
}
`,
			"greeting.go": `package main

func GetGreeting() string {
	return "Hello"
}
`,
			"name.go": `package main

func GetName() string {
	return "World"
}
`,
		},
	}
}

// WithDepsProjectFixture returns a project fixture with external dependencies
func WithDepsProjectFixture(basePath, name string) *TestProject {
	projectPath := filepath.Join(basePath, name)
	return &TestProject{
		Path: projectPath,
		Name: name,
		Type: "with-deps",
		Files: map[string]string{
			"main.go": `package main

import (
	"fmt"
	"github.com/fatih/color"
)

func main() {
	c := color.New(color.FgCyan)
	c.Println("Hello, World!")
}
`,
		},
	}
}

// Setup creates the project files and initializes the Go module
func (p *TestProject) Setup() error {
	if p.IsSetup {
		return nil
	}

	// Create project directory
	if err := os.MkdirAll(p.Path, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project files
	for filename, content := range p.Files {
		filePath := filepath.Join(p.Path, filename)

		// Ensure parent directory exists
		if dir := filepath.Dir(filePath); dir != p.Path {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Write file content
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	p.IsSetup = true
	return nil
}

// Cleanup removes the project files
func (p *TestProject) Cleanup() error {
	if !p.IsSetup {
		return nil
	}

	if err := os.RemoveAll(p.Path); err != nil {
		return fmt.Errorf("failed to clean up project: %w", err)
	}

	p.IsSetup = false
	return nil
}

// ModifyFile modifies a file in the project
func (p *TestProject) ModifyFile(filename, newContent string) error {
	if !p.IsSetup {
		return fmt.Errorf("project not set up")
	}

	filePath := filepath.Join(p.Path, filename)
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to modify %s: %w", filename, err)
	}

	// Update the Files map
	p.Files[filename] = newContent
	return nil
}

// AddFile adds a new file to the project
func (p *TestProject) AddFile(filename, content string) error {
	if !p.IsSetup {
		return fmt.Errorf("project not set up")
	}

	filePath := filepath.Join(p.Path, filename)

	// Ensure parent directory exists
	if dir := filepath.Dir(filePath); dir != p.Path {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to add %s: %w", filename, err)
	}

	// Update the Files map
	p.Files[filename] = content
	return nil
}

// GetFilePath returns the absolute path to a file in the project
func (p *TestProject) GetFilePath(filename string) string {
	return filepath.Join(p.Path, filename)
}
