package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"mpm/pkg/config"
	"mpm/pkg/ui"
)

// InitApp initializes the command line interface
func InitApp() {
	config.InitConfig()

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

			config.AddProject(name, path, category)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config.RemoveProject(args[0])
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Run: func(cmd *cobra.Command, args []string) {
			ListProjects()
		},
	}

	var goCmd = &cobra.Command{
		Use:   "go",
		Short: "Navigate to a project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			GoToProject(args[0])
		},
	}

	var interactiveCmd = &cobra.Command{
		Use:   "i",
		Short: "Start interactive mode",
		Run: func(cmd *cobra.Command, args []string) {
			result := ui.RunInteractive()
			if result != "" {
				fmt.Println(result)
			}
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

// ListProjects displays all projects by category
func ListProjects() {
	cfg := config.LoadConfig()

	if len(cfg.Projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	// Group projects by category
	categories := make(map[string][]config.Project)
	for _, p := range cfg.Projects {
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

// GoToProject prints the path for a project to be used by a shell wrapper
func GoToProject(name string) {
	path := config.GoToProject(name)
	if path != "" {
		// We need to print the command to change directory
		// The shell wrapper will execute it
		fmt.Printf("cd %s\n", path)
		return
	}

	fmt.Printf("Project '%s' not found\n", name)
}
