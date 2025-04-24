package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"mpm/pkg/config"
	"mpm/pkg/fs"
)

// ProjectItem represents a project in the UI list
type ProjectItem struct {
	Name     string
	Path     string
	Category string
}

// Title implements list.Item interface
func (i ProjectItem) Title() string { return i.Name }

// Description implements list.Item interface
func (i ProjectItem) Description() string {
	return fmt.Sprintf("%s %s", CategoryStyle.Render("["+i.Category+"]"), PathStyle.Render(i.Path))
}

// FilterValue implements list.Item interface
func (i ProjectItem) FilterValue() string { return i.Name + " " + i.Category + " " + i.Path }

// CategoryItem represents a category in the category view
type CategoryItem struct {
	Name  string
	Count int
}

// Title implements list.Item interface
func (i CategoryItem) Title() string { return i.Name }

// Description implements list.Item interface
func (i CategoryItem) Description() string {
	return fmt.Sprintf("%d projects", i.Count)
}

// FilterValue implements list.Item interface
func (i CategoryItem) FilterValue() string { return i.Name }

// ListModel is the main model for the list view
type ListModel struct {
	List             list.Model
	Keys             ListKeyMap
	SelectedItem     *ProjectItem
	ShowActions      bool
	ShowForm         bool
	FormInputs       []textinput.Model
	FormFocused      int
	Quitting         bool
	FileChart        []fs.FileEntry
	FileTypeCounts   []fs.FileTypeCount
	GitInfo          fs.GitInfo
	QuitCommand      string      // To store the command to execute after quitting
	ViewMode         string      // "projects" or "categories"
	SelectedCategory string      // Currently selected category in category view
	CategoryItems    []list.Item // Store category items for category view
	ProjectItems     []list.Item // Store project items for project view
	SortOrder        string      // "asc" or "desc" for project sorting
}

// ListKeyMap defines key bindings for the list view
type ListKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Select     key.Binding
	Add        key.Binding
	Delete     key.Binding
	Filter     key.Binding
	Quit       key.Binding
	ToggleView key.Binding
	Sort       key.Binding
}

// ActionKeyMap defines key bindings for the action view
type ActionKeyMap struct {
	GoTo     key.Binding
	VSCode   key.Binding
	Zed      key.Binding
	Cursor   key.Binding
	Neovim   key.Binding
	Finder   key.Binding
	Sublime  key.Binding
	Trae     key.Binding
	TextMate key.Binding
	Delete   key.Binding
	Back     key.Binding
}

// NewActionKeyMap creates a new action key map with default bindings
func NewActionKeyMap() ActionKeyMap {
	return ActionKeyMap{
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
		Trae: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "open in Trae"),
		),
		TextMate: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "open in TextMate"),
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

// NewListKeyMap creates a new list key map with default bindings
func NewListKeyMap() ListKeyMap {
	return ListKeyMap{
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
		ToggleView: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle view"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle sort order"),
		),
	}
}

// InitForm initializes the form for adding new projects
func InitForm() []textinput.Model {
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

// InitialModel initializes the list model for the main view
func InitialModel() ListModel {
	// Load projects
	config := config.LoadConfig()
	projectItems := []list.Item{}
	categoryMap := make(map[string]int)

	for _, p := range config.Projects {
		cat := p.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		projectItems = append(projectItems, ProjectItem{Name: p.Name, Path: p.Path, Category: cat})

		// Count projects per category
		categoryMap[cat]++
	}

	// Sort projects alphabetically by default (ascending)
	sort.Slice(projectItems, func(i, j int) bool {
		return projectItems[i].(ProjectItem).Name < projectItems[j].(ProjectItem).Name
	})

	// Create category items
	categoryItems := []list.Item{}
	for cat, count := range categoryMap {
		categoryItems = append(categoryItems, CategoryItem{Name: cat, Count: count})
	}

	// Sort categories alphabetically
	sort.Slice(categoryItems, func(i, j int) bool {
		return categoryItems[i].(CategoryItem).Name < categoryItems[j].(CategoryItem).Name
	})

	// Set up list with project items initially
	l := list.New(projectItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "My Project Manager (MPM) - Projects"
	l.Styles.Title = TitleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA07A"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Set up key mappings
	keys := NewListKeyMap()

	// Return the model
	return ListModel{
		List:           l,
		Keys:           keys,
		ShowActions:    false,
		ShowForm:       false,
		FormInputs:     InitForm(),
		FormFocused:    0,
		FileChart:      []fs.FileEntry{},
		FileTypeCounts: []fs.FileTypeCount{},
		ViewMode:       "projects",
		ProjectItems:   projectItems,
		CategoryItems:  categoryItems,
		SortOrder:      "asc", // Default sort order is ascending
	}
}

// HeaderView returns the header view
func (m ListModel) HeaderView() string {
	return TitleStyle.Render(" My Project Manager (MPM) ")
}

// FooterView returns the footer view
func (m ListModel) FooterView() string {
	var info string
	if m.ViewMode == "categories" {
		info = HelpStyle.Render("Press 'q' to quit, 'tab' to switch to projects view, '/' to filter, 'enter' to view projects in category")
	} else {
		// Show sort order in projects view
		sortStatus := ""
		sortStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		if m.SortOrder == "asc" {
			sortStatus = sortStyle.Render("[A→Z]")
		} else {
			sortStatus = sortStyle.Render("[Z→A]")
		}

		// Highlight the sort key
		sortInstruction := "'s' to sort"
		highlightedSort := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("'s'")
		sortInstruction = highlightedSort + " to sort"

		info = HelpStyle.Render(fmt.Sprintf("Press 'q' to quit, 'a' to add, %s %s, '/' to filter, 'tab' to switch to categories view",
			sortInstruction, sortStatus))
	}
	return info
}

// FormView returns the form view for adding new projects
func (m ListModel) FormView() string {
	var b strings.Builder

	b.WriteString("\n  Add New Project\n\n")

	for i := range m.FormInputs {
		b.WriteString(m.FormInputs[i].View())
		b.WriteString("\n")
	}

	b.WriteString("\nPress Enter to submit each field • ESC to cancel\n")

	return b.String()
}

// ActionView returns the action view for the selected project
func (m ListModel) ActionView() string {
	if m.SelectedItem == nil {
		return ""
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(m.SelectedItem.Name)
	category := lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C")).Render("[" + m.SelectedItem.Category + "]")
	path := lipgloss.NewStyle().Foreground(lipgloss.Color("#B2B2B2")).Render(m.SelectedItem.Path)

	b.WriteString(fmt.Sprintf("\n  %s %s\n  %s\n\n", title, category, path))

	// Add Git information
	b.WriteString(fs.RenderGitInfo(m.GitInfo))
	b.WriteString("\n")

	// Add file type statistics if we have data
	if len(m.FileChart) > 0 && len(m.FileTypeCounts) > 0 {
		b.WriteString(fs.RenderFileChart(m.FileChart, m.FileTypeCounts))
		b.WriteString("\n")
	}

	b.WriteString("  [g] Go to directory\n")
	b.WriteString("  [v] Open in VS Code\n")
	b.WriteString("  [z] Open in Zed\n")
	b.WriteString("  [c] Open in Cursor\n")
	b.WriteString("  [n] Open in Neovim\n")
	b.WriteString("  [f] Open in Finder/File Explorer\n")
	b.WriteString("  [s] Open in Sublime Text\n")
	b.WriteString("  [t] Open in Trae\n")
	b.WriteString("  [m] Open in TextMate\n")
	b.WriteString("  [d] Delete project\n")
	b.WriteString("\n  [ESC/q] Back to list\n")

	return b.String()
}
