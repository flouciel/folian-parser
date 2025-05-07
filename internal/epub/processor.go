package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flouciel/folian-parser/internal/parser"
	"github.com/flouciel/folian-parser/internal/restructure"
)

// Processor handles the EPUB processing workflow
type Processor struct {
	parser      *parser.EPUBParser
	restructure *restructure.Restructurer
}

// NewProcessor creates a new EPUB processor
func NewProcessor() *Processor {
	return &Processor{
		parser:      parser.NewEPUBParser(),
		restructure: restructure.NewRestructurer(),
	}
}

// Process takes an input EPUB file, restructures it, and saves it to the output path
func (p *Processor) Process(inputPath, outputPath string) error {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "epub-restructure-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract the EPUB file
	extractedPath, err := p.extractEPUB(inputPath, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract EPUB: %w", err)
	}

	// Parse the EPUB content
	book, err := p.parser.Parse(extractedPath)
	if err != nil {
		return fmt.Errorf("failed to parse EPUB: %w", err)
	}

	// Restructure the EPUB
	restructuredPath, err := p.restructure.Restructure(book, tempDir)
	if err != nil {
		return fmt.Errorf("failed to restructure EPUB: %w", err)
	}

	// Create the new EPUB file
	err = p.createEPUB(restructuredPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output EPUB: %w", err)
	}

	return nil
}

// extractEPUB extracts the EPUB file to a temporary directory
func (p *Processor) extractEPUB(epubPath, tempDir string) (string, error) {
	// Open the EPUB file (which is a ZIP archive)
	reader, err := zip.OpenReader(epubPath)
	if err != nil {
		return "", fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer reader.Close()

	// Create a directory for the extracted content
	extractPath := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Extract all files
	for _, file := range reader.File {
		// Validate file path to prevent path traversal
		filePath := filepath.Join(extractPath, file.Name)
		if !strings.HasPrefix(filePath, extractPath) {
			return "", fmt.Errorf("invalid file path (potential path traversal attack): %s", file.Name)
		}

		// Create directory structure if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract the file
		srcFile, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file in archive: %w", err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(filePath)
		if err != nil {
			srcFile.Close()
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		defer dstFile.Close()

		// Limit the size of extracted files to prevent zip bombs
		const maxSize = 100 * 1024 * 1024 // 100MB limit per file
		_, err = io.CopyN(dstFile, srcFile, maxSize)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to extract file or file too large: %w", err)
		}
	}

	return extractPath, nil
}

// createEPUB creates a new EPUB file from the restructured content
func (p *Processor) createEPUB(contentPath, outputPath string) error {
	// Create the output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(outputFile)
	defer zipWriter.Close()

	// Add mimetype file first (must be uncompressed and first in the archive)
	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // No compression for mimetype
	})
	if err != nil {
		return fmt.Errorf("failed to create mimetype entry: %w", err)
	}
	_, err = mimetypeWriter.Write([]byte("application/epub+zip"))
	if err != nil {
		return fmt.Errorf("failed to write mimetype: %w", err)
	}

	// Walk through the restructured content and add all files to the ZIP
	err = filepath.Walk(contentPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path for the ZIP entry
		relPath, err := filepath.Rel(contentPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Normalize path separators to forward slashes for EPUB
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Skip the mimetype file as we've already added it
		if relPath == "mimetype" {
			return nil
		}

		// Create a new file in the ZIP
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry: %w", err)
		}

		// Open the source file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Copy the file content to the ZIP
		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to write file to ZIP: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to add files to EPUB: %w", err)
	}

	return nil
}