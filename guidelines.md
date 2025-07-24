# EPUB Processing Guidelines for Folian-Parser

## Input Requirements
- **Required Format**: EPUB files only (.epub extension)
- **Content Type**: Text-based ebooks work best (novels, non-fiction, technical books)
- **Avoid**: Scanned books, image-heavy content, or PDFs converted to EPUB without proper text extraction
- **Quality**: EPUBs with proper HTML structure and metadata produce optimal results

## Pre-Processing Setup
1. **Format Conversion** (if needed):
   - If your source is not EPUB format (PDF, MOBI, AZW3, etc.), convert it using Calibre first
   - In Calibre: Add book → Select book → Convert books → Choose EPUB as output format
   - Ensure "Convert to EPUB" settings preserve text structure and metadata

2. **Tool Installation**:
   - Follow the installation instructions in the README.md
   - Verify the format directory is created automatically with all required assets
   - Test with: `./folian-parser -v` to confirm installation

## Processing Workflow
1. **Basic Processing**:
   ```bash
   ./folian-parser -i your-book.epub -o new-name.epub -d
   ```

## Post-Processing Quality Control
1. **Open in Calibre Editor**:
   - Right-click the output EPUB in Calibre → Edit book
   - Use Tools → Check book to identify any structural issues
   - Fix common issues: broken links, missing images, malformed HTML
   - Pay attention to: chapter navigation, table of contents, font references

2. **Manual Fixes** (if needed):
   - Correct any flagged HTML validation errors
   - Verify image paths and font references are working
   - Check chapter titles and navigation structure
   - Ensure proper EPUB 3.0 metadata compliance

## Final Verification
1. **Test in Multiple Readers**:
   - Apple Books (iOS/macOS) for responsive design testing
   - Adobe Digital Editions for EPUB standard compliance
   - Kindle app (if converting to MOBI later)

2. **Verify Key Features**:
   - Professional jacket page displays correctly
   - Chapter navigation works smoothly
   - Fonts render properly (Jura font integration)
   - Images and logo appear correctly
   - Responsive design adapts to different screen sizes

## Expected Results
- Clean, professional EPUB 3.0 format
- Consolidated chapters with intelligent organization
- Responsive jacket design with Folian branding
- Cross-device compatibility
- Improved typography and readability

## Common Issues & Solutions

### Font Issues
- **Font not displaying**: Check that `fonts/jura.ttf` path is correct in CSS
- **Font rendering problems**: Ensure the Jura font file is properly embedded
- **Inconsistent typography**: Verify all templates reference the correct font paths
- "Incorrect font type": check the font file type (TTF, OTF, WOFF, WOFF2) and update the `media-type` in the `content.opf` file accordingly

```bash
<item id="jura-font" href="fonts/jura.ttf" media-type="application/vnd.ms-opentype"/>
```

### Image Issues
- **Images missing**: Verify image paths use `images/` directory structure
- **Cover image not showing**: Check that cover image extension matches (jpg/jpeg/png)
- **Logo not appearing**: Ensure `folian.png` is in the correct `images/` directory

### Processing Issues
- **Chapter consolidation problems**: Use `-enhanced` flag for better results
- **Validation errors**: Run with `-d` flag to see detailed processing info
- **Large file processing**: Be patient with files >50MB, may take 2-3 minutes

### Navigation Issues
- **Broken table of contents**: Check that chapter titles are properly extracted
- **Missing navigation**: Ensure EPUB 3.0 nav.xhtml is generated correctly
- **Chapter order problems**: Verify original EPUB has proper spine order

## Performance Considerations

### File Size Guidelines
- **Small files (<5MB)**: Process in under 30 seconds
- **Medium files (5-20MB)**: Process in 1-2 minutes
- **Large files (20-50MB)**: Process in 2-3 minutes
- **Very large files (>50MB)**: May require 5+ minutes, consider splitting

### Content Complexity
- **Simple novels**: Minimal manual cleanup needed
- **Technical books**: May need manual chapter title adjustment
- **Complex formatting**: Some post-processing cleanup recommended
- **Many chapters (>100)**: Enhanced mode will consolidate intelligently

## Calibre Integration

### Import Workflow
1. **Add processed EPUB**: Import the enhanced EPUB into Calibre library
2. **Metadata sync**: Update book metadata in Calibre if needed
3. **Format conversion**: Convert to other formats (MOBI, AZW3) from the enhanced EPUB
4. **Quality check**: Use Calibre's built-in validation tools

### Success Indicators
- ✅ File size optimized (usually 10-20% smaller)
- ✅ Chapter count reduced from 50+ to 15-25 meaningful chapters
- ✅ Single consolidated stylesheet
- ✅ Professional jacket page with proper branding
- ✅ EPUB 3.0 validation passes
- ✅ Cross-device compatibility confirmed
- ✅ Improved typography and readability
- ✅ This tool can make the process 80% correctly if having a good epub input, the rest 20% needs to be done manually.