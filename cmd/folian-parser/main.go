package folianparser

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flouciel/folian-parser/internal/epub"
    "github.com/flouciel/folian-parser/internal/parser"
    "github.com/flouciel/folian-parser/internal/restructure"
)

// Main is the entry point for the folian-parser command
func Main() {
	// Parse command-line arguments
	inputPath := flag.String("i", "", "Input EPUB file path")
	outputPath := flag.String("o", "", "Output EPUB file path")
	flag.Parse()

	// Validate input path
	if *inputPath == "" {
		fmt.Println("Error: Input path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Generate output path if not provided
	if *outputPath == "" {
		ext := filepath.Ext(*inputPath)
		base := filepath.Base(*inputPath)
		dir := filepath.Dir(*inputPath)
		*outputPath = filepath.Join(dir, base[:len(base)-len(ext)]+"-fixed"+ext)
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
