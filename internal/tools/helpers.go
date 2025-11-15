package tools

// prepareWorkspaceArgs prepares command arguments based on input source and module selection.
// For workspace sources, it handles module-specific or all-module targeting.
// For project sources, it adds recursive package selection.
// For code sources, it returns baseArgs unchanged.
//
// Parameters:
//   - input: The resolved input context
//   - module: Optional module name for workspace targeting (empty string for all modules)
//   - baseArgs: The base command arguments to extend
//
// Returns the prepared argument slice.
func prepareWorkspaceArgs(input InputContext, module string, baseArgs []string) []string {
	args := make([]string, len(baseArgs))
	copy(args, baseArgs)

	switch input.Source {
	case SourceWorkspace:
		if module != "" {
			// Target specific module in workspace
			args = append(args, module)
		} else {
			// Target all modules in workspace
			args = append(args, "./...")
		}
	case SourceProjectPath, SourceHybrid:
		// For project execution, target all packages
		args = append(args, "./...")
	case SourceCode:
		// For code execution, no additional args needed
		// The caller will handle specific file paths
	}

	return args
}
