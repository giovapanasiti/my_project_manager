package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	configDir  string
	configFile string
)

// Styling
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#7D56F4"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	categoryStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C"))
	pathStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#B2B2B2"))
)

type Project struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

type Config struct {
	Projects []Project `json:"projects"`
}

// Project item for the list
type projectItem struct {
	name     string
	path     string
	category string
}

func (i projectItem) Title() string { return i.name }
func (i projectItem) Description() string {
	return fmt.Sprintf("%s %s", categoryStyle.Render("["+i.category+"]"), pathStyle.Render(i.path))
}
func (i projectItem) FilterValue() string { return i.name + " " + i.category + " " + i.path }

// Model for the main list view
type listModel struct {
	list         list.Model
	keys         listKeyMap
	selectedItem *projectItem
	showActions  bool
	showForm     bool
	formInputs   []textinput.Model
	formFocused  int
	quitting     bool
}

// Model for the action view
type actionKeyMap struct {
	GoTo     key.Binding
	VSCode   key.Binding
	Zed      key.Binding
	Cursor   key.Binding
	Neovim   key.Binding
	Finder   key.Binding
	Sublime  key.Binding
	IntelliJ key.Binding
	PyCharm  key.Binding
	Delete   key.Binding
	Back     key.Binding
}

// Key mappings for the list view
type listKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Add    key.Binding
	Delete key.Binding
	Filter key.Binding
	Quit   key.Binding
}

func newActionKeyMap() actionKeyMap {
	return actionKeyMap{
		GoTo: key.NewBinding(
			key.WithKeys("g", "enter"),
			key.WithHelp("g/enter", "go to directory"),
		),
		VSCode: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "open in VS Code"),
		),
		Zed: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "open in Zed"),
		),
		Cursor: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "open in Cursor"),
		),
		Neovim: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "open in Neovim"),
		),
		Finder: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "open in Finder/Explorer"),
		),
		Sublime: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open in Sublime Text"),
		),
		IntelliJ: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "open in IntelliJ IDEA"),
		),
		PyCharm: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "open in PyCharm"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete project"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back to list"),
		),
	}
}

func newListKeyMap() listKeyMap {
	return listKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add project"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete project"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	configDir = filepath.Join(home, ".mpm")
	configFile = filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println("Error creating config directory:", err)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		config := Config{Projects: []Project{}}
		saveConfig(config)
	}
}

func loadConfig() Config {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return Config{Projects: []Project{}}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Println("Error parsing config file:", err)
		return Config{Projects: []Project{}}
	}

	return config
}

func saveConfig(config Config) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error encoding config:", err)
		return
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		fmt.Println("Error writing config file:", err)
	}
}

func addProject(name, path, category string) {
	config := loadConfig()

	// Expand tilde to home directory if needed
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			os.Exit(1)
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		os.Exit(1)
	}

	// Check if the project already exists
	for i, p := range config.Projects {
		if p.Name == name {
			config.Projects[i] = Project{Name: name, Path: absPath, Category: category}
			saveConfig(config)
			fmt.Printf("Updated project '%s' with path '%s' and category '%s'\n", name, absPath, category)
			return
		}
	}

	// Add new project
	config.Projects = append(config.Projects, Project{Name: name, Path: absPath, Category: category})
	saveConfig(config)
	fmt.Printf("Added project '%s' with path '%s' and category '%s'\n", name, absPath, category)
}

func removeProject(name string) {
	config := loadConfig()
	found := false

	for i, p := range config.Projects {
		if p.Name == name {
			config.Projects = append(config.Projects[:i], config.Projects[i+1:]...)
			found = true
			break
		}
	}

	if found {
		saveConfig(config)
		fmt.Printf("Removed project '%s'\n", name)
	} else {
		fmt.Printf("Project '%s' not found\n", name)
	}
}

func listProjects() {
	config := loadConfig()

	if len(config.Projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	// Group projects by category
	categories := make(map[string][]Project)
	for _, p := range config.Projects {
		if p.Category == "" {
			p.Category = "Uncategorized"
		}
		categories[p.Category] = append(categories[p.Category], p)
	}

	// Sort categories
	var sortedCategories []string
	for c := range categories {
		sortedCategories = append(sortedCategories, c)
	}
	sort.Strings(sortedCategories)

	// Display projects by category
	for _, c := range sortedCategories {
		fmt.Printf("\n[%s]\n", c)
		projects := categories[c]
		for _, p := range projects {
			fmt.Printf("  - %s: %s\n", p.Name, p.Path)
		}
	}
}

func goToProject(name string) {
	config := loadConfig()

	for _, p := range config.Projects {
		if p.Name == name {
			// We need to print the command to change directory
			// The shell wrapper will execute it
			fmt.Printf("cd %s\n", p.Path)
			return
		}
	}

	fmt.Printf("Project '%s' not found\n", name)
}

// Initialize the form for adding new projects
func initForm() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	for i := range inputs {
		input := textinput.New()
		input.CharLimit = 150
		input.CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

		switch i {
		case 0:
			input.Placeholder = "Project Name"
			input.Focus()
		case 1:
			input.Placeholder = "Project Path (e.g., ~/projects/myapp)"
		case 2:
			input.Placeholder = "Category (optional)"
		}

		inputs[i] = input
	}

	return inputs
}

// Initialize the list model for the main view
func initialModel() listModel {
	// Load projects
	config := loadConfig()
	items := []list.Item{}

	for _, p := range config.Projects {
		cat := p.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		items = append(items, projectItem{name: p.Name, path: p.Path, category: cat})
	}

	// Set up list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "My Project Manager (MPM)"
	l.Styles.Title = titleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA07A"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Set up key mappings
	keys := newListKeyMap()

	// Return the model
	return listModel{
		list:        l,
		keys:        keys,
		showActions: false,
		showForm:    false,
		formInputs:  initForm(),
		formFocused: 0,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

// Update function for the main application
func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showForm {
			// Handle form view
			switch msg.String() {
			case "esc":
				m.showForm = false
				return m, nil

			case "tab", "shift+tab":
				// Handle tab navigation in the form
				if msg.String() == "tab" {
					m.formFocused = (m.formFocused + 1) % len(m.formInputs)
				} else {
					m.formFocused = (m.formFocused - 1 + len(m.formInputs)) % len(m.formInputs)
				}

				for i := range m.formInputs {
					if i == m.formFocused {
						m.formInputs[i].Focus()
					} else {
						m.formInputs[i].Blur()
					}
				}

				return m, nil

			case "enter":
				if m.formFocused == len(m.formInputs)-1 {
					// Submit form on Enter when on last field
					name := strings.TrimSpace(m.formInputs[0].Value())
					path := strings.TrimSpace(m.formInputs[1].Value())
					category := strings.TrimSpace(m.formInputs[2].Value())

					if name != "" && path != "" {
						addProject(name, path, category)

						// Reload projects for the list
						config := loadConfig()
						items := []list.Item{}

						for _, p := range config.Projects {
							cat := p.Category
							if cat == "" {
								cat = "Uncategorized"
							}
							items = append(items, projectItem{name: p.Name, path: p.Path, category: cat})
						}

						m.list.SetItems(items)

						// Reset form
						for i := range m.formInputs {
							m.formInputs[i].SetValue("")
						}
						m.formFocused = 0
						m.formInputs[0].Focus()
						for i := 1; i < len(m.formInputs); i++ {
							m.formInputs[i].Blur()
						}

						m.showForm = false
					}

					return m, nil
				} else {
					// Move to next field on Enter
					m.formFocused = (m.formFocused + 1) % len(m.formInputs)
					for i := range m.formInputs {
						if i == m.formFocused {
							m.formInputs[i].Focus()
						} else {
							m.formInputs[i].Blur()
						}
					}
					return m, nil
				}
			}

			// Handle text input
			m.formInputs[m.formFocused], cmd = m.formInputs[m.formFocused].Update(msg)
			return m, cmd

		} else if m.showActions {
			// Handle action view
			actionKeys := newActionKeyMap()

			switch {
			case key.Matches(msg, actionKeys.Back):
				m.showActions = false
				return m, nil

			case key.Matches(msg, actionKeys.GoTo):
				if m.selectedItem != nil {
					// Return command to change directory for shell wrapper
					m.quitting = true
					return m, tea.Sequence(
						tea.Quit,
						func() tea.Msg { return quitMsg(fmt.Sprintf("cd %s", m.selectedItem.path)) },
					)
				}

			case key.Matches(msg, actionKeys.VSCode):
				if m.selectedItem != nil {
					cmd := exec.Command("code", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.Zed):
				if m.selectedItem != nil {
					cmd := exec.Command("zed", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.Cursor):
				if m.selectedItem != nil {
					cmd := exec.Command("cursor", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.Neovim):
				if m.selectedItem != nil {
					// Return command to navigate and open nvim
					m.quitting = true
					return m, tea.Sequence(
						tea.Quit,
						func() tea.Msg { return quitMsg(fmt.Sprintf("cd %s && nvim .", m.selectedItem.path)) },
					)
				}

			case key.Matches(msg, actionKeys.Finder):
				if m.selectedItem != nil {
					var cmd *exec.Cmd
					switch runtime.GOOS {
					case "darwin":
						cmd = exec.Command("open", m.selectedItem.path)
					case "windows":
						cmd = exec.Command("explorer", m.selectedItem.path)
					default: // Linux
						cmd = exec.Command("xdg-open", m.selectedItem.path)
					}
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.Sublime):
				if m.selectedItem != nil {
					cmd := exec.Command("subl", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.IntelliJ):
				if m.selectedItem != nil {
					cmd := exec.Command("idea", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.PyCharm):
				if m.selectedItem != nil {
					cmd := exec.Command("pycharm", m.selectedItem.path)
					cmd.Start()
					m.quitting = true
					return m, tea.Quit
				}

			case key.Matches(msg, actionKeys.Delete):
				if m.selectedItem != nil {
					removeProject(m.selectedItem.name)

					// Reload projects
					config := loadConfig()
					items := []list.Item{}

					for _, p := range config.Projects {
						cat := p.Category
						if cat == "" {
							cat = "Uncategorized"
						}
						items = append(items, projectItem{name: p.Name, path: p.Path, category: cat})
					}

					m.list.SetItems(items)
					m.showActions = false
				}
				return m, nil
			}

		} else {
			// Handle main list view
			switch {
			case key.Matches(msg, m.keys.Quit):
				m.quitting = true
				return m, tea.Quit

			case key.Matches(msg, m.keys.Select):
				if len(m.list.Items()) > 0 {
					selected, ok := m.list.SelectedItem().(projectItem)
					if ok {
						m.selectedItem = &selected
						m.showActions = true
					}
				}
				return m, nil

			case key.Matches(msg, m.keys.Add):
				m.showForm = true
				for i := range m.formInputs {
					if i == 0 {
						m.formInputs[i].Focus()
					} else {
						m.formInputs[i].Blur()
					}
				}
				m.formFocused = 0
				return m, nil

			case key.Matches(msg, m.keys.Delete):
				if len(m.list.Items()) > 0 {
					selected, ok := m.list.SelectedItem().(projectItem)
					if ok {
						removeProject(selected.name)

						// Reload projects
						config := loadConfig()
						items := []list.Item{}

						for _, p := range config.Projects {
							cat := p.Category
							if cat == "" {
								cat = "Uncategorized"
							}
							items = append(items, projectItem{name: p.Name, path: p.Path, category: cat})
						}

						m.list.SetItems(items)
					}
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-verticalMarginHeight)
	}

	// Handle list updates by default
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) headerView() string {
	return titleStyle.Render(" My Project Manager (MPM) ")
}

func (m listModel) footerView() string {
	info := helpStyle.Render("Press 'q' to quit, 'a' to add, '/' to filter")
	return info
}

func (m listModel) formView() string {
	var b strings.Builder

	b.WriteString("\n  Add New Project\n\n")

	for i := range m.formInputs {
		b.WriteString(m.formInputs[i].View())
		b.WriteString("\n")
	}

	b.WriteString("\nPress Enter to submit each field • ESC to cancel\n")

	return b.String()
}

func (m listModel) actionView() string {
	if m.selectedItem == nil {
		return ""
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(m.selectedItem.name)
	category := lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C")).Render("[" + m.selectedItem.category + "]")
	path := lipgloss.NewStyle().Foreground(lipgloss.Color("#B2B2B2")).Render(m.selectedItem.path)

	b.WriteString(fmt.Sprintf("\n  %s %s\n  %s\n\n", title, category, path))
	b.WriteString("  [g] Go to directory\n")
	b.WriteString("  [v] Open in VS Code\n")
	b.WriteString("  [z] Open in Zed\n")
	b.WriteString("  [c] Open in Cursor\n")
	b.WriteString("  [n] Open in Neovim\n")
	b.WriteString("  [f] Open in Finder/File Explorer\n")
	b.WriteString("  [s] Open in Sublime Text\n")
	b.WriteString("  [i] Open in IntelliJ IDEA\n")
	b.WriteString("  [p] Open in PyCharm\n")
	b.WriteString("  [d] Delete project\n")
	b.WriteString("\n  [ESC/q] Back to list\n")

	return b.String()
}

// Main view function
func (m listModel) View() string {
	if m.quitting {
		return ""
	}

	if m.showForm {
		return m.formView()
	}

	if m.showActions {
		return m.actionView()
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.list.View(), m.footerView())
}

// Custom message type to hold command output
type quitMsg string

// Run the interactive mode
func interactiveMode() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	// Run the program
	model, err := p.Run()
	if err != nil {
		fmt.Printf("Error running interactive mode: %v\n", err)
		os.Exit(1)
	}

	// Check if we have a command to execute (like cd or nvim)
	if m, ok := model.(listModel); ok && m.quitting {
		// Access any command messages that were stored in the model itself
		return
	}
}

func main() {
	initConfig()

	var rootCmd = &cobra.Command{
		Use:   "mpm",
		Short: "My Project Manager - A CLI tool to manage your projects",
		Long: `My Project Manager (mpm) is a CLI tool that helps you manage your projects
by storing their locations on the filesystem and providing quick navigation.`,
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new project",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if -w flag is present
			useWorkingDir, _ := cmd.Flags().GetBool("working-dir")

			var name, path string
			var category string

			if useWorkingDir {
				// Get current working directory
				currentDir, err := os.Getwd()
				if err != nil {
					fmt.Println("Error getting current directory:", err)
					return
				}

				// Extract folder name from path
				folderName := filepath.Base(currentDir)

				name = folderName
				path = currentDir
				category, _ = cmd.Flags().GetString("category")
			} else {
				// Use the provided flags
				name, _ = cmd.Flags().GetString("name")
				path, _ = cmd.Flags().GetString("path")
				category, _ = cmd.Flags().GetString("category")

				// Validate input
				if name == "" || path == "" {
					fmt.Println("Error: Both name and path are required")
					return
				}
			}

			addProject(name, path, category)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			removeProject(args[0])
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Run: func(cmd *cobra.Command, args []string) {
			listProjects()
		},
	}

	var goCmd = &cobra.Command{
		Use:   "go",
		Short: "Navigate to a project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			goToProject(args[0])
		},
	}

	var interactiveCmd = &cobra.Command{
		Use:   "i",
		Short: "Start interactive mode",
		Run: func(cmd *cobra.Command, args []string) {
			interactiveMode()
		},
	}

	addCmd.Flags().StringP("name", "n", "", "Project name")
	addCmd.Flags().StringP("path", "p", "", "Project path")
	addCmd.Flags().StringP("category", "c", "", "Project category (optional)")
	addCmd.Flags().BoolP("working-dir", "w", false, "Use current directory as project path and folder name as project name")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(goCmd)
	rootCmd.AddCommand(interactiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
