package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FileEntry represents a file in the file chart
type FileEntry struct {
	Path      string
	IsDir     bool
	Size      int64
	Extension string
}

// FileTypeCount tracks the count of file types
type FileTypeCount struct {
	Extension string
	Count     int
	Color     lipgloss.Color
}

// ShouldExclude determines if a file/directory should be excluded from scanning
func ShouldExclude(path string) bool {
	// Common directories to exclude
	excludeDirs := []string{
		"node_modules", ".git", ".idea", ".vscode", "__pycache__",
		"dist", "build", "target", "bin", "obj", ".next", ".nuxt",
		".DS_Store", "vendor", "coverage", ".gradle", ".mvn",
		".cache", ".npm", ".yarn", "venv", "env", ".env", ".pytest_cache",
	}

	base := filepath.Base(path)

	// Check if it's one of the excluded directories
	for _, dir := range excludeDirs {
		if strings.EqualFold(base, dir) {
			return true
		}
	}

	// Exclude hidden files/directories (start with .)
	if strings.HasPrefix(base, ".") {
		return true
	}

	// Exclude common binary files
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".o", ".obj", ".a", ".lib",
		".bin", ".dat", ".db", ".sqlite", ".class",
	}

	for _, ext := range binaryExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}

// ScanDirectory scans the directory and builds a file chart
func ScanDirectory(rootPath string, maxDepth int, currentDepth int, prefix string) []FileEntry {
	if maxDepth > 0 && currentDepth > maxDepth {
		return nil
	}

	var fileEntries []FileEntry

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		path := filepath.Join(rootPath, entry.Name())

		// Skip excluded files/directories
		if ShouldExclude(path) {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		// Get file extension
		ext := ""
		if !entry.IsDir() {
			ext = strings.ToLower(filepath.Ext(entry.Name()))
		}

		// Add this entry
		fileEntries = append(fileEntries, FileEntry{
			Path:      prefix + entry.Name(),
			IsDir:     entry.IsDir(),
			Size:      fileInfo.Size(),
			Extension: ext,
		})

		// Recursively scan subdirectories
		if entry.IsDir() {
			subEntries := ScanDirectory(
				path,
				maxDepth,
				currentDepth+1,
				prefix+entry.Name()+"/",
			)
			fileEntries = append(fileEntries, subEntries...)
		}
	}

	return fileEntries
}

// CountFileTypes counts the occurrences of each file type and assigns colors
func CountFileTypes(files []FileEntry) []FileTypeCount {
	extCounts := make(map[string]int)

	// Count occurrences of each extension
	for _, file := range files {
		if !file.IsDir && file.Extension != "" {
			extCounts[file.Extension]++
		}
	}

	// Prepare result with colors
	var result []FileTypeCount
	extColors := map[string]lipgloss.Color{
		".go":    "#00ADD8", // Go - Blue
		".js":    "#F7DF1E", // JavaScript - Yellow
		".ts":    "#3178C6", // TypeScript - Blue
		".jsx":   "#61DAFB", // React - Light Blue
		".tsx":   "#61DAFB", // React - Light Blue
		".py":    "#3776AB", // Python - Blue
		".java":  "#ED8B00", // Java - Orange
		".html":  "#E34F26", // HTML - Orange
		".css":   "#1572B6", // CSS - Blue
		".scss":  "#CD6799", // SCSS - Pink
		".json":  "#000000", // JSON - Black
		".yml":   "#CB171E", // YAML - Red
		".yaml":  "#CB171E", // YAML - Red
		".md":    "#083FA1", // Markdown - Blue
		".php":   "#777BB4", // PHP - Purple
		".rb":    "#CC342D", // Ruby - Red
		".c":     "#555555", // C - Gray
		".cpp":   "#004482", // C++ - Blue
		".cs":    "#239120", // C# - Green
		".rs":    "#DEA584", // Rust - Orange
		".swift": "#F05138", // Swift - Orange
		".kt":    "#A97BFF", // Kotlin - Purple
		".sh":    "#89E051", // Shell - Green
	}

	// Convert map to sorted slice
	for ext, count := range extCounts {
		color, ok := extColors[ext]
		if !ok {
			color = "#AAAAAA" // Default color for unknown extensions
		}

		result = append(result, FileTypeCount{
			Extension: ext,
			Count:     count,
			Color:     color,
		})
	}

	// Sort by count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	// Limit to top 10 extensions
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// RenderFileChart renders file chart and type statistics
func RenderFileChart(files []FileEntry, typeCounts []FileTypeCount) string {
	var b strings.Builder

	if len(files) == 0 {
		b.WriteString("\n  No files found in project directory.\n")
		return b.String()
	}

	// Get total file count (excluding directories)
	totalFiles := 0
	for _, file := range files {
		if !file.IsDir {
			totalFiles++
		}
	}

	// File types statistics section
	b.WriteString(fmt.Sprintf("\n  Top File Extensions (Total files: %d)\n", totalFiles))

	fileTypeStyle := lipgloss.NewStyle().Bold(true)
	for i, tc := range typeCounts {
		typeStyle := fileTypeStyle.Copy().Foreground(tc.Color)
		percentage := float64(tc.Count) / float64(totalFiles) * 100
		b.WriteString(fmt.Sprintf("  %2d. %s: %d (%.1f%%)\n",
			i+1,
			typeStyle.Render(tc.Extension),
			tc.Count,
			percentage))
	}

	return b.String()
}
