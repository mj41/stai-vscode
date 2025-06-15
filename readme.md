Documentation and tools to setup development environment for Infinite Process Modelling and Tate (Shiftate Tate AI).

# Warning

This setup will allow AI to run commands on your computer. Make sure you trust the AI and understand the implications of running AI-generated commands.

# Pre-requisites

[Install system and create `stai` user](#machine-installation) and [configure git to use HTTPS instead of SSH for cloning repositories, or set up SSH keys for GitHub](#github-ssh-setup).

# Usage

To use `ws-config-gen`, navigate to the project directory and run the following commands:
```shell
cat /etc/fedora-release
# Fedora release 42 (Adams) - Only supported/tested OS version.
whoami
# stai - You should be logged in as user `stai` (Shiftate Tate AI user) as described below.
mkdir -p ~/work-stai && cd ~/work-stai
git clone https://github.com/mj41/stai-vscode.git
cd ~/work-stai/stai-vscode
go run ./cmd/ws-config-gen
code-insiders ~/work-stai/vscode/stai-all.code-workspace
```

# Use cases

## Machine installation

- User installs Fedora Linux and logs in with their personal user to
    - install git `sudo dnf install git`
    - install [Visual Studio Code Insiders](https://code.visualstudio.com/insiders/)
    - create a user account with Full Name `Tate Shiftate AI` and Username `stai`
- User logs in to Fedora as Tate Shiftate AI (`stai`) user
- User starts Visual Studio Code Insiders
- User opens GitHub Copilot in Visual Studio Code Insiders
- User authenticates GitHub Copilot with their GitHub account
- User changes Copilot to "Agent" and "Claude Sonnet 4"
- User sets up GitHub SSH key or forces git to use HTTPS for cloning repositories (see below)

## vscode setup

- User creates new `base-stai` directory e.g. `~/work-stai && cd ~/work-stai`
- User clones repository `git clone git@github.com:mj41/stai-vscode.git`
- User runs `cd ~/work-stai/stai-vscode && go run ./cmd/ws-config-gen`
- `ws-config-gen` checks that:
  - the `base-stai` directory is not $HOME directory
  - the `base-stai` directory is empty except for `stai-vscode` subdirectory
- `ws-config-gen` creates the following directories:
  - `~/work-stai/vscode`
  - `~/work-stai/stai-temp`
  - `~/work-stai/stai-temp/aitsk`
- `ws-config-gen` initializes git repository in `~/work-stai/stai-temp` and creates an empty initial commit with `readme.md` file based on embedded readme template (see [readme.md.tmpl](./internal/templates/readme.md.tmpl) for reference)
- `ws-config-gen` clones git repositories mentioned in embedded configuration (see [repos.json](./internal/config/repos.json) for reference)
- `ws-config-gen` creates a workspace file `~/work-stai/vscode/stai-all.code-workspace` based on embedded configuration and workspace template (see [stai-all.code-workspace.tmpl](./internal/templates/stai-all.code-workspace.tmpl) for reference). Paths to workspace folders are relative to `~/work-stai/vscode` directory
- User opens `~/work-stai/vscode/stai-all.code-workspace` in Visual Studio Code Insiders
- User starts to assign tasks to AI (Copilot)

# GitHub SSH setup

If your `stai` user's SSH keys are not set up, you can force git to use HTTPS instead of SSH for GitHub repositories.
```shell
whoami
# stai
git config --global url."https://github.com/".insteadOf "git@github.com:"
```
Alternatively, you can set up your [SSH keys for GitHub](https://github.com/settings/keys) as described in the [Connecting to GitHub with SSH](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).


# Tools

## ws-config-gen

`ws-config-gen` is a tool for generating Visual Studio Code workspace configuration files. It helps in setting up the development environment by creating a workspace file that includes all necessary settings and configurations for the project.

This will generate a VS Code workspace configuration file in the `../vscode` directory. All paths in the workspace file will be absolute except `folders` paths, which will be relative to the workspace file location.

Tool will check that it was started from `stai-vscode` directory.

Tool will check that required binaries are installed and that the user is logged in as `stai` user.

Tool will exit with exit code 1 on any error or warning. You can use the `--force` flag to ignore warnings and continue execution:

- `--force` - Ignore up to one warning and continue execution (safer default)
- `--force=N` - Ignore up to N warnings and continue execution (e.g., `--force=2` ignores the first 2 warnings)
- `--force=-1` - Ignore all warnings and continue execution (unlimited)

Examples:
```shell
# Ignore the first warning only (safer)
go run ./cmd/ws-config-gen --force

# Ignore the first 2 warnings only
go run ./cmd/ws-config-gen --force=2

# Ignore all warnings (use with caution)
go run ./cmd/ws-config-gen --force=-1
```
