package parser

import (
"encoding/xml"
"fmt"
"io/ioutil"
"path/filepath"
"strings"
)

// Book represents an EPUB book
type Book struct {
Path        string
Metadata    Metadata
Manifest    map[string]ManifestItem
Spine       []SpineItem
Guide       []GuideItem
Chapters    []Chapter
Stylesheets []string
Images      []string
Fonts       []string
CoverImage  string
}

// Metadata represents the metadata of an EPUB book
type Metadata struct {
Title       string
Creator     string
Language    string
Identifier  string
Publisher   string
Description string
Date        string
}

// ManifestItem represents an item in the manifest
type ManifestItem struct {
ID         string
Href       string
MediaType  string
Properties string
}

// SpineItem represents an item in the spine
type SpineItem struct {
IDRef      string
Linear     string
Properties string
}

// GuideItem represents an item in the guide
type GuideItem struct {
Type  string
Title string
Href  string
}

// Chapter represents a chapter in the book
type Chapter struct {
Title   string
Content string
}

// ParseEPUB parses an EPUB file
func ParseEPUB(epubPath string) (*Book, error) {
// Create a new book
book := &Book{
Path:     epubPath,
Manifest: make(map[string]ManifestItem),
}

// Find the content.opf file
contentOPFPath, err := findContentOPF(epubPath)
if err != nil {
return nil, fmt.Errorf("failed to find content.opf: %w", err)
}

// Read the content.opf file
contentOPF, err := ioutil.ReadFile(contentOPFPath)
if err != nil {
return nil, fmt.Errorf("failed to read content.opf: %w", err)
}

// Parse the content.opf file using XML
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
Guide struct {
Items []struct {
Type  string `xml:"type,attr"`
Title string `xml:"title,attr"`
Href  string `xml:"href,attr"`
} `xml:"reference"`
} `xml:"guide"`
}

var pkg Package
if err := xml.Unmarshal(contentOPF, &pkg); err != nil {
// If XML parsing fails, try manual parsing
if err := parseContentOPFManually(book, string(contentOPF)); err != nil {
return nil, fmt.Errorf("failed to parse content.opf: %w", err)
}
} else {
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

// Extract guide
for _, item := range pkg.Guide.Items {
book.Guide = append(book.Guide, GuideItem{
Type:  item.Type,
Title: item.Title,
Href:  item.Href,
})
}
}

// Find the chapters
if err := findChapters(book); err != nil {
return nil, fmt.Errorf("failed to find chapters: %w", err)
}

// Find the stylesheets
if err := findStylesheets(book); err != nil {
return nil, fmt.Errorf("failed to find stylesheets: %w", err)
}

// Find the images
if err := findImages(book); err != nil {
return nil, fmt.Errorf("failed to find images: %w", err)
}

// Find the fonts
if err := findFonts(book); err != nil {
return nil, fmt.Errorf("failed to find fonts: %w", err)
}

// Find the cover image if not already found
if book.CoverImage == "" {
if err := findCoverImage(book); err != nil {
return nil, fmt.Errorf("failed to find cover image: %w", err)
}
}

return book, nil
}

// findContentOPF finds the content.opf file
func findContentOPF(epubPath string) (string, error) {
// Check if the container.xml file exists
containerPath := filepath.Join(epubPath, "META-INF", "container.xml")
containerContent, err := ioutil.ReadFile(containerPath)
if err != nil {
return "", fmt.Errorf("failed to read container.xml: %w", err)
}

// Find the content.opf path
contentOPFPath := ""
if strings.Contains(string(containerContent), "full-path=") {
start := strings.Index(string(containerContent), "full-path=") + 11
end := strings.Index(string(containerContent)[start:], "\"") + start
contentOPFPath = string(containerContent)[start:end]
}

if contentOPFPath == "" {
return "", fmt.Errorf("failed to find content.opf path in container.xml")
}

// Return the full path
return filepath.Join(epubPath, contentOPFPath), nil
}

// parseContentOPFManually parses the content.opf file manually
func parseContentOPFManually(book *Book, contentOPF string) error {
// Parse metadata
if err := parseMetadataManually(book, contentOPF); err != nil {
return fmt.Errorf("failed to parse metadata: %w", err)
}

// Parse manifest
if err := parseManifestManually(book, contentOPF); err != nil {
return fmt.Errorf("failed to parse manifest: %w", err)
}

// Parse spine
if err := parseSpineManually(book, contentOPF); err != nil {
return fmt.Errorf("failed to parse spine: %w", err)
}

// Parse guide
if err := parseGuideManually(book, contentOPF); err != nil {
return fmt.Errorf("failed to parse guide: %w", err)
}

return nil
}

// parseMetadataManually parses the metadata section of the content.opf file manually
func parseMetadataManually(book *Book, contentOPF string) error {
// Find the metadata section
metadataStart := strings.Index(contentOPF, "<metadata")
if metadataStart == -1 {
return fmt.Errorf("failed to find metadata section")
}

metadataEnd := strings.Index(contentOPF[metadataStart:], "</metadata>")
if metadataEnd == -1 {
return fmt.Errorf("failed to find end of metadata section")
}
metadataEnd += metadataStart

metadata := contentOPF[metadataStart:metadataEnd]

// Parse title
if strings.Contains(metadata, "<dc:title") {
start := strings.Index(metadata, "<dc:title") + 10
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:title>")
if end != -1 {
book.Metadata.Title = metadata[start : start+end]
}
}
}

// Parse creator
if strings.Contains(metadata, "<dc:creator") {
start := strings.Index(metadata, "<dc:creator") + 12
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:creator>")
if end != -1 {
book.Metadata.Creator = metadata[start : start+end]
}
}
}

// Parse language
if strings.Contains(metadata, "<dc:language") {
start := strings.Index(metadata, "<dc:language") + 13
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:language>")
if end != -1 {
book.Metadata.Language = metadata[start : start+end]
}
}
}

// Parse identifier
if strings.Contains(metadata, "<dc:identifier") {
start := strings.Index(metadata, "<dc:identifier") + 15
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:identifier>")
if end != -1 {
book.Metadata.Identifier = metadata[start : start+end]
}
}
}

// Parse publisher
if strings.Contains(metadata, "<dc:publisher") {
start := strings.Index(metadata, "<dc:publisher") + 14
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:publisher>")
if end != -1 {
book.Metadata.Publisher = metadata[start : start+end]
}
}
}

// Parse description
if strings.Contains(metadata, "<dc:description") {
start := strings.Index(metadata, "<dc:description") + 16
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:description>")
if end != -1 {
book.Metadata.Description = metadata[start : start+end]
}
}
}

// Parse date
if strings.Contains(metadata, "<dc:date") {
start := strings.Index(metadata, "<dc:date") + 9
startTagEnd := strings.Index(metadata[start:], ">")
if startTagEnd != -1 {
start = start + startTagEnd + 1
end := strings.Index(metadata[start:], "</dc:date>")
if end != -1 {
book.Metadata.Date = metadata[start : start+end]
}
}
}

return nil
}

// parseManifestManually parses the manifest section of the content.opf file manually
func parseManifestManually(book *Book, contentOPF string) error {
// Find the manifest section
manifestStart := strings.Index(contentOPF, "<manifest")
if manifestStart == -1 {
return fmt.Errorf("failed to find manifest section")
}

manifestEnd := strings.Index(contentOPF[manifestStart:], "</manifest>")
if manifestEnd == -1 {
return fmt.Errorf("failed to find end of manifest section")
}
manifestEnd += manifestStart

manifest := contentOPF[manifestStart:manifestEnd]

// Parse items
itemStart := 0
for {
itemStart = strings.Index(manifest[itemStart:], "<item")
if itemStart == -1 {
break
}
itemStart += itemStart

itemEnd := strings.Index(manifest[itemStart:], "/>")
if itemEnd == -1 {
break
}
itemEnd += itemStart

item := manifest[itemStart:itemEnd]

// Parse ID
id := ""
if strings.Contains(item, "id=") {
start := strings.Index(item, "id=") + 4
end := strings.Index(item[start:], "\"")
if end != -1 {
id = item[start : start+end]
}
}

// Parse href
href := ""
if strings.Contains(item, "href=") {
start := strings.Index(item, "href=") + 6
end := strings.Index(item[start:], "\"")
if end != -1 {
href = item[start : start+end]
}
}

// Parse media-type
mediaType := ""
if strings.Contains(item, "media-type=") {
start := strings.Index(item, "media-type=") + 12
end := strings.Index(item[start:], "\"")
if end != -1 {
mediaType = item[start : start+end]
}
}

// Parse properties
properties := ""
if strings.Contains(item, "properties=") {
start := strings.Index(item, "properties=") + 12
end := strings.Index(item[start:], "\"")
if end != -1 {
properties = item[start : start+end]
}
}

// Add the item to the manifest
if id != "" && href != "" {
book.Manifest[id] = ManifestItem{
ID:         id,
Href:       href,
MediaType:  mediaType,
Properties: properties,
}
}

// Check for cover image
if properties == "cover-image" {
book.CoverImage = href
}

itemStart = itemEnd + 2
}

return nil
}

// parseSpineManually parses the spine section of the content.opf file manually
func parseSpineManually(book *Book, contentOPF string) error {
// Find the spine section
spineStart := strings.Index(contentOPF, "<spine")
if spineStart == -1 {
return fmt.Errorf("failed to find spine section")
}

spineEnd := strings.Index(contentOPF[spineStart:], "</spine>")
if spineEnd == -1 {
return fmt.Errorf("failed to find end of spine section")
}
spineEnd += spineStart

spine := contentOPF[spineStart:spineEnd]

// Parse items
itemStart := 0
for {
itemStart = strings.Index(spine[itemStart:], "<itemref")
if itemStart == -1 {
break
}
itemStart += itemStart

itemEnd := strings.Index(spine[itemStart:], "/>")
if itemEnd == -1 {
break
}
itemEnd += itemStart

item := spine[itemStart:itemEnd]

// Parse idref
idref := ""
if strings.Contains(item, "idref=") {
start := strings.Index(item, "idref=") + 7
end := strings.Index(item[start:], "\"")
if end != -1 {
idref = item[start : start+end]
}
}

// Parse linear
linear := ""
if strings.Contains(item, "linear=") {
start := strings.Index(item, "linear=") + 8
end := strings.Index(item[start:], "\"")
if end != -1 {
linear = item[start : start+end]
}
}

// Parse properties
properties := ""
if strings.Contains(item, "properties=") {
start := strings.Index(item, "properties=") + 12
end := strings.Index(item[start:], "\"")
if end != -1 {
properties = item[start : start+end]
}
}

// Add the item to the spine
if idref != "" {
book.Spine = append(book.Spine, SpineItem{
IDRef:      idref,
Linear:     linear,
Properties: properties,
})
}

itemStart = itemEnd + 2
}

return nil
}

// parseGuideManually parses the guide section of the content.opf file manually
func parseGuideManually(book *Book, contentOPF string) error {
// Find the guide section
guideStart := strings.Index(contentOPF, "<guide")
if guideStart == -1 {
return nil // Guide is optional
}

guideEnd := strings.Index(contentOPF[guideStart:], "</guide>")
if guideEnd == -1 {
return fmt.Errorf("failed to find end of guide section")
}
guideEnd += guideStart

guide := contentOPF[guideStart:guideEnd]

// Parse items
itemStart := 0
for {
itemStart = strings.Index(guide[itemStart:], "<reference")
if itemStart == -1 {
break
}
itemStart += itemStart

itemEnd := strings.Index(guide[itemStart:], "/>")
if itemEnd == -1 {
break
}
itemEnd += itemStart

item := guide[itemStart:itemEnd]

// Parse type
itemType := ""
if strings.Contains(item, "type=") {
start := strings.Index(item, "type=") + 6
end := strings.Index(item[start:], "\"")
if end != -1 {
itemType = item[start : start+end]
}
}

// Parse title
title := ""
if strings.Contains(item, "title=") {
start := strings.Index(item, "title=") + 7
end := strings.Index(item[start:], "\"")
if end != -1 {
title = item[start : start+end]
}
}

// Parse href
href := ""
if strings.Contains(item, "href=") {
start := strings.Index(item, "href=") + 6
end := strings.Index(item[start:], "\"")
if end != -1 {
href = item[start : start+end]
}
}

// Add the item to the guide
if itemType != "" && href != "" {
book.Guide = append(book.Guide, GuideItem{
Type:  itemType,
Title: title,
Href:  href,
})
}

itemStart = itemEnd + 2
}

return nil
}

// findChapters finds the chapters in the book
func findChapters(book *Book) error {
// Find the chapters
for _, item := range book.Spine {
// Skip non-chapter items
if item.IDRef == "" {
continue
}

// Get the manifest item
manifestItem, ok := book.Manifest[item.IDRef]
if !ok {
continue
}

// Skip non-XHTML items
if !strings.Contains(manifestItem.MediaType, "xhtml") && !strings.Contains(manifestItem.MediaType, "html") {
continue
}

// Read the chapter content
chapterPath := filepath.Join(filepath.Dir(filepath.Join(book.Path, manifestItem.Href)), filepath.Base(manifestItem.Href))
chapterContent, err := ioutil.ReadFile(chapterPath)
if err != nil {
return fmt.Errorf("failed to read chapter %s: %w", manifestItem.Href, err)
}

// Find the title
title := extractTitle(string(chapterContent))

// Add the chapter
book.Chapters = append(book.Chapters, Chapter{
Title:   title,
Content: string(chapterContent),
})
}

return nil
}

// extractTitle extracts the title from HTML content
func extractTitle(content string) string {
// Try to find title in title tag
titleStart := strings.Index(content, "<title>")
if titleStart != -1 {
titleStart += 7
titleEnd := strings.Index(content[titleStart:], "</title>")
if titleEnd != -1 {
return strings.TrimSpace(content[titleStart : titleStart+titleEnd])
}
}

// Try to find title in h1 tag
h1Start := strings.Index(content, "<h1")
if h1Start != -1 {
h1Start = strings.Index(content[h1Start:], ">")
if h1Start != -1 {
h1Start += h1Start + 1
h1End := strings.Index(content[h1Start:], "</h1>")
if h1End != -1 {
return strings.TrimSpace(content[h1Start : h1Start+h1End])
}
}
}

return "Untitled"
}

// findStylesheets finds the stylesheets in the book
func findStylesheets(book *Book) error {
// Find the stylesheets
for _, item := range book.Manifest {
// Skip non-CSS items
if !strings.Contains(item.MediaType, "css") {
continue
}

// Add the stylesheet
book.Stylesheets = append(book.Stylesheets, item.Href)
}

return nil
}

// findImages finds the images in the book
func findImages(book *Book) error {
// Find the images
for _, item := range book.Manifest {
// Skip non-image items
if !strings.Contains(item.MediaType, "image") {
continue
}

// Add the image
book.Images = append(book.Images, item.Href)
}

return nil
}

// findFonts finds the fonts in the book
func findFonts(book *Book) error {
// Find the fonts
for _, item := range book.Manifest {
// Skip non-font items
if !strings.Contains(item.MediaType, "font") {
continue
}

// Add the font
book.Fonts = append(book.Fonts, item.Href)
}

return nil
}

// findCoverImage finds the cover image in the book
func findCoverImage(book *Book) error {
// Find the cover image
for _, item := range book.Manifest {
// Check if the item has the cover-image property
if strings.Contains(item.Properties, "cover-image") {
book.CoverImage = item.Href
return nil
}
}

// Check if there's a cover item in the guide
for _, item := range book.Guide {
if item.Type == "cover" {
// Find the corresponding manifest item
for _, manifestItem := range book.Manifest {
if strings.HasSuffix(item.Href, manifestItem.Href) {
book.CoverImage = manifestItem.Href
return nil
}
}
}
}

// Check if there's a cover item in the manifest
for id, item := range book.Manifest {
if id == "cover" || strings.Contains(id, "cover") {
if strings.Contains(item.MediaType, "image") {
book.CoverImage = item.Href
return nil
}
}
}

return nil
}
