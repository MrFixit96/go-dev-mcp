package tools

import (
	"reflect"
	"testing"
)

func TestPrepareWorkspaceArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    InputContext
		module   string
		baseArgs []string
		want     []string
	}{
		{
			name: "SourceWorkspace with specific module",
			input: InputContext{
				Source:        SourceWorkspace,
				WorkspacePath: "/path/to/workspace",
			},
			module:   "./module1",
			baseArgs: []string{"go", "build"},
			want:     []string{"go", "build", "./module1"},
		},
		{
			name: "SourceWorkspace with empty module (all modules)",
			input: InputContext{
				Source:        SourceWorkspace,
				WorkspacePath: "/path/to/workspace",
			},
			module:   "",
			baseArgs: []string{"go", "build"},
			want:     []string{"go", "build", "./..."},
		},
		{
			name: "SourceProjectPath",
			input: InputContext{
				Source:      SourceProjectPath,
				ProjectPath: "/path/to/project",
			},
			module:   "",
			baseArgs: []string{"go", "test"},
			want:     []string{"go", "test", "./..."},
		},
		{
			name: "SourceProjectPath with module parameter (should be ignored)",
			input: InputContext{
				Source:      SourceProjectPath,
				ProjectPath: "/path/to/project",
			},
			module:   "./module1",
			baseArgs: []string{"go", "test"},
			want:     []string{"go", "test", "./..."},
		},
		{
			name: "SourceHybrid",
			input: InputContext{
				Source:      SourceHybrid,
				Code:        "package main",
				ProjectPath: "/path/to/project",
			},
			module:   "",
			baseArgs: []string{"go", "fmt"},
			want:     []string{"go", "fmt", "./..."},
		},
		{
			name: "SourceHybrid with module parameter (should be ignored)",
			input: InputContext{
				Source:      SourceHybrid,
				Code:        "package main",
				ProjectPath: "/path/to/project",
			},
			module:   "./module1",
			baseArgs: []string{"go", "fmt"},
			want:     []string{"go", "fmt", "./..."},
		},
		{
			name: "SourceCode",
			input: InputContext{
				Source: SourceCode,
				Code:   "package main\n\nfunc main() {}",
			},
			module:   "",
			baseArgs: []string{"go", "run"},
			want:     []string{"go", "run"},
		},
		{
			name: "SourceCode with module parameter (should be ignored)",
			input: InputContext{
				Source: SourceCode,
				Code:   "package main\n\nfunc main() {}",
			},
			module:   "./module1",
			baseArgs: []string{"go", "run"},
			want:     []string{"go", "run"},
		},
		{
			name: "Empty baseArgs with SourceWorkspace and module",
			input: InputContext{
				Source:        SourceWorkspace,
				WorkspacePath: "/path/to/workspace",
			},
			module:   "./module1",
			baseArgs: []string{},
			want:     []string{"./module1"},
		},
		{
			name: "Empty baseArgs with SourceProjectPath",
			input: InputContext{
				Source:      SourceProjectPath,
				ProjectPath: "/path/to/project",
			},
			module:   "",
			baseArgs: []string{},
			want:     []string{"./..."},
		},
		{
			name: "Complex baseArgs with flags",
			input: InputContext{
				Source:        SourceWorkspace,
				WorkspacePath: "/path/to/workspace",
			},
			module:   "./module1",
			baseArgs: []string{"go", "build", "-v", "-tags", "integration"},
			want:     []string{"go", "build", "-v", "-tags", "integration", "./module1"},
		},
		{
			name: "SourceWorkspace with workspace root module",
			input: InputContext{
				Source:        SourceWorkspace,
				WorkspacePath: "/path/to/workspace",
			},
			module:   "./",
			baseArgs: []string{"go", "test"},
			want:     []string{"go", "test", "./"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareWorkspaceArgs(tt.input, tt.module, tt.baseArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareWorkspaceArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrepareWorkspaceArgs_DefensiveCopy(t *testing.T) {
	// Test that baseArgs are not modified (defensive copy)
	originalBaseArgs := []string{"go", "build", "-v"}
	baseArgs := make([]string, len(originalBaseArgs))
	copy(baseArgs, originalBaseArgs)

	input := InputContext{
		Source:        SourceWorkspace,
		WorkspacePath: "/path/to/workspace",
	}

	result := prepareWorkspaceArgs(input, "./module1", baseArgs)

	// Verify baseArgs wasn't modified
	if !reflect.DeepEqual(baseArgs, originalBaseArgs) {
		t.Errorf("baseArgs was modified: got %v, want %v", baseArgs, originalBaseArgs)
	}

	// Verify result is correct
	expected := []string{"go", "build", "-v", "./module1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("prepareWorkspaceArgs() = %v, want %v", result, expected)
	}

	// Verify result is a different slice (not same reference)
	if &baseArgs[0] == &result[0] {
		t.Error("result should be a new slice, not a reference to baseArgs")
	}
}

func TestPrepareWorkspaceArgs_MultipleCallsDoNotInterfere(t *testing.T) {
	// Test that multiple calls with the same baseArgs don't interfere with each other
	baseArgs := []string{"go", "test"}

	input1 := InputContext{Source: SourceWorkspace}
	input2 := InputContext{Source: SourceProjectPath}
	input3 := InputContext{Source: SourceCode}

	result1 := prepareWorkspaceArgs(input1, "./module1", baseArgs)
	result2 := prepareWorkspaceArgs(input2, "", baseArgs)
	result3 := prepareWorkspaceArgs(input3, "", baseArgs)

	// Each result should be independent
	expected1 := []string{"go", "test", "./module1"}
	expected2 := []string{"go", "test", "./..."}
	expected3 := []string{"go", "test"}

	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("First call: got %v, want %v", result1, expected1)
	}
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Second call: got %v, want %v", result2, expected2)
	}
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Third call: got %v, want %v", result3, expected3)
	}

	// baseArgs should remain unchanged
	if !reflect.DeepEqual(baseArgs, []string{"go", "test"}) {
		t.Errorf("baseArgs was modified: %v", baseArgs)
	}
}

func TestPrepareWorkspaceArgs_AllSourceTypes(t *testing.T) {
	// Comprehensive test covering all source types in a single test
	baseArgs := []string{"go", "build"}
	module := "./mymodule"

	testCases := []struct {
		source   InputSource
		expected []string
	}{
		{SourceUnknown, []string{"go", "build"}},
		{SourceCode, []string{"go", "build"}},
		{SourceProjectPath, []string{"go", "build", "./..."}},
		{SourceHybrid, []string{"go", "build", "./..."}},
		{SourceWorkspace, []string{"go", "build", "./mymodule"}},
	}

	for _, tc := range testCases {
		t.Run(tc.source.String(), func(t *testing.T) {
			input := InputContext{Source: tc.source}
			result := prepareWorkspaceArgs(input, module, baseArgs)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Source %v: got %v, want %v", tc.source, result, tc.expected)
			}
		})
	}
}

// String method for InputSource to make test output more readable
func (s InputSource) String() string {
	switch s {
	case SourceUnknown:
		return "SourceUnknown"
	case SourceCode:
		return "SourceCode"
	case SourceProjectPath:
		return "SourceProjectPath"
	case SourceHybrid:
		return "SourceHybrid"
	case SourceWorkspace:
		return "SourceWorkspace"
	default:
		return "Unknown"
	}
}
