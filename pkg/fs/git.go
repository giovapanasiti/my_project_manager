package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GitInfo represents Git repository information
type GitInfo struct {
	HasGit  bool
	Remotes []GitRemote
}

// GitRemote represents a Git remote
type GitRemote struct {
	Name string
	URL  string
}

// CheckGitStatus checks if a directory is a Git repository and returns Git information
func CheckGitStatus(projectPath string) GitInfo {
	gitInfo := GitInfo{
		HasGit:  false,
		Remotes: []GitRemote{},
	}

	// Check if .git directory exists
	gitDir := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return gitInfo
	}

	// Set HasGit to true since .git directory exists
	gitInfo.HasGit = true

	// Get remotes
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return gitInfo
	}

	// Parse remotes
	remoteMap := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format is typically: "origin\thttps://github.com/user/repo.git (fetch)"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]

			// Only add fetch entries (avoid duplicates with push entries)
			if strings.Contains(line, "(fetch)") {
				remoteMap[name] = url
			}
		}
	}

	// Convert map to slice
	for name, url := range remoteMap {
		gitInfo.Remotes = append(gitInfo.Remotes, GitRemote{Name: name, URL: url})
	}

	return gitInfo
}

// RenderGitInfo renders Git information as a formatted string
func RenderGitInfo(gitInfo GitInfo) string {
	var b strings.Builder

	// Import UI styles
	// We can't directly import the UI package due to import cycle
	// So we'll use string constants for now
	gitPresentColor := "#A8CC8C"
	gitAbsentColor := "#FF5555"
	gitRemoteColor := "#7D56F4"

	// Git status indicator with styled output
	if gitInfo.HasGit {
		gitStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(gitPresentColor)).Bold(true).Render("✅")
		b.WriteString("  Git: " + gitStatus + "\n")
	} else {
		gitStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(gitAbsentColor)).Bold(true).Render("❌")
		b.WriteString("  Git: " + gitStatus + "\n")
		return b.String()
	}

	// Remotes with styled output
	if len(gitInfo.Remotes) > 0 {
		b.WriteString("  Remotes:\n")
		for _, remote := range gitInfo.Remotes {
			remoteName := lipgloss.NewStyle().Foreground(lipgloss.Color(gitRemoteColor)).Render(remote.Name)
			b.WriteString(fmt.Sprintf("    %s: %s\n", remoteName, remote.URL))
		}
	}

	return b.String()
}
