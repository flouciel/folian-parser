package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flouciel/folian-parser/internal/epub"
	"github.com/flouciel/folian-parser/internal/restructure"
)

// Version information
const (
	Version = "0.2.4"
	GitHubRepo = "flouciel/folian-parser"
)

// checkLatestVersion checks the latest version from GitHub releases
func checkLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to check latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to check latest version: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract version from response
	version := strings.TrimPrefix(string(body), "{\"tag_name\":\"v")
	version = strings.Split(version, "\"")[0]
	return version, nil
}

// compareVersions compares two version strings
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		if v1Parts[i] > v2Parts[i] {
			return 1
		}
		if v1Parts[i] < v2Parts[i] {
			return -1
		}
	}

	if len(v1Parts) > len(v2Parts) {
		return 1
	}
	if len(v1Parts) < len(v2Parts) {
		return -1
	}
	return 0
}

// updateToLatestVersion updates the tool to the latest version
func updateToLatestVersion() error {
	cmd := exec.Command("go", "install", fmt.Sprintf("github.com/%s@latest", GitHubRepo))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	return nil
}

// ensureFormatDirectory ensures that the format directory exists and contains all necessary files
func ensureFormatDirectory(formatDir string) error {
	// Check if the format directory exists
	if _, err := os.Stat(formatDir); os.IsNotExist(err) {
		// Create the format directory
		if err := os.MkdirAll(formatDir, 0755); err != nil {
			return fmt.Errorf("failed to create format directory: %w", err)
		}

		// Run the create-format-dir.sh script to populate the directory
		cmd := exec.Command("./create-format-dir.sh", formatDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run create-format-dir.sh: %w", err)
		}
	} else {
		// Check if the required files exist
		requiredFiles := []string{
			"stylesheet.css",
			"titlepage.xhtml",
			"jacket.xhtml",
			"nav.xhtml",
			"jura.ttf",
			"folian.png",
		}

		for _, file := range requiredFiles {
			filePath := filepath.Join(formatDir, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// If any required file is missing, run the create-format-dir.sh script
				cmd := exec.Command("./create-format-dir.sh", formatDir)
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to run create-format-dir.sh: %w", err)
				}
				break
			}
		}
	}

	return nil
}

func main() {
	// Parse command-line arguments
	inputPath := flag.String("i", "", "Input EPUB file path")
	outputPath := flag.String("o", "", "Output EPUB file path")
	formatDir := flag.String("f", "format", "Path to the format directory containing templates and assets")
	versionFlag := flag.Bool("v", false, "Display version information")
	debugFlag := flag.Bool("d", false, "Enable debug output")
	updateFlag := flag.Bool("u", false, "Check for updates and update if a newer version is available")
	flag.Parse()

	// Handle update check
	if *updateFlag {
		latestVersion, err := checkLatestVersion()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			os.Exit(1)
		}

		if compareVersions(latestVersion, Version) > 0 {
			fmt.Printf("A new version is available: %s (current: %s)\n", latestVersion, Version)
			fmt.Println("Updating to the latest version...")
			if err := updateToLatestVersion(); err != nil {
				fmt.Printf("Error updating: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Update completed successfully!")
			os.Exit(0)
		} else {
			fmt.Printf("You are running the latest version (%s)\n", Version)
			os.Exit(0)
		}
	}

	// Display version information if requested
	if *versionFlag {
		fmt.Printf("Folian Parser version %s\n", Version)
		os.Exit(0)
	}

	// Set the format directory path
	restructure.FormatDirPath = *formatDir

	// Set debug mode
	restructure.DebugMode = *debugFlag

	// Ensure the format directory exists and contains all necessary files
	if err := ensureFormatDirectory(*formatDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Validate input path
	if *inputPath == "" {
		fmt.Println("Error: Input file path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(*inputPath); os.IsNotExist(err) {
		fmt.Printf("Error: Input file does not exist: %s\n", *inputPath)
		os.Exit(1)
	}

	// Generate output path if not provided
	if *outputPath == "" {
		ext := filepath.Ext(*inputPath)
		base := filepath.Base(*inputPath)
		dir := filepath.Dir(*inputPath)
		*outputPath = filepath.Join(dir, base[:len(base)-len(ext)]+"-fixed"+ext)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputPath)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error: Failed to create output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Create a processor
	processor := epub.NewProcessor()

	// Process the EPUB file
	if err := processor.Process(*inputPath, *outputPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("EPUB file successfully restructured: %s\n", *outputPath)
}
