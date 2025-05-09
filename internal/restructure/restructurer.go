package restructure

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flouciel/folian-parser/internal/parser"
)

// FormatDirPath is the path to the format directory containing templates and assets
var FormatDirPath = filepath.Join("format")

// Restructurer handles the restructuring of EPUB content
type Restructurer struct{}

// NewRestructurer creates a new restructurer
func NewRestructurer() *Restructurer {
	return &Restructurer{}
}

// Restructure restructures the EPUB content according to the defined structure
func (r *Restructurer) Restructure(book *parser.Book, tempDir string) (string, error) {
	// Create a directory for the restructured content
	restructuredPath := filepath.Join(tempDir, "restructured")
	if err := os.MkdirAll(restructuredPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create restructured directory: %w", err)
	}

	// Create the standard EPUB structure
	if err := r.createStandardStructure(restructuredPath); err != nil {
		return "", fmt.Errorf("failed to create standard structure: %w", err)
	}

	// Copy and process the content
	if err := r.processContent(book, restructuredPath); err != nil {
		return "", fmt.Errorf("failed to process content: %w", err)
	}

	return restructuredPath, nil
}

// createStandardStructure creates the standard EPUB directory structure
func (r *Restructurer) createStandardStructure(basePath string) error {
	// Create META-INF directory
	metaInfPath := filepath.Join(basePath, "META-INF")
	if err := os.MkdirAll(metaInfPath, 0755); err != nil {
		return fmt.Errorf("failed to create META-INF directory: %w", err)
	}

	// Create container.xml
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if err := ioutil.WriteFile(filepath.Join(metaInfPath, "container.xml"), []byte(containerXML), 0644); err != nil {
		return fmt.Errorf("failed to create container.xml: %w", err)
	}

	// Create mimetype file
	if err := ioutil.WriteFile(filepath.Join(basePath, "mimetype"), []byte("application/epub+zip"), 0644); err != nil {
		return fmt.Errorf("failed to create mimetype file: %w", err)
	}

	// Create OEBPS directory and subdirectories
	oebpsPath := filepath.Join(basePath, "OEBPS")
	for _, dir := range []string{"", "images", "styles", "fonts", "chapters"} {
		path := filepath.Join(oebpsPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	return nil
}

// processContent processes and copies the book content to the restructured directory
func (r *Restructurer) processContent(book *parser.Book, restructuredPath string) error {
	oebpsPath := filepath.Join(restructuredPath, "OEBPS")
	basePath := filepath.Dir(filepath.Join(book.Path, book.Manifest[book.Spine[0].IDRef].Href))

	// Check if we have a cover image
	if book.CoverImage == "" {
		// Look for a cover image in the images directory
		coverFiles := []string{"cover.jpg", "cover.jpeg", "cover.png"}
		for _, coverFile := range coverFiles {
			coverPath := filepath.Join(basePath, coverFile)
			if _, err := os.Stat(coverPath); err == nil {
				book.CoverImage = coverFile
				break
			}
		}
	}

	// Process and copy stylesheets
	if err := r.processStylesheets(book, basePath, oebpsPath); err != nil {
		return fmt.Errorf("failed to process stylesheets: %w", err)
	}

	// Copy fonts
	if err := r.copyFonts(book, basePath, oebpsPath); err != nil {
		return fmt.Errorf("failed to copy fonts: %w", err)
	}

	// Copy images and handle cover
	if err := r.processImages(book, basePath, oebpsPath); err != nil {
		return fmt.Errorf("failed to process images: %w", err)
	}

	// Process chapters
	if err := r.processChapters(book, basePath, oebpsPath); err != nil {
		return fmt.Errorf("failed to process chapters: %w", err)
	}

	// Create nav.xhtml
	if err := r.createNavDocument(book, oebpsPath); err != nil {
		return fmt.Errorf("failed to create nav.xhtml: %w", err)
	}

	// Create content.opf
	if err := r.createContentOPF(book, oebpsPath); err != nil {
		return fmt.Errorf("failed to create content.opf: %w", err)
	}

	// Create toc.ncx
	if err := r.createTocNCX(book, oebpsPath); err != nil {
		return fmt.Errorf("failed to create toc.ncx: %w", err)
	}

	return nil
}

// processStylesheets processes and copies stylesheets
func (r *Restructurer) processStylesheets(book *parser.Book, basePath, oebpsPath string) error {
	// Read the stylesheet from the format directory
	stylesheetPath := filepath.Join(FormatDirPath, "stylesheet.css")
	stylesheetContent, err := ioutil.ReadFile(stylesheetPath)
	if err != nil {
		return fmt.Errorf("failed to read stylesheet from format directory: %w", err)
	}

	// Write the stylesheet
	stylesPath := filepath.Join(oebpsPath, "styles")
	if err := ioutil.WriteFile(filepath.Join(stylesPath, "stylesheet.css"), stylesheetContent, 0644); err != nil {
		return fmt.Errorf("failed to create stylesheet: %w", err)
	}

	// Copy the Jura font from the format directory
	fontPath := filepath.Join(FormatDirPath, "jura.ttf")
	fontData, err := ioutil.ReadFile(fontPath)
	if err != nil {
		fmt.Printf("Warning: Could not read Jura font: %v\n", err)
		// Continue without the font
	} else {
		// Write the font to the fonts directory
		fontsPath := filepath.Join(oebpsPath, "fonts")
		if err := ioutil.WriteFile(filepath.Join(fontsPath, "jura.ttf"), fontData, 0644); err != nil {
			return fmt.Errorf("failed to write Jura font: %w", err)
		}
	}

	return nil
}

// processCalibreStyles removes or modifies Calibre-specific styles
func (r *Restructurer) processCalibreStyles(content string) string {
	// Remove Calibre-specific comments
	content = regexp.MustCompile(`/\*\s*calibre.*?\*/`).ReplaceAllString(content, "")

	// Remove Calibre-specific classes
	content = regexp.MustCompile(`.calibre[0-9]+\s*\{.*?\}`).ReplaceAllString(content, "")

	// Clean up empty rules
	content = regexp.MustCompile(`[^}]+\{\s*\}`).ReplaceAllString(content, "")

	// Clean up multiple empty lines
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	return content
}

// copyFonts copies font files
func (r *Restructurer) copyFonts(book *parser.Book, basePath, oebpsPath string) error {
	fontsPath := filepath.Join(oebpsPath, "fonts")

	for _, fontPath := range book.Fonts {
		// Read the font file
		fullPath := filepath.Join(basePath, fontPath)
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read font %s: %w", fontPath, err)
		}

		// Write the font file
		filename := filepath.Base(fontPath)
		outputPath := filepath.Join(fontsPath, filename)
		if err := ioutil.WriteFile(outputPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write font %s: %w", filename, err)
		}
	}

	return nil
}

// processImages processes and copies image files
func (r *Restructurer) processImages(book *parser.Book, basePath, oebpsPath string) error {
	imagesPath := filepath.Join(oebpsPath, "images")

	// Process cover image if it exists
	var coverFilename string
	if book.CoverImage != "" {
		// Extract the base filename of the cover image
		coverImageBase := filepath.Base(book.CoverImage)

		// Try to find the cover image in the extracted directory
		// First, try the path as specified in the manifest
		fullPath := filepath.Join(basePath, book.CoverImage)
		content, err := ioutil.ReadFile(fullPath)

		// If that fails, try looking in the OEBPS directory
		if err != nil {
			fullPath = filepath.Join(filepath.Dir(basePath), "OEBPS", book.CoverImage)
			content, err = ioutil.ReadFile(fullPath)

			// If that fails, try looking in the OEBPS/images directory
			if err != nil {
				fullPath = filepath.Join(filepath.Dir(basePath), "OEBPS", "images", coverImageBase)
				content, err = ioutil.ReadFile(fullPath)

				// If that fails, try looking directly in the extracted directory
				if err != nil {
					fullPath = filepath.Join(filepath.Dir(basePath), coverImageBase)
					content, err = ioutil.ReadFile(fullPath)

					// If all attempts fail, search for any file with the same name
					if err != nil {
						// Search for the cover image in the entire extracted directory
						extractedDir := filepath.Dir(basePath)
						fmt.Printf("Searching for cover image %s in %s\n", coverImageBase, extractedDir)

						// Use filepath.Walk to search for the file
						var coverPath string
						filepath.Walk(extractedDir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return nil
							}
							if !info.IsDir() && filepath.Base(path) == coverImageBase {
								coverPath = path
								return filepath.SkipDir // Stop walking once we find the file
							}
							return nil
						})

						// If we found the file, read it
						if coverPath != "" {
							fullPath = coverPath
							content, err = ioutil.ReadFile(fullPath)
						}

						// If we still can't find the file, return an error
						if err != nil {
							return fmt.Errorf("failed to find cover image %s: %w", book.CoverImage, err)
						}
					}
				}
			}
		}

		// Write the cover image
		coverFilename = "cover" + filepath.Ext(book.CoverImage)
		outputPath := filepath.Join(imagesPath, coverFilename)
		if err := ioutil.WriteFile(outputPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write cover image: %w", err)
		}

		// Create titlepage.xhtml from template
		titlePagePath := filepath.Join(FormatDirPath, "titlepage.xhtml")
		titlePageContent, err := ioutil.ReadFile(titlePagePath)
		if err != nil {
			return fmt.Errorf("failed to read titlepage template from format directory: %w", err)
		}

		// Replace the cover image reference
		titlePageContentStr := string(titlePageContent)
		titlePageContentStr = strings.Replace(titlePageContentStr,
			`<image width="1038" height="1380" xlink:href="images/cover.jpg"/>`,
			fmt.Sprintf(`<image width="1038" height="1380" xlink:href="images/%s"/>`, coverFilename),
			-1)
		titlePageContent = []byte(titlePageContentStr)

		if err := ioutil.WriteFile(filepath.Join(oebpsPath, "titlepage.xhtml"), titlePageContent, 0644); err != nil {
			return fmt.Errorf("failed to create titlepage.xhtml: %w", err)
		}

		// Create jacket.xhtml from template
		jacketPath := filepath.Join(FormatDirPath, "jacket.xhtml")
		jacketContent, err := ioutil.ReadFile(jacketPath)
		if err != nil {
			return fmt.Errorf("failed to read jacket template from format directory: %w", err)
		}

		// Extract book title and author from metadata
		title := book.Metadata.Title
		if title == "" {
			title = "Book Title"
		}

		author := book.Metadata.Creator
		if author == "" {
			author = "Author"
		}

		// Replace template variables in jacket.xhtml
		jacketContentStr := string(jacketContent)
		jacketContentStr = strings.Replace(jacketContentStr, "{{BOOK_TITLE}}", title, -1)
		jacketContentStr = strings.Replace(jacketContentStr, "{{BOOK_AUTHOR}}", author, -1)

		// Set a default subtitle or use a description if available
		subtitle := "A Folian Book"
		if book.Metadata.Description != "" {
			// Use a shortened version of the description as subtitle
			if len(book.Metadata.Description) > 60 {
				subtitle = book.Metadata.Description[:57] + "..."
			} else {
				subtitle = book.Metadata.Description
			}
		}
		jacketContentStr = strings.Replace(jacketContentStr, "{{BOOK_SUBTITLE}}", subtitle, -1)

		jacketContent = []byte(jacketContentStr)

		if err := ioutil.WriteFile(filepath.Join(oebpsPath, "jacket.xhtml"), jacketContent, 0644); err != nil {
			return fmt.Errorf("failed to create jacket.xhtml: %w", err)
		}

		// Copy Folian logo if it exists
		folianLogoPath := filepath.Join(FormatDirPath, "folian.png")
		if _, err := os.Stat(folianLogoPath); err == nil {
			folianLogoContent, err := ioutil.ReadFile(folianLogoPath)
			if err == nil {
				if err := ioutil.WriteFile(filepath.Join(imagesPath, "folian.png"), folianLogoContent, 0644); err != nil {
					fmt.Printf("Warning: Failed to copy Folian logo: %v\n", err)
				}
			}
		}
	}

	// Copy other images
	for _, imagePath := range book.Images {
		// Skip the cover image as we've already processed it
		if imagePath == book.CoverImage {
			continue
		}

		// Extract the base filename of the image
		imageBase := filepath.Base(imagePath)

		// Try to find the image in the extracted directory
		// First, try the path as specified in the manifest
		fullPath := filepath.Join(basePath, imagePath)
		content, err := ioutil.ReadFile(fullPath)

		// If that fails, try looking in the OEBPS directory
		if err != nil {
			fullPath = filepath.Join(filepath.Dir(basePath), "OEBPS", imagePath)
			content, err = ioutil.ReadFile(fullPath)

			// If that fails, try looking in the OEBPS/images directory
			if err != nil {
				fullPath = filepath.Join(filepath.Dir(basePath), "OEBPS", "images", imageBase)
				content, err = ioutil.ReadFile(fullPath)

				// If that fails, try looking directly in the extracted directory
				if err != nil {
					fullPath = filepath.Join(filepath.Dir(basePath), imageBase)
					content, err = ioutil.ReadFile(fullPath)

					// If all attempts fail, search for any file with the same name
					if err != nil {
						// Search for the image in the entire extracted directory
						extractedDir := filepath.Dir(basePath)

						// Use filepath.Walk to search for the file
						var imagePath string
						filepath.Walk(extractedDir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return nil
							}
							if !info.IsDir() && filepath.Base(path) == imageBase {
								imagePath = path
								return filepath.SkipDir // Stop walking once we find the file
							}
							return nil
						})

						// If we found the file, read it
						if imagePath != "" {
							fullPath = imagePath
							content, err = ioutil.ReadFile(fullPath)
						}

						// If we still can't find the file, log a warning and continue
						if err != nil {
							fmt.Printf("Warning: failed to find image %s: %v\n", imagePath, err)
							continue
						}
					}
				}
			}
		}

		// Write the image file
		filename := filepath.Base(imagePath)
		outputPath := filepath.Join(imagesPath, filename)
		if err := ioutil.WriteFile(outputPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write image %s: %w", filename, err)
		}
	}

	return nil
}

// processChapters processes and copies chapter files
func (r *Restructurer) processChapters(book *parser.Book, basePath, oebpsPath string) error {
	chaptersPath := filepath.Join(oebpsPath, "chapters")

	// Process each chapter
	for i, chapter := range book.Chapters {
		// Extract title from content or use book title
		titleMatch := regexp.MustCompile(`<title>([^<]+)</title>`).FindStringSubmatch(chapter.Content)
		chapterTitle := fmt.Sprintf("Chapter %d", i+1)

		// Use book title if provided
		bookTitle := book.Metadata.Title
		if bookTitle != "" {
			if len(titleMatch) > 1 && titleMatch[1] != bookTitle {
				chapterTitle = fmt.Sprintf("Chapter %d: %s", i+1, titleMatch[1])
			} else {
				chapterTitle = fmt.Sprintf("Chapter %d: %s", i+1, bookTitle)
			}
		} else if len(titleMatch) > 1 {
			chapterTitle = fmt.Sprintf("Chapter %d: %s", i+1, titleMatch[1])
		}

		// Create the chapter content
		processedContent := fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
  <title>%s</title>
  <link href="../styles/stylesheet.css" rel="stylesheet" type="text/css"/>
</head>

<body>

    <h1>Chapter %d</h1>

`, chapterTitle, i+1)

		// Extract paragraphs from content
		bodyContent := chapter.Content

		// Remove the XML declaration and DOCTYPE
		bodyContent = regexp.MustCompile(`<\?xml[^>]*\?>`).ReplaceAllString(bodyContent, "")
		bodyContent = regexp.MustCompile(`<!DOCTYPE[^>]*>`).ReplaceAllString(bodyContent, "")

		// Extract the body content
		bodyMatch := regexp.MustCompile(`<body[^>]*>(.*?)</body>`).FindStringSubmatch(bodyContent)
		if len(bodyMatch) > 1 {
			bodyContent = bodyMatch[1]
		}

		// Remove Calibre-specific classes
		bodyContent = regexp.MustCompile(`class="calibre[^"]*"`).ReplaceAllString(bodyContent, "")

		// Extract paragraphs
		paragraphs := regexp.MustCompile(`<p[^>]*>(.*?)</p>`).FindAllStringSubmatch(bodyContent, -1)

		// If no paragraphs found, try to extract text directly
		if len(paragraphs) == 0 {
			// Split by newlines and create paragraphs
			lines := strings.Split(bodyContent, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "<") && !strings.HasSuffix(line, ">") {
					processedContent += fmt.Sprintf("  <p>%s</p>\n\n", line)
				}
			}
		} else {
			// Process each paragraph
			for _, p := range paragraphs {
				if len(p) > 1 {
					paragraphContent := p[1]

					// Remove nested spans with Calibre classes
					paragraphContent = regexp.MustCompile(`<span class="calibre[^"]*">(.*?)</span>`).ReplaceAllString(paragraphContent, "$1")

					// Fix image paths
					paragraphContent = regexp.MustCompile(`src="([^"]+\.(jpg|jpeg|png|gif))"`).ReplaceAllString(paragraphContent, `src="../images/$1"`)

					// Only add non-empty paragraphs
					if strings.TrimSpace(paragraphContent) != "" {
						processedContent += fmt.Sprintf("  <p>%s</p>\n\n", paragraphContent)
					}
				}
			}
		}

		// If no paragraphs were added, add a default one
		if !strings.Contains(processedContent, "<p>") {
			processedContent += "  <p>Chapter content</p>\n\n"
		}

		// Close the HTML
		processedContent += "</body>\n\n</html>"

		// Write the processed chapter
		filename := fmt.Sprintf("chapter_%03d.xhtml", i+1)
		outputPath := filepath.Join(chaptersPath, filename)
		if err := ioutil.WriteFile(outputPath, []byte(processedContent), 0644); err != nil {
			return fmt.Errorf("failed to write chapter %s: %w", filename, err)
		}
	}

	return nil
}

// processChapterContent processes chapter content
func (r *Restructurer) processChapterContent(content string, chapterNum int) string {
	// Extract title from content
	titleMatch := regexp.MustCompile(`<title>([^<]+)</title>`).FindStringSubmatch(content)
	title := fmt.Sprintf("Chapter %d", chapterNum)
	if len(titleMatch) > 1 {
		title = fmt.Sprintf("Chapter %d: %s", chapterNum, titleMatch[1])
	}

	// Extract paragraphs from content
	bodyMatch := regexp.MustCompile(`<body[^>]*>(.*?)</body>`).FindStringSubmatch(content)
	if len(bodyMatch) < 2 {
		// If no body found, return a basic chapter template
		return r.createBasicChapterTemplate(title, chapterNum)
	}

	bodyContent := bodyMatch[1]

	// Remove Calibre-specific elements and classes
	bodyContent = regexp.MustCompile(`<div class="calibre[^"]*"[^>]*>|</div>`).ReplaceAllString(bodyContent, "")
	bodyContent = regexp.MustCompile(`class="calibre[^"]*"`).ReplaceAllString(bodyContent, "")

	// Extract paragraphs
	paragraphs := regexp.MustCompile(`<p[^>]*>(.*?)</p>`).FindAllStringSubmatch(bodyContent, -1)

	// Create new chapter content based on the template
	chapterContent := fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
  <title>%s</title>
  <link href="../styles/stylesheet.css" rel="stylesheet" type="text/css"/>
</head>

<body>

    <h1>%s</h1>

`, title, fmt.Sprintf("Chapter %d", chapterNum))

	// Add paragraphs
	for _, p := range paragraphs {
		if len(p) > 1 && strings.TrimSpace(p[1]) != "" {
			// Clean up any nested tags but preserve basic formatting
			paragraphContent := p[1]
			// Fix image paths
			paragraphContent = regexp.MustCompile(`src="([^"]+\.(jpg|jpeg|png|gif))"`).ReplaceAllString(paragraphContent, `src="../images/$1"`)

			chapterContent += fmt.Sprintf("  <p>%s</p>\n\n", paragraphContent)
		}
	}

	chapterContent += "</body>\n\n</html>"
	return chapterContent
}

// createBasicChapterTemplate creates a basic chapter template
func (r *Restructurer) createBasicChapterTemplate(title string, chapterNum int) string {
	return fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
  <title>%s</title>
  <link href="../styles/stylesheet.css" rel="stylesheet" type="text/css"/>
</head>

<body>

    <h1>%s</h1>

  <p>Chapter content</p>

</body>

</html>`, title, fmt.Sprintf("Chapter %d", chapterNum))
}

// createContentOPF creates the content.opf file
func (r *Restructurer) createContentOPF(book *parser.Book, oebpsPath string) error {
	// Start building the OPF content
	subtitleMeta := ""

	opfContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="BookID">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>%s</dc:language>
    <dc:identifier id="BookID">%s</dc:identifier>
    <dc:publisher>%s</dc:publisher>
    <dc:description>%s</dc:description>
    <dc:date>%s</dc:date>
    <meta name="cover" content="cover-image"/>%s
  </metadata>
  <manifest>
`,
		book.Metadata.Title,
		book.Metadata.Creator,
		book.Metadata.Language,
		book.Metadata.Identifier,
		book.Metadata.Publisher,
		book.Metadata.Description,
		book.Metadata.Date,
		subtitleMeta)

	// Add items to manifest
	manifestItems := []string{
		`    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>`,
		`    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>`,
	}

	// Add titlepage and jacket
	hasCover := book.CoverImage != ""
	if hasCover {
		manifestItems = append(manifestItems, `    <item id="titlepage" href="titlepage.xhtml" media-type="application/xhtml+xml" properties="svg"/>`)
		manifestItems = append(manifestItems, `    <item id="jacket" href="jacket.xhtml" media-type="application/xhtml+xml"/>`)
		manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="cover-image" href="images/cover%s" media-type="image/%s" properties="cover-image"/>`,
			filepath.Ext(book.CoverImage),
			strings.TrimPrefix(filepath.Ext(book.CoverImage), ".")))

		// Add Folian logo if it exists
		folianLogoPath := filepath.Join(FormatDirPath, "folian.png")
		if _, err := os.Stat(folianLogoPath); err == nil {
			manifestItems = append(manifestItems, `    <item id="folian-logo" href="images/folian.png" media-type="image/png"/>`)
		}
	}

	// Add stylesheets
	manifestItems = append(manifestItems, `    <item id="stylesheet" href="styles/stylesheet.css" media-type="text/css"/>`)
	//for i, stylesheet := range book.Stylesheets {
	//	manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="style%d" href="styles/%s" media-type="text/css"/>`, i+1, filepath.Base(stylesheet)))
	//}

	// Add chapters
	for i := range book.Chapters {
		manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="chapter%d" href="chapters/chapter_%03d.xhtml" media-type="application/xhtml+xml"/>`, i+1, i+1))
	}

	// Add images
	for i, imagePath := range book.Images {
		if imagePath != book.CoverImage {
			ext := filepath.Ext(imagePath)
			mediaType := "image/jpeg"
			if ext == ".png" {
				mediaType = "image/png"
			} else if ext == ".gif" {
				mediaType = "image/gif"
			}
			manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="image%d" href="images/%s" media-type="%s"/>`, i+1, filepath.Base(imagePath), mediaType))
		}
	}

	// Add fonts
	manifestItems = append(manifestItems, `    <item id="jura-font" href="fonts/jura.ttf" media-type="application/x-font-ttf"/>`)
	for i, fontPath := range book.Fonts {
		ext := filepath.Ext(fontPath)
		mediaType := "application/font-sfnt"
		if ext == ".ttf" {
			mediaType = "application/x-font-ttf"
		} else if ext == ".otf" {
			mediaType = "application/x-font-opentype"
		} else if ext == ".woff" {
			mediaType = "application/font-woff"
		}
		manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="font%d" href="fonts/%s" media-type="%s"/>`, i+1, filepath.Base(fontPath), mediaType))
	}

	// Add manifest items to OPF
	opfContent += strings.Join(manifestItems, "\n") + "\n  </manifest>\n  <spine toc=\"ncx\">\n"

	// Add spine items
	spineItems := []string{}

	// Add titlepage and jacket to spine
	if hasCover {
		spineItems = append(spineItems, `    <itemref idref="titlepage"/>`)
		spineItems = append(spineItems, `    <itemref idref="jacket"/>`)
	}

	// Add nav document to spine
	spineItems = append(spineItems, `    <itemref idref="nav"/>`)

	// Add chapters to spine
	for i := range book.Chapters {
		spineItems = append(spineItems, fmt.Sprintf(`    <itemref idref="chapter%d"/>`, i+1))
	}

	// Add spine items to OPF
	opfContent += strings.Join(spineItems, "\n") + "\n  </spine>\n</package>"

	// Write the OPF file
	return ioutil.WriteFile(filepath.Join(oebpsPath, "content.opf"), []byte(opfContent), 0644)
}

// createNavDocument creates the nav.xhtml file for EPUB3 navigation
func (r *Restructurer) createNavDocument(book *parser.Book, oebpsPath string) error {
	// Read the nav.xhtml template from the format directory
	navTemplatePath := filepath.Join(FormatDirPath, "nav.xhtml")
	navTemplate, err := ioutil.ReadFile(navTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read nav.xhtml template from format directory: %w", err)
	}

	// Replace book title
	navContent := string(navTemplate)
	navContent = strings.Replace(navContent, "{{BOOK_TITLE}}", book.Metadata.Title, -1)

	// Generate TOC entries
	var tocEntries strings.Builder

	// Add chapters to TOC
	for i, chapter := range book.Chapters {
		chapterPath := fmt.Sprintf("chapters/chapter_%03d.xhtml", i+1)
		tocEntries.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", chapterPath, chapter.Title))
	}

	// Replace TOC entries placeholder
	navContent = strings.Replace(navContent, "{{TOC_ENTRIES}}", tocEntries.String(), -1)

	// Write the nav.xhtml file
	if err := ioutil.WriteFile(filepath.Join(oebpsPath, "nav.xhtml"), []byte(navContent), 0644); err != nil {
		return fmt.Errorf("failed to write nav.xhtml: %w", err)
	}

	return nil
}

// createTocNCX creates the toc.ncx file
func (r *Restructurer) createTocNCX(book *parser.Book, oebpsPath string) error {
	// Start building the NCX content
	ncxContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="%s"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>%s</text>
  </docTitle>
  <docAuthor>
    <text>%s</text>
  </docAuthor>
  <navMap>
`,
		book.Metadata.Identifier,
		book.Metadata.Title,
		book.Metadata.Creator)

	// Add nav points
	navPoints := []string{}

	// Add titlepage and jacket if cover exists
	playOrder := 1
	if book.CoverImage != "" {
		navPoints = append(navPoints, fmt.Sprintf(`    <navPoint id="navpoint-titlepage" playOrder="%d">
      <navLabel>
        <text>Cover</text>
      </navLabel>
      <content src="titlepage.xhtml"/>
    </navPoint>`, playOrder))
		playOrder++

		navPoints = append(navPoints, fmt.Sprintf(`    <navPoint id="navpoint-jacket" playOrder="%d">
      <navLabel>
        <text>Title Page</text>
      </navLabel>
      <content src="jacket.xhtml"/>
    </navPoint>`, playOrder))
		playOrder++
	}

	// Add chapters
	for i, chapter := range book.Chapters {
		navPoints = append(navPoints, fmt.Sprintf(`    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel>
        <text>%s</text>
      </navLabel>
      <content src="chapters/chapter_%03d.xhtml"/>
    </navPoint>`, i+1, i+playOrder, chapter.Title, i+1))
	}

	// Add nav points to NCX
	ncxContent += strings.Join(navPoints, "\n") + "\n  </navMap>\n</ncx>"

	// Write the NCX file
	return ioutil.WriteFile(filepath.Join(oebpsPath, "toc.ncx"), []byte(ncxContent), 0644)
}