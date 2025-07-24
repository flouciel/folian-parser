package main

import (
	"archive/zip"
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
	Version = "0.3"
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

// validateEPUB validates the structure and integrity of an EPUB file
func validateEPUB(epubPath string) error {
	fmt.Printf("🔍 Validating EPUB: %s\n", epubPath)

	// Check if file exists
	if _, err := os.Stat(epubPath); os.IsNotExist(err) {
		return fmt.Errorf("EPUB file not found: %s", epubPath)
	}

	// Check if it's a valid ZIP file
	reader, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("invalid EPUB file (not a valid ZIP): %w", err)
	}
	defer reader.Close()

	// Check for required files
	var hasMimetype, hasContainer, hasOPF bool

	for _, file := range reader.File {
		switch file.Name {
		case "mimetype":
			hasMimetype = true
		case "META-INF/container.xml":
			hasContainer = true
		}
		if strings.HasSuffix(file.Name, ".opf") {
			hasOPF = true
		}
	}

	if !hasMimetype {
		fmt.Println("⚠️  Warning: Missing mimetype file")
	}
	if !hasContainer {
		return fmt.Errorf("missing required META-INF/container.xml")
	}
	if !hasOPF {
		return fmt.Errorf("missing required OPF file")
	}

	fmt.Println("✅ EPUB validation passed")
	return nil
}

// analyzeEPUB analyzes the structure and content of an EPUB file
func analyzeEPUB(epubPath string) error {
	fmt.Printf("📊 Analyzing EPUB structure: %s\n", epubPath)

	reader, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer reader.Close()

	var contentFiles, imageFiles, cssFiles, fontFiles int
	var totalSize int64

	for _, file := range reader.File {
		totalSize += int64(file.UncompressedSize64)

		ext := strings.ToLower(filepath.Ext(file.Name))
		switch {
		case ext == ".html" || ext == ".xhtml":
			if !strings.Contains(file.Name, "nav") &&
			   !strings.Contains(file.Name, "toc") &&
			   !strings.Contains(file.Name, "title") &&
			   !strings.Contains(file.Name, "cover") {
				contentFiles++
			}
		case ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp":
			imageFiles++
		case ext == ".css":
			cssFiles++
		case ext == ".ttf" || ext == ".otf" || ext == ".woff" || ext == ".woff2":
			fontFiles++
		}
	}

	fmt.Printf("📄 Content files: %d\n", contentFiles)
	fmt.Printf("🖼️  Images: %d\n", imageFiles)
	fmt.Printf("🎨 CSS files: %d\n", cssFiles)
	fmt.Printf("🔤 Fonts: %d\n", fontFiles)
	fmt.Printf("📦 Total size: %.2f MB\n", float64(totalSize)/(1024*1024))

	// Provide recommendations
	if contentFiles > 50 {
		fmt.Printf("💡 Recommendation: %d content files detected. Enhanced processing will consolidate these into meaningful chapters.\n", contentFiles)
	}
	if cssFiles > 3 {
		fmt.Printf("💡 Recommendation: %d CSS files detected. Processing will consolidate these into a single optimized stylesheet.\n", cssFiles)
	}
	if fontFiles == 0 {
		fmt.Println("💡 Recommendation: No fonts detected. Processing will add the Jura font for consistent typography.")
	}

	return nil
}

// compareEPUBs compares two EPUB files and shows the differences
func compareEPUBs(epub1Path, epub2Path string) error {
	fmt.Printf("📊 Comparing EPUBs:\n")
	fmt.Printf("   📖 Original: %s\n", epub1Path)
	fmt.Printf("   ✨ Enhanced: %s\n", epub2Path)
	fmt.Println()

	// Analyze both files
	stats1, err := getEPUBStats(epub1Path)
	if err != nil {
		return fmt.Errorf("failed to analyze %s: %w", epub1Path, err)
	}

	stats2, err := getEPUBStats(epub2Path)
	if err != nil {
		return fmt.Errorf("failed to analyze %s: %w", epub2Path, err)
	}

	// Display comparison
	fmt.Printf("📄 Content Files:  %d → %d", stats1.ContentFiles, stats2.ContentFiles)
	if stats2.ContentFiles < stats1.ContentFiles {
		fmt.Printf(" (📉 %d fewer)", stats1.ContentFiles-stats2.ContentFiles)
	}
	fmt.Println()

	fmt.Printf("🖼️  Images:        %d → %d", stats1.ImageFiles, stats2.ImageFiles)
	if stats2.ImageFiles > stats1.ImageFiles {
		fmt.Printf(" (📈 %d added)", stats2.ImageFiles-stats1.ImageFiles)
	}
	fmt.Println()

	fmt.Printf("🎨 CSS Files:     %d → %d", stats1.CSSFiles, stats2.CSSFiles)
	if stats2.CSSFiles < stats1.CSSFiles {
		fmt.Printf(" (📉 %d consolidated)", stats1.CSSFiles-stats2.CSSFiles)
	}
	fmt.Println()

	fmt.Printf("🔤 Fonts:         %d → %d", stats1.FontFiles, stats2.FontFiles)
	if stats2.FontFiles > stats1.FontFiles {
		fmt.Printf(" (📈 %d added)", stats2.FontFiles-stats1.FontFiles)
	}
	fmt.Println()

	fmt.Printf("📦 Size:          %.2f MB → %.2f MB",
		float64(stats1.TotalSize)/(1024*1024),
		float64(stats2.TotalSize)/(1024*1024))
	sizeDiff := float64(stats2.TotalSize-stats1.TotalSize) / (1024 * 1024)
	if sizeDiff > 0 {
		fmt.Printf(" (📈 +%.2f MB)", sizeDiff)
	} else if sizeDiff < 0 {
		fmt.Printf(" (📉 %.2f MB)", sizeDiff)
	}
	fmt.Println()

	return nil
}

// EPUBStats holds statistics about an EPUB file
type EPUBStats struct {
	ContentFiles int
	ImageFiles   int
	CSSFiles     int
	FontFiles    int
	TotalSize    int64
}

// getEPUBStats extracts statistics from an EPUB file
func getEPUBStats(epubPath string) (*EPUBStats, error) {
	reader, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	stats := &EPUBStats{}

	for _, file := range reader.File {
		stats.TotalSize += int64(file.UncompressedSize64)

		ext := strings.ToLower(filepath.Ext(file.Name))
		switch {
		case ext == ".html" || ext == ".xhtml":
			if !strings.Contains(file.Name, "nav") &&
			   !strings.Contains(file.Name, "toc") &&
			   !strings.Contains(file.Name, "title") &&
			   !strings.Contains(file.Name, "cover") {
				stats.ContentFiles++
			}
		case ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp":
			stats.ImageFiles++
		case ext == ".css":
			stats.CSSFiles++
		case ext == ".ttf" || ext == ".otf" || ext == ".woff" || ext == ".woff2":
			stats.FontFiles++
		}
	}

	return stats, nil
}

func main() {
	// Parse command-line arguments
	inputPath := flag.String("i", "", "Input EPUB file path")
	outputPath := flag.String("o", "", "Output EPUB file path")
	formatDir := flag.String("f", "format", "Path to the format directory containing templates and assets")
	versionFlag := flag.Bool("v", false, "Display version information")
	debugFlag := flag.Bool("d", false, "Enable debug output")
	updateFlag := flag.Bool("u", false, "Check for updates and update if a newer version is available")
	analyzeFlag := flag.Bool("a", false, "Analyze EPUB structure without processing")
	validateFlag := flag.Bool("validate", false, "Validate EPUB structure only")
	enhancedFlag := flag.Bool("enhanced", false, "Use enhanced processing with intelligent chapter consolidation")
	compareFlag := flag.String("compare", "", "Compare two EPUB files (provide second file path)")
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

	// Set enhanced mode
	restructure.EnhancedMode = *enhancedFlag

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

	// Handle analyze-only mode
	if *analyzeFlag {
		if err := analyzeEPUB(*inputPath); err != nil {
			fmt.Printf("Error analyzing EPUB: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle validate-only mode
	if *validateFlag {
		if err := validateEPUB(*inputPath); err != nil {
			fmt.Printf("Error validating EPUB: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle compare mode
	if *compareFlag != "" {
		if err := compareEPUBs(*inputPath, *compareFlag); err != nil {
			fmt.Printf("Error comparing EPUBs: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Validate input EPUB before processing
	if err := validateEPUB(*inputPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Analyze input structure
	if *debugFlag || *enhancedFlag {
		fmt.Println("\n📊 Input Analysis:")
		if err := analyzeEPUB(*inputPath); err != nil {
			fmt.Printf("Warning: Could not analyze input EPUB: %v\n", err)
		}
		fmt.Println()
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
	fmt.Printf("🔄 Processing EPUB: %s → %s\n", *inputPath, *outputPath)
	if err := processor.Process(*inputPath, *outputPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ EPUB file successfully restructured: %s\n", *outputPath)

	// Post-processing validation and analysis
	if *debugFlag || *enhancedFlag {
		fmt.Println("\n🔍 Post-processing Validation:")
		if err := validateEPUB(*outputPath); err != nil {
			fmt.Printf("Warning: Output validation failed: %v\n", err)
		}

		fmt.Println("\n📊 Output Analysis:")
		if err := analyzeEPUB(*outputPath); err != nil {
			fmt.Printf("Warning: Could not analyze output EPUB: %v\n", err)
		}
	}
}
