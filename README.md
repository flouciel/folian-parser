# Folian Parser

A tool for restructuring EPUB files according to a standardized format.

## Overview

Folian Parser is a command-line tool that restructures EPUB files by:

- Organizing content into a standard EPUB structure
- Applying consistent styling using templates from a format directory
- Creating a standardized navigation system (toc.ncx and nav.xhtml)
- Importing the Jura font for consistent typography
- Creating professional title and jacket pages
- Removing publisher-specific classes from HTML/XHTML files
- Creating properly formatted chapter files

## Installation

```bash
go install github.com/flouciel/folian-parser@latest
```

Or clone the repository and build it manually:

```bash
git clone https://github.com/flouciel/folian-parser.git
cd folian-parser
go build -o folian-parser
```

## Usage

```bash
folian-parser -i input.epub -o output.epub [-f path/to/format/directory] [-v] [-d]
```

### Command-line Options

- `-i`: Input EPUB file path (required)
- `-o`: Output EPUB file path (optional, defaults to input-fixed.epub)
- `-f`: Path to the format directory (optional, defaults to "format")
- `-v`: Display version information and exit
- `-d`: Enable debug output to verify file creation
- `-u`: Check for updates and update if a newer version is available
- `-a`: Analyze EPUB structure without processing
- `-validate`: Validate EPUB structure only
- `-enhanced`: Use enhanced processing with intelligent chapter consolidation
- `-compare`: Compare two EPUB files and show differences

If the output path is not provided, the tool will generate one based on the input path:

```bash
folian-parser -i input.epub
# Output will be saved as input-fixed.epub
```

You can specify a custom format directory containing templates and assets:

```bash
folian-parser -i input.epub -f /path/to/format/directory
```

### Enhanced Processing & Analysis (NEW)

```bash
# Enhanced processing with intelligent chapter consolidation
folian-parser -i input.epub -enhanced

# Analyze EPUB structure without processing
folian-parser -i input.epub -a

# Validate EPUB structure only
folian-parser -i input.epub -validate

# Enhanced processing with debug output
folian-parser -i input.epub -o output.epub -enhanced -d

# Compare original and enhanced versions
folian-parser -i original.epub -compare enhanced.epub
```

### Advanced Usage

For better results with complex EPUBs, use the enhanced processing script:

```bash
# Basic enhanced processing
./enhance-epub.sh -i input.epub

# With custom output and format directory
./enhance-epub.sh -i input.epub -o enhanced-output.epub -f custom-format

# Analyze EPUB structure only
./enhance-epub.sh -a -i input.epub

# Validate EPUB only
./enhance-epub.sh -v -i input.epub
```

The enhanced script provides:
- **Pre-processing validation** of input EPUB files
- **Structure analysis** to understand content organization
- **Post-processing validation** of output files
- **Detailed logging** of the enhancement process

### Format Directory

The format directory contains templates and assets used to standardize the EPUB files. It should contain the following files:

- `stylesheet.css` - CSS stylesheet for the EPUB content
- `titlepage.xhtml` - Template for the title page with `{{BOOK_TITLE}}` placeholder
- `jacket.xhtml` - Template for the jacket page with `{{BOOK_TITLE}}`, `{{BOOK_SUBTITLE}}`, and `{{BOOK_AUTHOR}}` placeholders
- `nav.xhtml` - Template for the navigation document with `{{BOOK_TITLE}}` and `{{TOC_ENTRIES}}` placeholders
- `jura.ttf` - The Jura font used in the EPUB
- `folian.png` - Folian logo image

To create the format directory, use the included script:

```bash
./create-format-dir.sh [path/to/format/directory]
```

If no path is provided, it will create a `format` directory in the current location.

After creating the format directory, you should:

1. Replace the placeholder `jura.ttf` with the actual Jura font file
2. Replace the placeholder `folian.png` with your actual logo
3. Customize the templates and stylesheet as needed

The templates use placeholders that will be replaced with actual content from the EPUB:

- `{{BOOK_TITLE}}` - The title of the book
- `{{BOOK_SUBTITLE}}` - The subtitle (or a shortened description)
- `{{BOOK_AUTHOR}}` - The author's name
- `{{TOC_ENTRIES}}` - Table of contents entries (for nav.xhtml)

## Features

### 🚀 **Enhanced Processing (NEW)**
- **Intelligent Chapter Consolidation**: Automatically merges small chapters and removes navigation duplicates
- **Advanced HTML Parsing**: Uses proper DOM parsing instead of regex for better accuracy
- **EPUB 3.0 Compliance**: Generates modern EPUB 3.0 format with enhanced metadata
- **Responsive Design**: Improved CSS with mobile and print optimizations
- **Content Analysis**: Smart detection of table of contents and navigation pages

### 📚 **Core Features**
- **Standardized Structure**: Organizes EPUB content into a clean, consistent structure
- **Template-Based Styling**: Uses customizable templates for title pages, jackets, and navigation
- **Calibre Cleanup**: Removes publisher-specific classes and styling artifacts
- **Font Integration**: Includes the Jura font for consistent typography
- **Professional Layout**: Creates polished title and jacket pages with logo integration
- **Navigation Enhancement**: Generates proper EPUB3 navigation documents
- **Batch Processing**: Can process multiple files efficiently

### 🎨 **Quality Improvements**
- **Enhanced Typography**: Better font hierarchy and spacing
- **Accessibility Support**: High contrast mode and reduced motion support
- **Error Reduction**: Proper HTML entity encoding and validation
- **Image Optimization**: Smart image path fixing and optimization

## Directory Structure

The restructured EPUB will have the following structure:

```
├── META-INF
│   └── container.xml
├── mimetype
└── OEBPS
    ├── chapters
    │   ├── chapter_001.xhtml
    │   ├── chapter_002.xhtml
    │   └── ...
    ├── content.opf
    ├── fonts
    │   └── jura.ttf
    ├── images
    │   ├── cover.jpg
    │   ├── folian.png
    │   └── ...
    ├── jacket.xhtml
    ├── nav.xhtml
    ├── styles
    │   └── stylesheet.css
    ├── titlepage.xhtml
    └── toc.ncx
```

## Workflow

The Folian Parser workflow consists of three main steps:

1. **Create the format directory** with templates and assets
2. **Replace placeholder files** with actual font and logo files
3. **Run the tool** to process EPUB files

### Step 1: Create the Format Directory

Run the included script to create the format directory:

```bash
./create-format-dir.sh
```

This will create a `format` directory in the current location with all necessary template files.

For a custom location:

```bash
./create-format-dir.sh /path/to/custom/format
```

### Step 2: Replace Placeholder Files

The script creates placeholder files that you should replace:

- Replace `format/jura.ttf` with the actual Jura font file
- Replace `format/folian.png` with your actual logo

You can also customize the templates and stylesheet to match your requirements.

### Step 3: Run the Tool

Process an EPUB file using the format directory:

```bash
./folian-parser -i your-book.epub -o your-book-fixed.epub
```

For a custom format directory:

```bash
./folian-parser -i your-book.epub -o your-book-fixed.epub -f /path/to/custom/format
```

To verify that the jacket.xhtml and nav.xhtml files are correctly created, use the debug flag:

```bash
./folian-parser -i your-book.epub -o your-book-fixed.epub -d
```

### Step 4: Verify the Output

- Check the generated EPUB file in your e-reader or EPUB viewer
- Ensure the styling, navigation, and structure are as expected
- Verify that the templates have been properly applied

## License

MIT
