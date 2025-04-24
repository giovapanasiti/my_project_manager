# My Project Manager (MPM)

A CLI tool to easily manage and navigate your projects from the terminal.

## Overview

MPM (My Project Manager) is a command-line tool that helps you organize your projects by storing their locations and quickly navigating to them. It features a beautiful interactive terminal UI, project categorization, and quick integration with popular editors.

## Features

- **Manage projects**: Add, remove, and list your projects
- **Categorize projects**: Organize projects by categories
- **Quick navigation**: Jump to project directories with a single command
- **Interactive mode**: Navigate projects with a beautiful TUI (Text User Interface)
- **Editor integration**: Open projects directly in VS Code, Zed, Cursor, Neovim, Sublime Text, IntelliJ IDEA, or PyCharm
- **File explorer integration**: Open projects in Finder (macOS), Explorer (Windows), or file manager (Linux)

## Installation

### Prerequisites

- Go 1.16 or higher
- Git

### Building from source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/mpm.git
   cd mpm
   ```

2. Install dependencies:
   ```bash
   go mod init mpm
   go get github.com/charmbracelet/bubbles/list
   go get github.com/charmbracelet/bubbles/textinput
   go get github.com/charmbracelet/bubbles/key
   go get github.com/charmbracelet/bubbletea
   go get github.com/charmbracelet/lipgloss
   go get github.com/spf13/cobra
   ```

3. Build the binary:
   ```bash
   go build -o mpm
   ```

4. Move the binary to a directory in your PATH:
   ```bash
   sudo mv mpm /usr/local/bin/
   ```

### Shell Integration

To enable directory navigation, add this function to your `~/.bashrc`, `~/.zshrc`, or equivalent shell configuration file:

```bash
# mpm wrapper function
mpm() {
  if [ "$1" = "go" ] || [ "$1" = "i" ]; then
    output=$(command mpm "$@")
    if [[ $output == cd* ]]; then
      eval "$output"
    else
      echo "$output"
    fi
  else
    command mpm "$@"
  fi
}
```

Then reload your shell:
```bash
source ~/.bashrc  # or source ~/.zshrc
```

## Usage

### Adding projects

You can add projects in two ways:

1. Specifying name and path:
   ```bash
   mpm add -n project_name -p /path/to/project -c category
   ```

2. Using current directory (new in latest version):
   ```bash
   cd /path/to/your/project
   mpm add -w -c category
   ```
   This automatically uses the current directory as path and the folder name as project name.

### Removing projects

```bash
mpm remove project_name
```

### Listing projects

```bash
mpm list
```

### Navigating to a project

```bash
mpm go project_name
```

### Interactive mode

```bash
mpm i
```

## Interactive Mode Controls

### Main List View

- `↑/k` and `↓/j`: Navigate up and down
- `/`: Filter projects
- `Enter`: Select project for actions
- `a`: Add new project
- `d`: Delete selected project
- `q` or `Ctrl+C`: Quit

### Action View (after selecting a project)

- `g` or `Enter`: Navigate to project directory
- `v`: Open in VS Code
- `z`: Open in Zed
- `c`: Open in Cursor
- `n`: Open in Neovim
- `f`: Open in Finder/File Explorer
- `s`: Open in Sublime Text
- `i`: Open in IntelliJ IDEA
- `p`: Open in PyCharm
- `d`: Delete project
- `Esc` or `q`: Back to list

### Add Project Form

- `Tab`/`Shift+Tab`: Navigate between fields
- `Enter`: Move to next field or submit form
- `Esc`: Cancel

## Configuration

MPM stores your projects in a configuration file located at:

```
~/.mpm/config.json
```

You can edit this file manually if needed, but it's recommended to use the CLI commands or interactive mode.

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.