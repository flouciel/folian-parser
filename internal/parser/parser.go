package parser

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// EPUBParser parses EPUB files
type EPUBParser struct{}

// NewEPUBParser creates a new EPUB parser
func NewEPUBParser() *EPUBParser {
	return &EPUBParser{}
}

// Book represents the parsed EPUB book
type Book struct {
	Path        string
	Metadata    Metadata
	Spine       []SpineItem
	Manifest    map[string]ManifestItem
	CoverImage  string
	Stylesheets []string
	Fonts       []string
	Images      []string
	Chapters    []Chapter
}

// Metadata contains the book metadata
type Metadata struct {
	Title       string
	Creator     string
	Language    string
	Identifier  string
	Publisher   string
	Description string
	Date        string
}

// ManifestItem represents an item in the EPUB manifest
type ManifestItem struct {
	ID         string
	Href       string
	MediaType  string
	Properties string
}

// SpineItem represents an item in the EPUB spine
type SpineItem struct {
	IDRef      string
	Linear     string
	Properties string
}

// Chapter represents a chapter in the book
type Chapter struct {
	ID      string
	Title   string
	Content string
	Order   int
}

// Parse parses an extracted EPUB file
func (p *EPUBParser) Parse(epubPath string) (*Book, error) {
	book := &Book{
		Path:     epubPath,
		Manifest: make(map[string]ManifestItem),
	}

	// Find and parse the container.xml file
	containerPath := filepath.Join(epubPath, "META-INF", "container.xml")
	rootFilePath, err := p.parseContainer(containerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %w", err)
	}

	// Parse the OPF file
	opfPath := filepath.Join(epubPath, rootFilePath)
	err = p.parseOPF(opfPath, book)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF file: %w", err)
	}

	// Categorize files and parse chapters
	err = p.categorizeFiles(book, filepath.Dir(opfPath))
	if err != nil {
		return nil, fmt.Errorf("failed to categorize files: %w", err)
	}

	return book, nil
}

// parseContainer parses the container.xml file to find the OPF file
func (p *EPUBParser) parseContainer(containerPath string) (string, error) {
	data, err := ioutil.ReadFile(containerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read container.xml: %w", err)
	}

	type RootFile struct {
		FullPath  string `xml:"full-path,attr"`
		MediaType string `xml:"media-type,attr"`
	}

	type RootFiles struct {
		RootFile []RootFile `xml:"rootfile"`
	}

	type Container struct {
		RootFiles RootFiles `xml:"rootfiles"`
	}

	var container Container
	if err := xml.Unmarshal(data, &container); err != nil {
		return "", fmt.Errorf("failed to unmarshal container.xml: %w", err)
	}

	if len(container.RootFiles.RootFile) == 0 {
		return "", fmt.Errorf("no root file found in container.xml")
	}

	return container.RootFiles.RootFile[0].FullPath, nil
}

// parseOPF parses the OPF file
func (p *EPUBParser) parseOPF(opfPath string, book *Book) error {
	data, err := ioutil.ReadFile(opfPath)
	if err != nil {
		return fmt.Errorf("failed to read OPF file: %w", err)
	}

	type Package struct {
		Metadata struct {
			Title       []string `xml:"title"`
			Creator     []string `xml:"creator"`
			Language    []string `xml:"language"`
			Identifier  []string `xml:"identifier"`
			Publisher   []string `xml:"publisher"`
			Description []string `xml:"description"`
			Date        []string `xml:"date"`
		} `xml:"metadata"`
		Manifest struct {
			Items []struct {
				ID         string `xml:"id,attr"`
				Href       string `xml:"href,attr"`
				MediaType  string `xml:"media-type,attr"`
				Properties string `xml:"properties,attr"`
			} `xml:"item"`
		} `xml:"manifest"`
		Spine struct {
			Items []struct {
				IDRef      string `xml:"idref,attr"`
				Linear     string `xml:"linear,attr"`
				Properties string `xml:"properties,attr"`
			} `xml:"itemref"`
		} `xml:"spine"`
	}

	var pkg Package
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to unmarshal OPF file: %w", err)
	}

	// Extract metadata
	if len(pkg.Metadata.Title) > 0 {
		book.Metadata.Title = pkg.Metadata.Title[0]
	}
	if len(pkg.Metadata.Creator) > 0 {
		book.Metadata.Creator = pkg.Metadata.Creator[0]
	}
	if len(pkg.Metadata.Language) > 0 {
		book.Metadata.Language = pkg.Metadata.Language[0]
	}
	if len(pkg.Metadata.Identifier) > 0 {
		book.Metadata.Identifier = pkg.Metadata.Identifier[0]
	}
	if len(pkg.Metadata.Publisher) > 0 {
		book.Metadata.Publisher = pkg.Metadata.Publisher[0]
	}
	if len(pkg.Metadata.Description) > 0 {
		book.Metadata.Description = pkg.Metadata.Description[0]
	}
	if len(pkg.Metadata.Date) > 0 {
		book.Metadata.Date = pkg.Metadata.Date[0]
	}

	// Extract manifest
	for _, item := range pkg.Manifest.Items {
		book.Manifest[item.ID] = ManifestItem{
			ID:         item.ID,
			Href:       item.Href,
			MediaType:  item.MediaType,
			Properties: item.Properties,
		}

		// Check for cover image
		if item.Properties == "cover-image" {
			book.CoverImage = item.Href
		}
	}

	// Extract spine
	for _, item := range pkg.Spine.Items {
		book.Spine = append(book.Spine, SpineItem{
			IDRef:      item.IDRef,
			Linear:     item.Linear,
			Properties: item.Properties,
		})
	}

	return nil
}

// categorizeFiles categorizes files in the EPUB and parses chapters
func (p *EPUBParser) categorizeFiles(book *Book, basePath string) error {
	// Categorize files by type
	for _, item := range book.Manifest {
		switch {
		case strings.Contains(item.MediaType, "text/css"):
			book.Stylesheets = append(book.Stylesheets, item.Href)
		case strings.Contains(item.MediaType, "font/"):
			book.Fonts = append(book.Fonts, item.Href)
		case strings.Contains(item.MediaType, "image/"):
			book.Images = append(book.Images, item.Href)
		}
	}

	// Parse chapters based on spine
	for i, spineItem := range book.Spine {
		manifestItem, ok := book.Manifest[spineItem.IDRef]
		if !ok {
			continue
		}

		// Skip non-XHTML files
		if !strings.Contains(manifestItem.MediaType, "application/xhtml+xml") {
			continue
		}

		// Read the chapter content
		chapterPath := filepath.Join(basePath, manifestItem.Href)
		content, err := ioutil.ReadFile(chapterPath)
		if err != nil {
			return fmt.Errorf("failed to read chapter file: %w", err)
		}

		// Extract title from content (simplified)
		title := p.extractTitle(string(content))

		// Create chapter
		chapter := Chapter{
			ID:      manifestItem.ID,
			Title:   title,
			Content: string(content),
			Order:   i,
		}

		book.Chapters = append(book.Chapters, chapter)
	}

	return nil
}

// extractTitle extracts the title from HTML content (simplified)
func (p *EPUBParser) extractTitle(content string) string {
	// Try to find title in h1 tag
	h1Start := strings.Index(content, "<h1")
	if h1Start != -1 {
		h1End := strings.Index(content[h1Start:], "</h1>")
		if h1End != -1 {
			h1Content := content[h1Start : h1Start+h1End+5]
			// Extract text between > and <
			gtIndex := strings.Index(h1Content, ">")
			if gtIndex != -1 {
				return strings.TrimSpace(h1Content[gtIndex+1 : len(h1Content)-5])
			}
		}
	}

	// Try to find title in title tag
	titleStart := strings.Index(content, "<title>")
	if titleStart != -1 {
		titleEnd := strings.Index(content[titleStart:], "</title>")
		if titleEnd != -1 {
			return strings.TrimSpace(content[titleStart+7 : titleStart+titleEnd])
		}
	}

	return "Untitled"
}