package ui

import (
	"mpm/pkg/health"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// HealthStyle defines styling for health indicators
var (
	HealthTitleStyle = lipgloss.NewStyle().Bold(true).MarginLeft(2).Foreground(lipgloss.Color("#7D56F4"))
	HealthStyle      = lipgloss.NewStyle().MarginLeft(4)
	IndicatorStyle   = lipgloss.NewStyle().Bold(true)
	SectionStyle     = lipgloss.NewStyle().Bold(true).MarginLeft(2).Foreground(lipgloss.Color("#FFB86C"))

	HealthyColor  = lipgloss.Color("#A8CC8C")
	WarningColor  = lipgloss.Color("#FFB86C")
	CriticalColor = lipgloss.Color("#FF5555")
	NeutralColor  = lipgloss.Color("#B2B2B2")

	// New styles for health dashboard
	DashboardBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginLeft(2).
			MarginRight(2)

	KeyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ValueStyle = lipgloss.NewStyle()
)

// formatCount converts int to string, with color styling based on value
func formatCount(count int, zeroIsOk bool) string {
	countStr := strconv.Itoa(count)
	if count == 0 && zeroIsOk {
		return lipgloss.NewStyle().Foreground(HealthyColor).Render(countStr)
	} else if count == 0 && !zeroIsOk {
		return lipgloss.NewStyle().Foreground(NeutralColor).Render(countStr)
	} else if count > 5 {
		return lipgloss.NewStyle().Foreground(CriticalColor).Render(countStr)
	} else if count > 2 {
		return lipgloss.NewStyle().Foreground(WarningColor).Render(countStr)
	}
	return lipgloss.NewStyle().Foreground(HealthyColor).Render(countStr)
}

// RenderHealthDashboard returns a formatted health dashboard view
func (m ListModel) RenderHealthDashboard() string {
	if m.SelectedItem == nil {
		return ""
	}

	// Use the cached health status if available, otherwise scan the project
	healthStatus := m.HealthStatus
	if !m.HealthScanned {
		healthStatus = health.ScanProjectHealth(m.SelectedItem.Path)
	}

	// Create styled health indicators
	depIndicator := IndicatorStyle.Copy().Foreground(HealthyColor).Render("⬤")
	if healthStatus.DependencyStatus.OutdatedDeps > 0 {
		depIndicator = IndicatorStyle.Copy().Foreground(WarningColor).Render("⬤")
	}
	if healthStatus.DependencyStatus.Vulnerabilities > 0 {
		depIndicator = IndicatorStyle.Copy().Foreground(CriticalColor).Render("⬤")
	}

	ciIndicator := IndicatorStyle.Copy().Foreground(CriticalColor).Render("⬤")
	if healthStatus.CIStatus.HasCI {
		if healthStatus.CIStatus.LastBuildStatus == "Success" && healthStatus.CIStatus.LastTestStatus == "Success" {
			ciIndicator = IndicatorStyle.Copy().Foreground(HealthyColor).Render("⬤")
		} else if healthStatus.CIStatus.LastBuildStatus == "Success" || healthStatus.CIStatus.LastTestStatus == "Success" {
			ciIndicator = IndicatorStyle.Copy().Foreground(WarningColor).Render("⬤")
		}
	}

	gitIndicator := IndicatorStyle.Copy().Foreground(CriticalColor).Render("⬤")
	if m.GitInfo.HasGit {
		gitIndicator = IndicatorStyle.Copy().Foreground(HealthyColor).Render("⬤")
	}

	// Format dependency section
	dependencySection := lipgloss.JoinVertical(lipgloss.Left,
		SectionStyle.Render("Dependencies"),
		HealthStyle.Render(depIndicator+" "+healthStatus.DependencyStatus.PackageManager),
		HealthStyle.Render(KeyStyle.Render("   Total: ")+formatCount(healthStatus.DependencyStatus.TotalDeps, true)),
		HealthStyle.Render(KeyStyle.Render("   Outdated: ")+formatCount(healthStatus.DependencyStatus.OutdatedDeps, true)),
		HealthStyle.Render(KeyStyle.Render("   Vulnerabilities: ")+formatCount(healthStatus.DependencyStatus.Vulnerabilities, true)),
		"", // Add extra newline for better spacing
	)

	// Format Git section
	gitSection := lipgloss.JoinVertical(lipgloss.Left,
		SectionStyle.Render("Git Status"),
		HealthStyle.Render(gitIndicator+" Git Integration"),
		HealthStyle.Render(KeyStyle.Render("   Last Commit: ")+healthStatus.GitMetrics.LastCommitDate.Format("2006-01-02")),
		HealthStyle.Render(KeyStyle.Render("   Open PRs: ")+formatCount(healthStatus.GitMetrics.OpenPRs, true)),
		HealthStyle.Render(KeyStyle.Render("   Open Issues: ")+formatCount(healthStatus.GitMetrics.OpenIssues, true)),
		"", // Add extra newline for better spacing
	)

	// Format CI/CD section
	ciSection := lipgloss.JoinVertical(lipgloss.Left,
		SectionStyle.Render("CI/CD Status"),
		HealthStyle.Render(ciIndicator+" Pipeline Integration"),
		HealthStyle.Render(KeyStyle.Render("   Build Status: ")+formatCIStatus(healthStatus.CIStatus.LastBuildStatus)),
		HealthStyle.Render(KeyStyle.Render("   Test Status: ")+formatCIStatus(healthStatus.CIStatus.LastTestStatus)),
		"", // Add extra newline for better spacing
	)

	// Add navigation hint
	scrollHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B2B2B2")).
		Align(lipgloss.Center).
		Render("(Use ↑ and ↓ to scroll)")

	// Put it all together in a nice box
	dashboardContent := lipgloss.JoinVertical(lipgloss.Left,
		HealthTitleStyle.Render("Project Health Dashboard"),
		scrollHint,
		"",
		dependencySection,
		gitSection,
		ciSection,
	)

	return DashboardBox.Render(dashboardContent)
}

// formatCIStatus formats CI status with appropriate colors
func formatCIStatus(status string) string {
	switch status {
	case "Success":
		return lipgloss.NewStyle().Foreground(HealthyColor).Render(status)
	case "Failed":
		return lipgloss.NewStyle().Foreground(CriticalColor).Render(status)
	case "Running":
		return lipgloss.NewStyle().Foreground(WarningColor).Render(status)
	default:
		return lipgloss.NewStyle().Foreground(NeutralColor).Render(status)
	}
}
