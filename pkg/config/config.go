package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	configDir  string
	configFile string
)

// Project represents a managed project in the application
type Project struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

// Config holds the application configuration
type Config struct {
	Projects []Project `json:"projects"`
}

// InitConfig initializes the config directory and file
func InitConfig() {
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
		SaveConfig(config)
	}
}

// LoadConfig loads the configuration from the file
func LoadConfig() Config {
	data, err := os.ReadFile(configFile)
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

// SaveConfig saves the configuration to the file
func SaveConfig(config Config) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error encoding config:", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		fmt.Println("Error writing config file:", err)
	}
}

// AddProject adds or updates a project in the configuration
func AddProject(name, path, category string) {
	config := LoadConfig()

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
			SaveConfig(config)
			fmt.Printf("Updated project '%s' with path '%s' and category '%s'\n", name, absPath, category)
			return
		}
	}

	// Add new project
	config.Projects = append(config.Projects, Project{Name: name, Path: absPath, Category: category})
	SaveConfig(config)
	fmt.Printf("Added project '%s' with path '%s' and category '%s'\n", name, absPath, category)
}

// RemoveProject removes a project from the configuration
func RemoveProject(name string) {
	config := LoadConfig()
	found := false

	for i, p := range config.Projects {
		if p.Name == name {
			config.Projects = append(config.Projects[:i], config.Projects[i+1:]...)
			found = true
			break
		}
	}

	if found {
		SaveConfig(config)
		fmt.Printf("Removed project '%s'\n", name)
	} else {
		fmt.Printf("Project '%s' not found\n", name)
	}
}

// GoToProject returns the path of a project for navigation
func GoToProject(name string) string {
	config := LoadConfig()

	for _, p := range config.Projects {
		if p.Name == name {
			return p.Path
		}
	}

	return ""
}
