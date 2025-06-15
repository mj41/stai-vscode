package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/mj41/stai-vscode/internal/flags"
)

//go:embed config/repos.json
var embeddedConfig []byte

// Default directory permissions for created directories
const defaultDirPerms = 0750

// Config represents the repositories configuration
type Config struct {
	Repos []Repository `json:"repos"`
}

// Repository represents a single repository configuration
type Repository struct {
	Name    string  `json:"name"`
	GitRepo *string `json:"git-repo"`
	Type    string  `json:"type"`
}

// TemplateData contains data for template processing
type TemplateData struct {
	Folders     string
	BaseWorkDir string
}

// FolderEntry represents a folder in the VS Code workspace
type FolderEntry struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

// ForceFlag implements flag.Value to handle --force and --force=N syntax
type ForceFlag struct {
	enabled bool
	level   int
}

func (f *ForceFlag) String() string {
	if !f.enabled {
		return "false"
	}
	if f.level == -1 {
		return "true"
	}
	return strconv.Itoa(f.level)
}

func (f *ForceFlag) Set(value string) error {
	f.enabled = true
	if value == "" || value == "true" {
		f.level = 1 // ignore up to one warning by default
		return nil
	}
	if value == "false" {
		f.enabled = false
		f.level = 0
		return nil
	}

	level, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid force level '%s', must be a number or -1 for unlimited", value)
	}
	if level == -1 {
		f.level = -1 // unlimited warnings (explicit -1)
	} else if level < 0 {
		return fmt.Errorf("invalid force level '%d', must be 0 or positive, or -1 for unlimited", level)
	} else {
		f.level = level
	}
	return nil
}

func (f *ForceFlag) IsBoolFlag() bool {
	return true
}

var (
	forceFlag    ForceFlag
	warningCount int
)

func main() {
	// Setup command line flags
	flagConfig := flags.FlagConfig{
		ToolName:    "ws-config-gen",
		Usage:       "ws-config-gen [--force[=N|-1]] [--version] [--help]",
		Description: "Generate Visual Studio Code workspace configuration for Tate AI development environment",
		HasReadme:   false,
	}

	commonFlags := flags.SetupCommonFlags(flagConfig)

	// Add tool-specific flags
	flag.Var(&forceFlag, "force", "Force execution, ignore warnings. Default ignores 1 warning. Use --force=N for specific count, --force=-1 for unlimited")

	flag.Parse()

	// Force flag parsing is handled automatically by the ForceFlag.Set method

	// Handle common flags
	flags.HandleCommonFlags(commonFlags, flagConfig)

	// Main execution
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Setup complete")
}

func run() error {
	fmt.Println("Checking user and environment...")

	// Check current user
	if err := checkUser(); err != nil {
		return err
	}

	// Check required binaries
	if err := checkBinaries(); err != nil {
		return err
	}

	// Validate current directory
	workDir, err := validateWorkingDirectory()
	if err != nil {
		return err
	}

	// Determine base directory
	baseDir := filepath.Dir(workDir)

	// Validate base directory
	if err := validateBaseDirectory(baseDir); err != nil {
		return err
	}

	fmt.Println("Creating directories...")

	// Create required directories
	if err := createDirectories(baseDir); err != nil {
		return err
	}

	// Initialize stai-temp git repository
	if err := initStaiTempRepo(baseDir); err != nil {
		return err
	}

	fmt.Println("Cloning repositories...")

	// Load repository configuration
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Clone repositories
	if err := cloneRepositories(baseDir, config); err != nil {
		return err
	}

	fmt.Println("Generating workspace file...")

	// Generate workspace file
	if err := generateWorkspace(baseDir, config); err != nil {
		return err
	}

	return nil
}

// canSkipWarning checks if we can skip a warning based on force level
func canSkipWarning() bool {
	if !forceFlag.enabled {
		return false // no force flag
	}
	if forceFlag.level == -1 {
		return true // unlimited warnings
	}
	if warningCount < forceFlag.level {
		warningCount++
		return true
	}
	return false
}

func checkUser() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	if currentUser.Username != "stai" {
		if canSkipWarning() {
			fmt.Printf("Warning: Current user is '%s', expected 'stai' (continuing due to --force)\n", currentUser.Username)
		} else {
			return fmt.Errorf("current user is '%s', expected 'stai'. Use --force to ignore this check", currentUser.Username)
		}
	}

	return nil
}

func checkBinaries() error {
	binaries := []string{"git", "code-insiders"}

	for _, binary := range binaries {
		if _, err := exec.LookPath(binary); err != nil {
			if canSkipWarning() {
				fmt.Printf("Warning: Binary '%s' not found in PATH (continuing due to --force)\n", binary)
			} else {
				return fmt.Errorf("required binary '%s' not found in PATH. Use --force to ignore this check", binary)
			}
		}
	}

	return nil
}

func validateWorkingDirectory() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	if filepath.Base(workDir) != "stai-vscode" {
		return "", fmt.Errorf("current directory must be named 'stai-vscode', got '%s'", filepath.Base(workDir))
	}

	return workDir, nil
}

func validateBaseDirectory(baseDir string) error {
	// Check that base directory is not $HOME
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	if baseDir == homeDir {
		return fmt.Errorf("base directory cannot be the home directory (%s)", homeDir)
	}

	// Check that base directory is under home directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for base directory: %w", err)
	}

	absHomeDir, err := filepath.Abs(homeDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for home directory: %w", err)
	}

	relPath, err := filepath.Rel(absHomeDir, absBaseDir)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("base directory must be under home directory (%s), got %s", homeDir, baseDir)
	}

	// Check that base directory is empty except for stai-vscode
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() != "stai-vscode" {
			if canSkipWarning() {
				fmt.Printf("Warning: Base directory contains additional files/directories (continuing due to --force)\n")
				break
			} else {
				return fmt.Errorf("base directory must be empty except for 'stai-vscode' directory. Found: %s. Use --force to ignore this check", entry.Name())
			}
		}
	}

	return nil
}

func createDirectories(baseDir string) error {
	dirs := []string{
		filepath.Join(baseDir, "vscode"),
		filepath.Join(baseDir, "stai-temp"),
		filepath.Join(baseDir, "stai-temp", "aitsk"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, defaultDirPerms); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func initStaiTempRepo(baseDir string) error {
	staiTempDir := filepath.Join(baseDir, "stai-temp")

	// Check if already a git repository
	if _, err := os.Stat(filepath.Join(staiTempDir, ".git")); err == nil {
		fmt.Printf("stai-temp is already a git repository, skipping initialization\n")
		return nil
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = staiTempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository in stai-temp: %w", err)
	}

	// Create readme.md from template
	readmeContent := getReadmeTemplate()
	readmePath := filepath.Join(staiTempDir, "readme.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create readme.md: %w", err)
	}

	// Add and commit
	cmd = exec.Command("git", "add", "readme.md")
	cmd.Dir = staiTempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add readme.md to git: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit - stai-temp workspace")
	cmd.Dir = staiTempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit initial files: %w", err)
	}

	return nil
}

func loadConfig() (*Config, error) {
	var config Config
	if err := json.Unmarshal(embeddedConfig, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config: %w", err)
	}

	return &config, nil
}

func cloneRepositories(baseDir string, config *Config) error {
	for _, repo := range config.Repos {
		repoDir := filepath.Join(baseDir, repo.Name)

		// Skip if directory already exists
		if _, err := os.Stat(repoDir); err == nil {
			fmt.Printf("Repository %s already exists, skipping\n", repo.Name)
			continue
		}

		switch repo.Type {
		case "git-repo":
			if repo.GitRepo == nil {
				return fmt.Errorf("git-repo type requires git-repo URL for %s", repo.Name)
			}

			cmd := exec.Command("git", "clone", *repo.GitRepo, repoDir)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to clone repository %s: %w", repo.Name, err)
			}

		case "local-git-repo":
			// For local-git-repo, we already handled stai-temp above
			if repo.Name == "stai-temp" {
				continue
			}

			if err := os.MkdirAll(repoDir, defaultDirPerms); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", repo.Name, err)
			}

			cmd := exec.Command("git", "init")
			cmd.Dir = repoDir
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to initialize git repository for %s: %w", repo.Name, err)
			}

		default:
			return fmt.Errorf("unknown repository type %s for %s", repo.Type, repo.Name)
		}
	}

	return nil
}

func generateWorkspace(baseDir string, config *Config) error {
	// Use embedded workspace template
	tmpl, err := template.New("workspace").Parse(getWorkspaceTemplate())
	if err != nil {
		return fmt.Errorf("failed to parse workspace template: %w", err)
	}

	// Generate folders JSON
	var folders []FolderEntry
	for _, repo := range config.Repos {
		folders = append(folders, FolderEntry{
			Path: "../" + repo.Name,
		})
	}

	foldersJSON, err := json.MarshalIndent(folders, "\t", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal folders JSON: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		Folders:     string(foldersJSON),
		BaseWorkDir: baseDir,
	}

	// Generate workspace file
	workspacePath := filepath.Join(baseDir, "vscode", "stai-all.code-workspace")
	file, err := os.Create(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to create workspace file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute workspace template: %w", err)
	}

	return nil
}
