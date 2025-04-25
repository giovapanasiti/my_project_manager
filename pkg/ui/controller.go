package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mpm/pkg/config"
	"mpm/pkg/fs"
	"mpm/pkg/health"
)

// Custom message type to hold command output
type QuitMsg string

// Init initializes the model
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update function handles the main application logic
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShowForm {
			return handleFormView(m, msg)
		} else if m.ShowActions {
			return handleActionsView(m, msg)
		} else {
			return handleListView(m, msg)
		}

	case tea.WindowSizeMsg:
		// Store window dimensions in the model
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height

		// Adjust the list size
		headerHeight := lipgloss.Height(m.HeaderView())
		footerHeight := lipgloss.Height(m.FooterView())
		verticalMarginHeight := headerHeight + footerHeight

		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v-verticalMarginHeight)
	}

	// Handle list updates by default
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// handleFormView handles keyboard events in the form view
func handleFormView(m ListModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.ShowForm = false
		return m, nil

	case "tab", "shift+tab":
		// Handle tab navigation in the form
		if msg.String() == "tab" {
			m.FormFocused = (m.FormFocused + 1) % len(m.FormInputs)
		} else {
			m.FormFocused = (m.FormFocused - 1 + len(m.FormInputs)) % len(m.FormInputs)
		}

		for i := range m.FormInputs {
			if i == m.FormFocused {
				m.FormInputs[i].Focus()
			} else {
				m.FormInputs[i].Blur()
			}
		}

		return m, nil

	case "enter":
		if m.FormFocused == len(m.FormInputs)-1 {
			// Submit form on Enter when on last field
			name := strings.TrimSpace(m.FormInputs[0].Value())
			path := strings.TrimSpace(m.FormInputs[1].Value())
			category := strings.TrimSpace(m.FormInputs[2].Value())

			if name != "" && path != "" {
				config.AddProject(name, path, category)

				// Reload projects for the list
				refreshedConfig := config.LoadConfig()
				m.ProjectItems = []list.Item{}

				for _, p := range refreshedConfig.Projects {
					cat := p.Category
					if cat == "" {
						cat = "Uncategorized"
					}
					m.ProjectItems = append(m.ProjectItems, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})
				}

				// Apply current sort order to the updated project list
				if m.SortOrder == "asc" {
					sort.Slice(m.ProjectItems, func(i, j int) bool {
						return m.ProjectItems[i].(ProjectItem).Name < m.ProjectItems[j].(ProjectItem).Name
					})
				} else {
					sort.Slice(m.ProjectItems, func(i, j int) bool {
						return m.ProjectItems[i].(ProjectItem).Name > m.ProjectItems[j].(ProjectItem).Name
					})
				}

				m.List.SetItems(m.ProjectItems)

				// Reset form
				for i := range m.FormInputs {
					m.FormInputs[i].SetValue("")
				}
				m.FormFocused = 0
				m.FormInputs[0].Focus()
				for i := 1; i < len(m.FormInputs); i++ {
					m.FormInputs[i].Blur()
				}

				m.ShowForm = false
			}

			return m, nil
		} else {
			// Move to next field on Enter
			m.FormFocused = (m.FormFocused + 1) % len(m.FormInputs)
			for i := range m.FormInputs {
				if i == m.FormFocused {
					m.FormInputs[i].Focus()
				} else {
					m.FormInputs[i].Blur()
				}
			}
			return m, nil
		}
	}

	// Handle text input
	m.FormInputs[m.FormFocused], cmd = m.FormInputs[m.FormFocused].Update(msg)
	return m, cmd
}

// handleActionsView handles keyboard events in the action view
func handleActionsView(m ListModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actionKeys := NewActionKeyMap()

	switch {
	case key.Matches(msg, actionKeys.Back):
		m.ShowActions = false
		// Reset health scan status when going back to the list
		m.HealthScanned = false
		// Reset scroll position
		m.ScrollOffset = 0
		return m, nil

	// Handle arrow keys for scrolling
	case msg.String() == "up":
		// Scroll up (decrease offset)
		if m.ScrollOffset > 0 {
			m.ScrollOffset--
		}
		return m, nil

	case msg.String() == "down":
		// Scroll down (increase offset)
		m.ScrollOffset++
		return m, nil

	// Add page up/down for faster scrolling
	case msg.String() == "pgup":
		// Page up - scroll up by ~half screen
		pageSize := 10
		if m.WindowHeight > 0 {
			pageSize = m.WindowHeight / 2
		}

		if m.ScrollOffset >= pageSize {
			m.ScrollOffset -= pageSize
		} else {
			m.ScrollOffset = 0 // Go to top if near the top
		}
		return m, nil

	case msg.String() == "pgdown":
		// Page down - scroll down by ~half screen
		pageSize := 10
		if m.WindowHeight > 0 {
			pageSize = m.WindowHeight / 2
		}
		m.ScrollOffset += pageSize
		// View method will handle maximum bounds checking
		return m, nil

	// Home and End keys for quick navigation
	case msg.String() == "home":
		// Jump to top
		m.ScrollOffset = 0
		return m, nil

	case msg.String() == "end":
		// Jump to bottom - we'll use a large value and let
		// the View method handle the actual maximum
		m.ScrollOffset = 9999
		return m, nil

	case key.Matches(msg, actionKeys.GoTo):
		if m.SelectedItem != nil {
			// Instead of just returning a cd command, directly change the directory
			// This requires different approaches depending on the OS
			if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
				// For macOS and Linux, we can use a special technique:
				// Write the cd command to a temporary file that the shell can source
				cdFile := filepath.Join(os.TempDir(), "mpm_cd_command")
				cdCommand := fmt.Sprintf("cd %s", m.SelectedItem.Path)
				err := os.WriteFile(cdFile, []byte(cdCommand), 0644)
				if err == nil {
					// Inform the user what to do
					fmt.Printf("\nTo change to the selected directory after exiting:\nsource %s\n", cdFile)
				}
			} else {
				// For other platforms, we'll still use the return value approach
				m.QuitCommand = fmt.Sprintf("cd %s", m.SelectedItem.Path)
			}
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.VSCode):
		if m.SelectedItem != nil {
			cmd := exec.Command("code", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Zed):
		if m.SelectedItem != nil {
			cmd := exec.Command("zed", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Cursor):
		if m.SelectedItem != nil {
			cmd := exec.Command("cursor", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Neovim):
		if m.SelectedItem != nil {
			// Return command to navigate and open nvim
			m.Quitting = true
			m.QuitCommand = fmt.Sprintf("cd %s && nvim .", m.SelectedItem.Path)
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Finder):
		if m.SelectedItem != nil {
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", m.SelectedItem.Path)
			case "windows":
				cmd = exec.Command("explorer", m.SelectedItem.Path)
			default: // Linux
				cmd = exec.Command("xdg-open", m.SelectedItem.Path)
			}
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Sublime):
		if m.SelectedItem != nil {
			cmd := exec.Command("subl", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Trae):
		if m.SelectedItem != nil {
			cmd := exec.Command("trae", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.TextMate):
		if m.SelectedItem != nil {
			cmd := exec.Command("mate", m.SelectedItem.Path)
			cmd.Start()
			m.Quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, actionKeys.Delete):
		if m.SelectedItem != nil {
			config.RemoveProject(m.SelectedItem.Name)

			// Reload projects
			refreshedConfig := config.LoadConfig()
			m.ProjectItems = []list.Item{}

			for _, p := range refreshedConfig.Projects {
				cat := p.Category
				if cat == "" {
					cat = "Uncategorized"
				}
				m.ProjectItems = append(m.ProjectItems, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})
			}

			// Apply current sort order
			if m.SortOrder == "asc" {
				sort.Slice(m.ProjectItems, func(i, j int) bool {
					return m.ProjectItems[i].(ProjectItem).Name < m.ProjectItems[j].(ProjectItem).Name
				})
			} else {
				sort.Slice(m.ProjectItems, func(i, j int) bool {
					return m.ProjectItems[i].(ProjectItem).Name > m.ProjectItems[j].(ProjectItem).Name
				})
			}

			m.List.SetItems(m.ProjectItems)
			m.ShowActions = false
		}
		return m, nil
	}

	return m, nil
}

// handleListView handles keyboard events in the list view
func handleListView(m ListModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if filtering is active before processing shortcuts
	if m.List.FilterState() == list.Filtering {
		// When filtering is active, only handle the Quit shortcut
		// and pass all other key events to the list
		if key.Matches(msg, m.Keys.Quit) {
			m.Quitting = true
			return m, tea.Quit
		}
	} else {
		// Handle main list view when not filtering
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Sort):
			// Only sort in projects view
			if m.ViewMode == "projects" {
				// Toggle sort order between asc and desc
				if m.SortOrder == "asc" {
					m.SortOrder = "desc"
					// Sort projects in descending order
					sort.Slice(m.ProjectItems, func(i, j int) bool {
						return m.ProjectItems[i].(ProjectItem).Name > m.ProjectItems[j].(ProjectItem).Name
					})
				} else {
					m.SortOrder = "asc"
					// Sort projects in ascending order
					sort.Slice(m.ProjectItems, func(i, j int) bool {
						return m.ProjectItems[i].(ProjectItem).Name < m.ProjectItems[j].(ProjectItem).Name
					})
				}

				// Apply sorting to the current view
				if m.SelectedCategory != "" {
					// We're in a filtered category view, reapply the filter with new sorting
					filteredProjects := []list.Item{}
					for _, item := range m.ProjectItems {
						project := item.(ProjectItem)
						if project.Category == m.SelectedCategory {
							filteredProjects = append(filteredProjects, project)
						}
					}
					m.List.SetItems(filteredProjects)
				} else {
					// We're in the main projects view
					m.List.SetItems(m.ProjectItems)
				}
			}
			return m, nil

		case key.Matches(msg, m.Keys.Up):
			// Handle up navigation directly
			if m.List.Index() > 0 {
				m.List.Select(m.List.Index() - 1)
			}
			return m, nil

		case key.Matches(msg, m.Keys.Down):
			// Handle down navigation directly
			if m.List.Index() < len(m.List.Items())-1 {
				m.List.Select(m.List.Index() + 1)
			}
			return m, nil

		case key.Matches(msg, m.Keys.ToggleView):
			// Toggle between projects and categories view
			if m.ViewMode == "projects" {
				// Switch to categories view
				m.ViewMode = "categories"
				m.SelectedCategory = "" // Reset selected category when switching to categories view
				m.List.SetItems(m.CategoryItems)
				m.List.Title = "(MPM) - Categories"
			} else {
				// Switch to projects view
				m.ViewMode = "projects"
				m.List.SetItems(m.ProjectItems)
				m.List.Title = "(MPM) - Projects"
			}
			return m, nil

		case key.Matches(msg, m.Keys.Select):
			if len(m.List.Items()) > 0 {
				// Handle selection based on current view mode
				if m.ViewMode == "categories" {
					// In category view, selecting a category shows projects in that category
					selected, ok := m.List.SelectedItem().(CategoryItem)
					if ok {
						m.SelectedCategory = selected.Name

						// Filter projects by the selected category
						filteredProjects := []list.Item{}
						for _, item := range m.ProjectItems {
							project := item.(ProjectItem)
							if project.Category == selected.Name {
								filteredProjects = append(filteredProjects, project)
							}
						}

						// Apply current sort order to filtered projects
						if m.SortOrder == "asc" {
							sort.Slice(filteredProjects, func(i, j int) bool {
								return filteredProjects[i].(ProjectItem).Name < filteredProjects[j].(ProjectItem).Name
							})
						} else {
							sort.Slice(filteredProjects, func(i, j int) bool {
								return filteredProjects[i].(ProjectItem).Name > filteredProjects[j].(ProjectItem).Name
							})
						}

						// Switch to projects view with filtered projects
						m.ViewMode = "projects"
						m.List.SetItems(filteredProjects)
						m.List.Title = fmt.Sprintf("(MPM) - %s", selected.Name)
					}
				} else {
					// In projects view, selecting a project shows actions
					selected, ok := m.List.SelectedItem().(ProjectItem)
					if ok {
						m.SelectedItem = &selected
						m.ShowActions = true
						m.ScrollOffset = 0 // Reset scroll position when entering details view

						// Scan the project directory for file chart and health status
						projectPath := m.SelectedItem.Path
						if _, err := os.Stat(projectPath); err == nil {
							// Scan directory with max depth of 3
							m.FileChart = fs.ScanDirectory(projectPath, 3, 0, "")
							// Count file types and assign colors
							m.FileTypeCounts = fs.CountFileTypes(m.FileChart)
							// Check Git status
							m.GitInfo = fs.CheckGitStatus(projectPath)
							// Perform project health scan just once and cache the result
							if !m.HealthScanned {
								m.HealthStatus = health.ScanProjectHealth(projectPath)
								m.HealthScanned = true
							}
						}
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.Keys.Add):
			// Only allow adding projects in projects view
			if m.ViewMode == "projects" {
				m.ShowForm = true
				for i := range m.FormInputs {
					if i == 0 {
						m.FormInputs[i].Focus()
					} else {
						m.FormInputs[i].Blur()
					}
				}
				m.FormFocused = 0
			}
			return m, nil

		case key.Matches(msg, m.Keys.Delete):
			// Only allow deleting projects in projects view
			if m.ViewMode == "projects" && len(m.List.Items()) > 0 {
				selected, ok := m.List.SelectedItem().(ProjectItem)
				if ok {
					config.RemoveProject(selected.Name)

					// Reload projects and rebuild category items
					refreshedConfig := config.LoadConfig()
					projectItems := []list.Item{}
					categoryMap := make(map[string]int)

					for _, p := range refreshedConfig.Projects {
						cat := p.Category
						if cat == "" {
							cat = "Uncategorized"
						}
						projectItems = append(projectItems, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})

						// Count projects per category
						categoryMap[cat]++
					}

					// Apply current sort order
					if m.SortOrder == "asc" {
						sort.Slice(projectItems, func(i, j int) bool {
							return projectItems[i].(ProjectItem).Name < projectItems[j].(ProjectItem).Name
						})
					} else {
						sort.Slice(projectItems, func(i, j int) bool {
							return projectItems[i].(ProjectItem).Name > projectItems[j].(ProjectItem).Name
						})
					}

					// Create category items
					categoryItems := []list.Item{}
					for cat, count := range categoryMap {
						categoryItems = append(categoryItems, CategoryItem{Name: cat, Count: count})
					}

					// Sort categories alphabetically
					sort.Slice(categoryItems, func(i, j int) bool {
						return categoryItems[i].(CategoryItem).Name < categoryItems[j].(CategoryItem).Name
					})

					// Update model with new items
					m.ProjectItems = projectItems
					m.CategoryItems = categoryItems
					m.List.SetItems(projectItems)
				}
			}
			return m, nil
		}
	}

	// If we get here, it means the key wasn't handled above
	// So we pass it to the list for default handling
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// View returns the main view for the application
func (m ListModel) View() string {
	if m.Quitting {
		return ""
	}

	if m.ShowForm {
		return m.FormView()
	}

	if m.ShowActions {
		// Get content for both sections
		actionView := m.ActionView()
		healthDashboard := m.RenderHealthDashboard()

		// Add a separator between sections
		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Render("─────────────────────────────────────────")

		// Combine all sections
		fullContent := lipgloss.JoinVertical(lipgloss.Left,
			actionView,
			separator,
			healthDashboard,
		)

		// Split the content into lines
		lines := strings.Split(fullContent, "\n")

		// Use the terminal height to determine visible area
		var visibleHeight int
		if m.WindowHeight > 0 {
			// Calculate visible area based on window height with some margins
			// Subtract some lines for UI chrome, indicators, etc.
			visibleHeight = m.WindowHeight - 6
		} else {
			// Fallback to a reasonable default if window height is not available
			visibleHeight = 30
		}

		// Apply scrolling offset, ensuring bounds checking
		startLine := m.ScrollOffset
		if startLine < 0 {
			startLine = 0
		}

		// Make sure we don't scroll past the content
		maxStartLine := len(lines) - visibleHeight
		if maxStartLine < 0 {
			maxStartLine = 0 // Content is shorter than visible area
		}

		if startLine > maxStartLine {
			startLine = maxStartLine
		}

		endLine := startLine + visibleHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}

		// Create the scrolled view
		visibleContent := strings.Join(lines[startLine:endLine], "\n")

		// Add scroll indicators if needed
		hasMoreAbove := startLine > 0
		hasMoreBelow := endLine < len(lines)

		// Width for scroll indicators - adjust based on window width
		indicatorWidth := 40
		if m.WindowWidth > 0 && m.WindowWidth < 80 {
			indicatorWidth = m.WindowWidth / 2
		}

		// Calculate scroll position indicator
		var scrollPositionInfo string
		if len(lines) > visibleHeight {
			// Calculate percentage
			percentScrolled := float64(startLine) / float64(maxStartLine) * 100
			scrollPositionInfo = fmt.Sprintf("(%d of %d lines, %.0f%%)",
				startLine+1, len(lines), percentScrolled)
		}

		// Create the final view with scroll indicators
		if hasMoreAbove {
			scrollUpIndicator := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Align(lipgloss.Center).
				Width(indicatorWidth).
				Render("▲ Scroll Up with ↑ ▲ " + scrollPositionInfo)
			visibleContent = scrollUpIndicator + "\n" + visibleContent
		}

		if hasMoreBelow {
			scrollDownIndicator := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Align(lipgloss.Center).
				Width(indicatorWidth).
				Render("▼ Scroll Down with ↓ ▼")
			visibleContent = visibleContent + "\n" + scrollDownIndicator
		}

		return visibleContent
	}

	return fmt.Sprintf("%s\n%s\n%s", m.HeaderView(), m.List.View(), m.FooterView())
}

// RunInteractive starts the interactive mode of the application
func RunInteractive() string {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())

	// Run the program
	model, err := p.Run()
	if err != nil {
		fmt.Printf("Error running interactive mode: %v\n", err)
		os.Exit(1)
	}

	// Check if we have a command to execute (like cd or nvim)
	if m, ok := model.(ListModel); ok && m.Quitting {
		// Return the command stored in the model
		return m.QuitCommand
	}

	return ""
}
