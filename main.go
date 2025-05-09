package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flouciel/folian-parser/internal/epub"
	"github.com/flouciel/folian-parser/internal/restructure"
)

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
	formatDir := flag.String("format", "format", "Path to the format directory containing templates and assets")
	flag.Parse()

	// Set the format directory path
	restructure.FormatDirPath = *formatDir

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
