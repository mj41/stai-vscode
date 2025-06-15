package version

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// Version is the semantic version
const Version = "0.1.0"

// BuildInfo contains version and build information
type BuildInfo struct {
	Version   string
	GitCommit string
	GitDirty  bool
	BuildTime string
	GoVersion string
	Module    string
}

// GetBuildInfo returns build information using debug.ReadBuildInfo()
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.Module = buildInfo.Main.Path

		// Parse VCS information from build settings
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.GitCommit = setting.Value
				if len(info.GitCommit) > 12 {
					info.GitCommit = info.GitCommit[:12] // Shorten to 12 chars
				}
			case "vcs.modified":
				info.GitDirty = setting.Value == "true"
			case "vcs.time":
				info.BuildTime = setting.Value
			}
		}
	}

	return info
}

// FormatVersion returns a formatted version string for a tool
func FormatVersion(toolName string) string {
	info := GetBuildInfo()

	versionStr := fmt.Sprintf("%s version %s", toolName, info.Version)

	if info.GitCommit != "" {
		versionStr += fmt.Sprintf(" (%s", info.GitCommit)
		if info.GitDirty {
			versionStr += "-dirty"
		}
		versionStr += ")"
	}

	if info.BuildTime != "" {
		versionStr += fmt.Sprintf(" built at %s", info.BuildTime)
	}

	versionStr += fmt.Sprintf(" with %s", info.GoVersion)

	return versionStr
}

// ShowVersionAndExit displays version information and exits
func ShowVersionAndExit(toolName string) {
	fmt.Println(FormatVersion(toolName))
	os.Exit(0)
}
