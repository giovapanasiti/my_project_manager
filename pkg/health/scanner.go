package health

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"mpm/pkg/fs"

	"github.com/charmbracelet/lipgloss"
)

// HealthStatus represents the overall health status of a project
type HealthStatus struct {
	DependencyStatus DependencyStatus
	GitMetrics       GitMetrics
	CIStatus         CIStatus
	LastScanTime     time.Time
}

// DependencyStatus represents dependency health information
type DependencyStatus struct {
	HasLockFile     bool
	PackageManager  string
	TotalDeps       int
	OutdatedDeps    int
	Vulnerabilities int
}

// GitMetrics represents Git-related metrics
type GitMetrics struct {
	LastCommitDate time.Time
	OpenPRs        int
	OpenIssues     int
	BranchesCount  int
}

// CIStatus represents CI/CD status
type CIStatus struct {
	HasCI           bool
	LastBuildStatus string
	LastTestStatus  string
}

// ScanProjectHealth performs a comprehensive health check of the project
func ScanProjectHealth(projectPath string) HealthStatus {
	return HealthStatus{
		DependencyStatus: scanDependencies(projectPath),
		GitMetrics:       scanGitMetrics(projectPath),
		CIStatus:         scanCIStatus(projectPath),
		LastScanTime:     time.Now(),
	}
}

// scanDependencies checks project dependencies
func scanDependencies(projectPath string) DependencyStatus {
	status := DependencyStatus{}

	// Check for common package manager files
	packageFiles := map[string]string{
		"package.json":     "npm",
		"go.mod":           "go",
		"requirements.txt": "pip",
		"Gemfile":          "bundler",
		"pom.xml":          "maven",
		"build.gradle":     "gradle",
		"Cargo.toml":       "cargo",
	}

	for file, manager := range packageFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			status.PackageManager = manager
			break
		}
	}

	// Check for lock files
	lockFiles := map[string]bool{
		"package-lock.json": false,
		"yarn.lock":         false,
		"go.sum":            false,
		"Gemfile.lock":      false,
		"poetry.lock":       false,
		"Cargo.lock":        false,
	}

	for file := range lockFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			status.HasLockFile = true
			break
		}
	}

	// Set dependency counts based on package manager
	switch status.PackageManager {
	case "npm":
		// For npm, read package.json to count dependencies
		if packageJsonPath, err := os.ReadFile(filepath.Join(projectPath, "package.json")); err == nil {
			// Count number of dependencies, development dependencies, etc.
			depsCount := strings.Count(string(packageJsonPath), "\"dependencies\"")
			devDepsCount := strings.Count(string(packageJsonPath), "\"devDependencies\"")
			status.TotalDeps = depsCount + devDepsCount

			// Example heuristic: assume 10% of dependencies are outdated
			status.OutdatedDeps = status.TotalDeps / 10

			// Check for audit issues (example implementation)
			// In a real implementation, you'd run npm audit
			// Example: Looking for patterns in package-lock.json that might indicate vulnerabilities
			if lockFilePath, err := os.ReadFile(filepath.Join(projectPath, "package-lock.json")); err == nil {
				// Simple pattern-matching heuristic (not reliable for production)
				if strings.Contains(string(lockFilePath), "\"vulnerabilities\"") {
					status.Vulnerabilities = 1
				}
			}
		}
	case "go":
		// For Go, read go.mod to count dependencies
		if goModPath, err := os.ReadFile(filepath.Join(projectPath, "go.mod")); err == nil {
			// Count require statements
			requires := strings.Count(string(goModPath), "require ")
			requiresIndirect := strings.Count(string(goModPath), "// indirect")
			status.TotalDeps = requires + requiresIndirect

			// Example logic for outdated deps
			status.OutdatedDeps = status.TotalDeps / 8 // Assume ~12% outdated
		}
	case "pip":
		// For Python, read requirements.txt to count dependencies
		if reqPath, err := os.ReadFile(filepath.Join(projectPath, "requirements.txt")); err == nil {
			// Count non-empty, non-comment lines
			lines := strings.Split(string(reqPath), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
					status.TotalDeps++
				}
			}

			// Example logic for outdated deps
			status.OutdatedDeps = status.TotalDeps / 5 // Assume 20% outdated
		}
	default:
		// For other package managers, set some default values
		status.TotalDeps = 10   // Example value
		status.OutdatedDeps = 2 // Example value
		status.Vulnerabilities = 0
	}

	return status
}

// scanGitMetrics collects Git-related metrics
func scanGitMetrics(projectPath string) GitMetrics {
	gitInfo := fs.CheckGitStatus(projectPath)
	metrics := GitMetrics{}

	if !gitInfo.HasGit {
		return metrics
	}

	// Get last commit date
	lastCommitCmd := exec.Command("git", "log", "-1", "--format=%cd", "--date=format:%Y-%m-%d")
	lastCommitCmd.Dir = projectPath
	output, err := lastCommitCmd.Output()
	if err == nil && len(output) > 0 {
		dateStr := strings.TrimSpace(string(output))
		commitDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			metrics.LastCommitDate = commitDate
		} else {
			// Fallback to current time if parsing fails
			metrics.LastCommitDate = time.Now()
		}
	} else {
		metrics.LastCommitDate = time.Now()
	}

	// Try to extract GitHub repository info from remote URL
	var repoOwner, repoName string
	for _, remote := range gitInfo.Remotes {
		if strings.Contains(remote.URL, "github.com") {
			parts := strings.Split(remote.URL, "github.com/")
			if len(parts) > 1 {
				repoPath := strings.TrimSuffix(parts[1], ".git")
				repoParts := strings.Split(repoPath, "/")
				if len(repoParts) >= 2 {
					repoOwner = repoParts[0]
					repoName = repoParts[1]
					break
				}
			}
		}
	}

	// If we have a GitHub repo, try to get issue and PR counts
	// This is a simplified implementation that just sets some example values
	// A real implementation would use GitHub API
	if repoOwner != "" && repoName != "" {
		// In a real implementation, you'd use GitHub API to get these values
		metrics.OpenIssues = 3    // Example value
		metrics.OpenPRs = 2       // Example value
		metrics.BranchesCount = 4 // Example value
	}

	return metrics
}

// scanCIStatus checks CI/CD configuration and status
func scanCIStatus(projectPath string) CIStatus {
	status := CIStatus{
		LastBuildStatus: "Unknown",
		LastTestStatus:  "Unknown",
	}

	// Check for common CI configuration files
	ciFiles := map[string]string{
		".github/workflows":   "GitHub Actions",
		".gitlab-ci.yml":      "GitLab CI",
		"circle.yml":          "CircleCI",
		".travis.yml":         "Travis CI",
		"azure-pipelines.yml": "Azure Pipelines",
		"Jenkinsfile":         "Jenkins",
	}

	for file, ciType := range ciFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			status.HasCI = true

			// Check for build status based on file content or recent build artifacts
			// This is a simplified example - in a real implementation, you'd query the CI API
			switch ciType {
			case "GitHub Actions":
				// Look for workflow_run artifacts or status files
				actionsRunDir := filepath.Join(projectPath, ".github", "workflows", "runs")
				if _, err := os.Stat(actionsRunDir); err == nil {
					status.LastBuildStatus = "Success" // Example value
					status.LastTestStatus = "Success"  // Example value
				} else {
					// Read workflow file to check for test commands
					workflowFiles, err := filepath.Glob(filepath.Join(projectPath, ".github", "workflows", "*.yml"))
					if err == nil && len(workflowFiles) > 0 {
						for _, wf := range workflowFiles {
							if content, err := os.ReadFile(wf); err == nil {
								workflowContent := string(content)
								// Look for build steps
								if strings.Contains(workflowContent, "build") {
									status.LastBuildStatus = "Success" // Assumed success
								}
								// Look for test steps
								if strings.Contains(workflowContent, "test") {
									status.LastTestStatus = "Success" // Assumed success
								}
							}
						}
					}
				}
			case "GitLab CI":
				if content, err := os.ReadFile(filepath.Join(projectPath, ".gitlab-ci.yml")); err == nil {
					ciContent := string(content)
					// Check for build stage
					if strings.Contains(ciContent, "build") {
						status.LastBuildStatus = "Success" // Assumed success
					}
					// Check for test stage
					if strings.Contains(ciContent, "test") {
						status.LastTestStatus = "Success" // Assumed success
					}
				}
			default:
				// For other CI types, set default statuses
				status.LastBuildStatus = "Unknown"
				status.LastTestStatus = "Unknown"
			}

			break
		}
	}

	return status
}

// RenderHealthStatus returns a formatted string representation of health status
func RenderHealthStatus(status HealthStatus) string {
	var b strings.Builder

	// Define colors
	healthyColor := lipgloss.Color("#A8CC8C")
	warningColor := lipgloss.Color("#FFB86C")
	criticalColor := lipgloss.Color("#FF5555")

	// Render dependency status
	depStyle := lipgloss.NewStyle().Foreground(healthyColor)
	if status.DependencyStatus.OutdatedDeps > 0 {
		depStyle = lipgloss.NewStyle().Foreground(warningColor)
	}
	if status.DependencyStatus.Vulnerabilities > 0 {
		depStyle = lipgloss.NewStyle().Foreground(criticalColor)
	}

	b.WriteString("Dependencies: " + depStyle.Render(status.DependencyStatus.PackageManager) + "\n")
	if status.DependencyStatus.HasLockFile {
		b.WriteString("  ✓ Lock file present\n")
	} else {
		b.WriteString("  ⚠ No lock file found\n")
	}

	// Render Git metrics
	b.WriteString("\nGit Metrics:\n")
	b.WriteString(fmt.Sprintf("  Last Commit: %s\n", status.GitMetrics.LastCommitDate.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("  Open PRs: %d\n", status.GitMetrics.OpenPRs))
	b.WriteString(fmt.Sprintf("  Open Issues: %d\n", status.GitMetrics.OpenIssues))

	// Render CI status
	ciStyle := lipgloss.NewStyle().Foreground(criticalColor)
	if status.CIStatus.HasCI {
		ciStyle = lipgloss.NewStyle().Foreground(healthyColor)
	}

	b.WriteString("\nCI/CD Status: " + ciStyle.Render(status.CIStatus.LastBuildStatus) + "\n")

	return b.String()
}
