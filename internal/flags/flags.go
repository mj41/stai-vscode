package flags

import (
	"flag"
	"fmt"
	"os"

	"github.com/mj41/stai-vscode/internal/version"
)

// CommonFlags contains the standard flags that all tools should support
type CommonFlags struct {
	ShowVersion bool
	ShowHelp    bool
	ShowReadme  bool
}

// FlagConfig contains configuration for setting up common flags
type FlagConfig struct {
	ToolName      string
	Usage         string
	Description   string
	HasReadme     bool // Whether this tool has embedded documentation
	ReadmeContent string
	HelpContent   string // Full help content for --help flag
}

// SetupCommonFlags sets up standard flags for a tool
func SetupCommonFlags(config FlagConfig) *CommonFlags {
	flags := &CommonFlags{}

	flag.BoolVar(&flags.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&flags.ShowHelp, "help", false, "Show usage information")

	if config.HasReadme {
		flag.BoolVar(&flags.ShowReadme, "readme", false, "Show full documentation")
	}

	// Set up custom usage function
	flag.Usage = func() {
		ShowHelp(config.ToolName, config.Usage, config.Description)
		if config.HasReadme {
			fmt.Println("\nUse --readme to show full documentation")
		}
		fmt.Println("\nUse --version to show version information")
	}

	return flags
}

// HandleCommonFlags processes common flags and exits if appropriate
func HandleCommonFlags(flags *CommonFlags, config FlagConfig) {
	if flags.ShowVersion {
		version.ShowVersionAndExit(config.ToolName)
	}

	if flags.ShowHelp {
		if config.HelpContent != "" {
			// Use full help content if available
			fmt.Print(config.HelpContent)
		} else {
			// Fallback to minimal usage
			flag.Usage()
		}
		os.Exit(0)
	}

	if flags.ShowReadme {
		if config.ReadmeContent != "" {
			fmt.Print(config.ReadmeContent)
		} else {
			fmt.Printf("No documentation available for %s\n", config.ToolName)
		}
		os.Exit(0)
	}
}

// ShowHelp displays help information consistently across tools
func ShowHelp(tool, usage, description string) {
	fmt.Printf("Usage: %s\n", usage)
	fmt.Printf("%s\n", description)
}
