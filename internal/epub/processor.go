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

// Processor handles the processing of EPUB files
type Processor struct{}

// NewProcessor creates a new EPUB processor
func NewProcessor() *Processor {
return &Processor{}
}

// Process processes an EPUB file
func (p *Processor) Process(inputPath, outputPath string) error {
// Create a temporary directory
tempDir, err := os.MkdirTemp("", "epub-restructure-*")
if err != nil {
return fmt.Errorf("failed to create temporary directory: %w", err)
}
defer os.RemoveAll(tempDir)

// Extract the EPUB file
extractedPath, err := p.extractEPUB(inputPath, tempDir)
if err != nil {
return fmt.Errorf("failed to extract EPUB: %w", err)
}

// Parse the EPUB content
book, err := parser.ParseEPUB(extractedPath)
if err != nil {
return fmt.Errorf("failed to parse EPUB: %w", err)
}

// Restructure the EPUB
restructurer := restructure.NewRestructurer()
restructuredPath, err := restructurer.Restructure(book, tempDir)
if err != nil {
return fmt.Errorf("failed to restructure EPUB: %w", err)
}

// Create the new EPUB file
if err := p.createEPUB(restructuredPath, outputPath); err != nil {
return fmt.Errorf("failed to create EPUB: %w", err)
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
// Prevent path traversal attacks
filePath := filepath.Join(extractPath, file.Name)
if !strings.HasPrefix(filePath, filepath.Clean(extractPath)+string(os.PathSeparator)) {
return "", fmt.Errorf("illegal file path: %s", file.Name)
}

// Create directory structure if needed
if file.FileInfo().IsDir() {
os.MkdirAll(filePath, 0755)
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

dstFile, err := os.Create(filePath)
if err != nil {
srcFile.Close()
return "", fmt.Errorf("failed to create file: %w", err)
}

// Limit the size of extracted files to prevent zip bombs
const maxSize = 100 * 1024 * 1024 // 100MB limit
_, err = io.CopyN(dstFile, srcFile, maxSize)
if err != nil && err != io.EOF {
srcFile.Close()
dstFile.Close()
return "", fmt.Errorf("failed to extract file or file too large: %w", err)
}

srcFile.Close()
dstFile.Close()
}

return extractPath, nil
}

// createEPUB creates a new EPUB file from the restructured content
func (p *Processor) createEPUB(restructuredPath, outputPath string) error {
// Create the output directory if it doesn't exist
outputDir := filepath.Dir(outputPath)
if err := os.MkdirAll(outputDir, 0755); err != nil {
return fmt.Errorf("failed to create output directory: %w", err)
}

// Create a new ZIP file
zipFile, err := os.Create(outputPath)
if err != nil {
return fmt.Errorf("failed to create output file: %w", err)
}
defer zipFile.Close()

// Create a new ZIP writer
zipWriter := zip.NewWriter(zipFile)
defer zipWriter.Close()

// Add the mimetype file first (uncompressed)
mimetypePath := filepath.Join(restructuredPath, "mimetype")
if err := p.addFileToZip(zipWriter, mimetypePath, "mimetype", false); err != nil {
return fmt.Errorf("failed to add mimetype to ZIP: %w", err)
}

// Walk the restructured directory and add all files to the ZIP
err = filepath.Walk(restructuredPath, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}

// Skip directories and the mimetype file (already added)
if info.IsDir() || path == mimetypePath {
return nil
}

// Get the relative path
relPath, err := filepath.Rel(restructuredPath, path)
if err != nil {
return fmt.Errorf("failed to get relative path: %w", err)
}

// Add the file to the ZIP
if err := p.addFileToZip(zipWriter, path, relPath, true); err != nil {
return fmt.Errorf("failed to add file to ZIP: %w", err)
}

return nil
})

if err != nil {
return fmt.Errorf("failed to walk restructured directory: %w", err)
}

return nil
}

// addFileToZip adds a file to the ZIP archive
func (p *Processor) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string, compress bool) error {
// Open the file
file, err := os.Open(filePath)
if err != nil {
return fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()

// Get the file info
info, err := file.Stat()
if err != nil {
return fmt.Errorf("failed to get file info: %w", err)
}

// Create a ZIP header
header, err := zip.FileInfoHeader(info)
if err != nil {
return fmt.Errorf("failed to create ZIP header: %w", err)
}

// Set the ZIP path
header.Name = zipPath

// Set the compression method
if compress {
header.Method = zip.Deflate
} else {
header.Method = zip.Store
}

// Create a writer for the file
writer, err := zipWriter.CreateHeader(header)
if err != nil {
return fmt.Errorf("failed to create ZIP writer: %w", err)
}

// Copy the file content
_, err = io.Copy(writer, file)
if err != nil {
return fmt.Errorf("failed to copy file content: %w", err)
}

return nil
}
