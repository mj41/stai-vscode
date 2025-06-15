package main

import (
	_ "embed"
)

// Embedded template files for the workspace generation tool.
// These templates are embedded into the binary at build time,
// eliminating the need for external template files.

//go:embed templates/stai-all.code-workspace.tmpl
var workspaceTemplate string

//go:embed templates/readme.md.tmpl
var readmeTemplate string

// getWorkspaceTemplate returns the embedded VS Code workspace template.
// This template is used to generate the .code-workspace file with
// proper folder structure and VS Code settings.
func getWorkspaceTemplate() string {
	return workspaceTemplate
}

// getReadmeTemplate returns the embedded readme.md template.
// This template is used to create the initial readme.md file
// in the stai-temp repository.
func getReadmeTemplate() string {
	return readmeTemplate
}
