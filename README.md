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
go get github.com/flouciel/folian-parser@latest
```

Or clone the repository and build it manually:

```bash
git clone https://github.com/flouciel/folian-parser.git
cd folian-parser
go build -o folian-parser
```

## Usage

```bash
folian-parser -i input.epub -o output.epub
```

If the output path is not provided, the tool will generate one based on the input path:

```bash
folian-parser -i input.epub
# Output will be saved as input-fixed.epub
```

You can specify a custom format directory containing templates and assets:

```bash
folian-parser -i input.epub -format /path/to/format/directory
```

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

- Automatically extracts metadata from the EPUB file
- Restructures the content according to a standardized format
- Removes publisher-specific formatting and classes
- Creates a properly formatted titlepage and jacket using templates
- Generates an EPUB3-compliant navigation document (nav.xhtml)
- Applies consistent styling using a single stylesheet
- Organizes content into a standard directory structure
- Uses template variables for easy customization

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

## Quick Start Guide

Follow these steps to get started with Folian Parser:

1. **Build the tool**:
   ```bash
   go build -o folian-parser
   ```

2. **Create the format directory**:
   ```bash
   ./create-format-dir.sh
   ```

3. **Replace placeholder files**:
   - Replace `format/jura.ttf` with the actual Jura font file
   - Replace `format/folian.png` with your logo

4. **Process an EPUB file**:
   ```bash
   ./folian-parser -i your-book.epub -o your-book-fixed.epub
   ```

5. **Verify the output**:
   - Check the generated EPUB file in your e-reader or EPUB viewer
   - Ensure the styling, navigation, and structure are as expected

## License

MIT
