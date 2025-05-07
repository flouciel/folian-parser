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
// Create a main stylesheet
mainStylesheet := `/* Main stylesheet */
@font-face {
  font-family: 'Jura';
  src: url('../fonts/jura.ttf');
  font-weight: normal;
  font-style: normal;
}

body {
  font-family: 'Jura', serif;
  margin: 5%;
  text-align: justify;
  line-height: 1.5;
}

h1, h2, h3, h4, h5, h6 {
  text-align: center;
  font-weight: bold;
  margin: 1em 0;
  font-family: 'Jura', sans-serif;
}

img {
  max-width: 100%;
}

.cover {
  text-align: center;
  page-break-after: always;
}

.chapter {
  page-break-before: always;
}

.title {
  font-size: 2em;
  font-weight: bold;
  text-align: center;
  margin: 1em 0;
}

.author {
  font-size: 1.5em;
  text-align: center;
  margin: 1em 0;
}
`

// Write the main stylesheet
stylesPath := filepath.Join(oebpsPath, "styles")
if err := ioutil.WriteFile(filepath.Join(stylesPath, "stylesheet.css"), []byte(mainStylesheet), 0644); err != nil {
return fmt.Errorf("failed to create main stylesheet: %w", err)
}

// Process and copy original stylesheets
for _, stylesheetPath := range book.Stylesheets {
// Read the original stylesheet
fullPath := filepath.Join(basePath, stylesheetPath)
content, err := ioutil.ReadFile(fullPath)
if err != nil {
return fmt.Errorf("failed to read stylesheet %s: %w", stylesheetPath, err)
}

// Process the stylesheet to remove Calibre-specific styles
processedContent := r.processCalibreStyles(string(content))

// Write the processed stylesheet
filename := filepath.Base(stylesheetPath)
outputPath := filepath.Join(stylesPath, filename)
if err := ioutil.WriteFile(outputPath, []byte(processedContent), 0644); err != nil {
return fmt.Errorf("failed to write processed stylesheet %s: %w", filename, err)
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

// Copy Jura font
juraFontPath := filepath.Join("/Users/4azy/lune/code/pub/1984-custom-fixed-dir/OEBPS/fonts/jura.ttf")
if _, err := os.Stat(juraFontPath); err == nil {
juraFontContent, err := ioutil.ReadFile(juraFontPath)
if err != nil {
return fmt.Errorf("failed to read Jura font: %w", err)
}
if err := ioutil.WriteFile(filepath.Join(fontsPath, "jura.ttf"), juraFontContent, 0644); err != nil {
return fmt.Errorf("failed to write Jura font: %w", err)
}
} else {
fmt.Println("Warning: Jura font not found, using default fonts")
}

// Copy other fonts
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

// Copy Folian logo
folianLogoPath := filepath.Join("/Users/4azy/lune/code/pub/1984-custom-fixed-dir/OEBPS/images/folian.png")
if _, err := os.Stat(folianLogoPath); err == nil {
folianLogoContent, err := ioutil.ReadFile(folianLogoPath)
if err != nil {
return fmt.Errorf("failed to read Folian logo: %w", err)
}
if err := ioutil.WriteFile(filepath.Join(imagesPath, "folian.png"), folianLogoContent, 0644); err != nil {
return fmt.Errorf("failed to write Folian logo: %w", err)
}
} else {
fmt.Println("Warning: Folian logo not found")
}

// Process cover image if it exists
if book.CoverImage != "" {
fullPath := filepath.Join(basePath, book.CoverImage)
content, err := ioutil.ReadFile(fullPath)
if err != nil {
return fmt.Errorf("failed to read cover image %s: %w", book.CoverImage, err)
}

// Write the cover image
coverFilename := "cover" + filepath.Ext(book.CoverImage)
outputPath := filepath.Join(imagesPath, coverFilename)
if err := ioutil.WriteFile(outputPath, content, 0644); err != nil {
return fmt.Errorf("failed to write cover image: %w", err)
}

// Create titlepage.xhtml
titlePageXHTML := fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:svg="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xml:lang="en">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
    <meta name="calibre:cover" content="true"/>
    <title>Cover</title>
    <style type="text/css" title="override_css">
        @page {
            margin: 0pt;
            padding: 0pt;
        }
        html, body {
            height: 100%%;
            width: 100%%;
            margin: 0;
            padding: 0;
        }
        body {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        svg {
            max-width: 100%%;
            max-height: 100%%;
        }
    </style>
</head>
<body>
    <svg xmlns="http://www.w3.org/2000/svg"
         xmlns:xlink="http://www.w3.org/1999/xlink"
         version="1.1"
         viewBox="0 0 1038 1380"
         preserveAspectRatio="xMidYMid meet">
        <image width="1038" height="1380" xlink:href="images/%s"/>
    </svg>
</body>
</html>`, coverFilename)

if err := ioutil.WriteFile(filepath.Join(oebpsPath, "titlepage.xhtml"), []byte(titlePageXHTML), 0644); err != nil {
return fmt.Errorf("failed to create titlepage.xhtml: %w", err)
}

// Create jacket.html
jacketHTML := fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>
  <title>%s</title>
  <link href="styles/stylesheet.css" rel="stylesheet" type="text/css"/>
  <style>
  @page {
    margin: 0;
    padding: 0;
  }
  html, body {
    margin: 0;
    padding: 0;
    width: 100%%;
    height: 100%%;
    display: flex;
    justify-content: center;
    align-items: center;
    background: #f8f8f8;
    color: #333;
    font-family: serif;
  }
  .book-cover {
    width: 90vw;
    height: 90vh;
    max-width: 600px;
    background: white;
    padding: 60px 50px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    text-align: center;
  }
  .title {
    font-size: 32px;
    font-family: monospace;
    font-weight: 200;
    letter-spacing: 4px;
    margin: 20px 0 5px;
    line-height: 1.3;
  }
  .subtitle {
    font-size: 18px;
    font-family: serif;
    font-style: italic;
    color: #666;
    margin: 0 0 15px;
  }
  .author {
    font-size: 14px;
    font-family: serif;
    color: #555;
    margin-top: auto;
    padding-top: 60px;
    letter-spacing: 2px;
    text-transform: uppercase;
  }
  </style>
</head>
<body>
  <div class="book-cover">
    <div class="title">%s</div>
    <div class="author">%s</div>
  </div>
</body>
</html>`, book.Metadata.Title, book.Metadata.Title, book.Metadata.Creator)

if err := ioutil.WriteFile(filepath.Join(oebpsPath, "jacket.html"), []byte(jacketHTML), 0644); err != nil {
return fmt.Errorf("failed to create jacket.html: %w", err)
}
}

// Copy other images
for _, imagePath := range book.Images {
// Skip the cover image as we've already processed it
if imagePath == book.CoverImage {
continue
}

// Read the image file
fullPath := filepath.Join(basePath, imagePath)
content, err := ioutil.ReadFile(fullPath)
if err != nil {
return fmt.Errorf("failed to read image %s: %w", imagePath, err)
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

for i, chapter := range book.Chapters {
// Process chapter content
processedContent := r.processChapterContent(chapter.Content, i+1, book.Metadata.Title)

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
func (r *Restructurer) processChapterContent(content string, chapterNum int, bookTitle string) string {
// Fix relative paths
content = regexp.MustCompile(`href="([^"]+\.css)"`).ReplaceAllString(content, `href="../styles/$1"`)
content = regexp.MustCompile(`href="([^"]+\.xhtml)"`).ReplaceAllString(content, `href="../chapters/$1"`)
content = regexp.MustCompile(`src="([^"]+\.(jpg|jpeg|png|gif))"`).ReplaceAllString(content, `src="../images/$1"`)

// Add chapter class to body
content = regexp.MustCompile(`<body[^>]*>`).ReplaceAllString(content, `<body class="chapter">`)

// Add chapter number to title if not present
titleMatch := regexp.MustCompile(`<title>([^<]+)</title>`).FindStringSubmatch(content)
if len(titleMatch) > 1 && !strings.Contains(titleMatch[1], fmt.Sprintf("Chapter %d", chapterNum)) {
newTitle := fmt.Sprintf("Chapter %d: %s", chapterNum, bookTitle)
content = regexp.MustCompile(`<title>[^<]+</title>`).ReplaceAllString(content, fmt.Sprintf(`<title>%s</title>`, newTitle))
}

// Remove Calibre-specific elements
content = regexp.MustCompile(`<div class="calibre[^"]*"[^>]*>|</div>`).ReplaceAllString(content, "")
content = regexp.MustCompile(`class="calibre[^"]*"`).ReplaceAllString(content, "")

return content
}

// createContentOPF creates the content.opf file
func (r *Restructurer) createContentOPF(book *parser.Book, oebpsPath string) error {
// Start building the OPF content
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
    <meta name="cover" content="cover-image"/>
  </metadata>
  <manifest>
`,
book.Metadata.Title,
book.Metadata.Creator,
book.Metadata.Language,
book.Metadata.Identifier,
book.Metadata.Publisher,
book.Metadata.Description,
book.Metadata.Date)

// Add items to manifest
manifestItems := []string{
`    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>`,
}

// Add titlepage and jacket if cover exists
hasCover := book.CoverImage != ""
if hasCover {
manifestItems = append(manifestItems, `    <item id="titlepage" href="titlepage.xhtml" media-type="application/xhtml+xml" properties="svg"/>`)
manifestItems = append(manifestItems, `    <item id="jacket" href="jacket.html" media-type="application/xhtml+xml"/>`)
manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="cover-image" href="images/cover%s" media-type="image/%s" properties="cover-image"/>`,
filepath.Ext(book.CoverImage),
strings.TrimPrefix(filepath.Ext(book.CoverImage), ".")))

// Add Folian logo
folianLogoPath := filepath.Join("/Users/4azy/lune/code/pub/1984-custom-fixed-dir/OEBPS/images/folian.png")
if _, err := os.Stat(folianLogoPath); err == nil {
manifestItems = append(manifestItems, `    <item id="folian-logo" href="images/folian.png" media-type="image/png"/>`)
}
}

// Add stylesheets
manifestItems = append(manifestItems, `    <item id="stylesheet" href="styles/stylesheet.css" media-type="text/css"/>`)
for i, stylesheet := range book.Stylesheets {
manifestItems = append(manifestItems, fmt.Sprintf(`    <item id="style%d" href="styles/%s" media-type="text/css"/>`, i+1, filepath.Base(stylesheet)))
}

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

// Add chapters to spine
for i := range book.Chapters {
spineItems = append(spineItems, fmt.Sprintf(`    <itemref idref="chapter%d"/>`, i+1))
}

// Add spine items to OPF
opfContent += strings.Join(spineItems, "\n") + "\n  </spine>\n</package>"

// Write the OPF file
return ioutil.WriteFile(filepath.Join(oebpsPath, "content.opf"), []byte(opfContent), 0644)
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
      <content src="jacket.html"/>
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
