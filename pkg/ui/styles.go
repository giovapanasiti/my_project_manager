package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styling constants for the UI
var (
	// TitleStyle for application title
	TitleStyle = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)

	// ItemStyle for normal list items
	ItemStyle = lipgloss.NewStyle().PaddingLeft(4)

	// SelectedItemStyle for selected list items
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#7D56F4"))

	// HelpStyle for help text
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	// CategoryStyle for project categories
	CategoryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C"))

	// PathStyle for project paths
	PathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#B2B2B2"))

	// FileChartStyle for file chart display
	FileChartStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).PaddingLeft(4)

	// FileTypeStyle for file types
	FileTypeStyle = lipgloss.NewStyle().Bold(true)
)
