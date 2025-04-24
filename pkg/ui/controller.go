package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mpm/pkg/config"
	"mpm/pkg/fs"
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
				items := []list.Item{}

				for _, p := range refreshedConfig.Projects {
					cat := p.Category
					if cat == "" {
						cat = "Uncategorized"
					}
					items = append(items, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})
				}

				m.List.SetItems(items)

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
		return m, nil

	case key.Matches(msg, actionKeys.GoTo):
		if m.SelectedItem != nil {
			// Return command to change directory for shell wrapper
			m.Quitting = true
			m.QuitCommand = fmt.Sprintf("cd %s", m.SelectedItem.Path)
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
			items := []list.Item{}

			for _, p := range refreshedConfig.Projects {
				cat := p.Category
				if cat == "" {
					cat = "Uncategorized"
				}
				items = append(items, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})
			}

			m.List.SetItems(items)
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

		case key.Matches(msg, m.Keys.Select):
			if len(m.List.Items()) > 0 {
				selected, ok := m.List.SelectedItem().(ProjectItem)
				if ok {
					m.SelectedItem = &selected
					m.ShowActions = true

					// Scan the project directory for file chart
					projectPath := m.SelectedItem.Path
					if _, err := os.Stat(projectPath); err == nil {
						// Scan directory with max depth of 3
						m.FileChart = fs.ScanDirectory(projectPath, 3, 0, "")
						// Count file types and assign colors
						m.FileTypeCounts = fs.CountFileTypes(m.FileChart)
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.Keys.Add):
			m.ShowForm = true
			for i := range m.FormInputs {
				if i == 0 {
					m.FormInputs[i].Focus()
				} else {
					m.FormInputs[i].Blur()
				}
			}
			m.FormFocused = 0
			return m, nil

		case key.Matches(msg, m.Keys.Delete):
			if len(m.List.Items()) > 0 {
				selected, ok := m.List.SelectedItem().(ProjectItem)
				if ok {
					config.RemoveProject(selected.Name)

					// Reload projects
					refreshedConfig := config.LoadConfig()
					items := []list.Item{}

					for _, p := range refreshedConfig.Projects {
						cat := p.Category
						if cat == "" {
							cat = "Uncategorized"
						}
						items = append(items, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})
					}

					m.List.SetItems(items)
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
		return m.ActionView()
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
