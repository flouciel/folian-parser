# Folian Parser

A tool for restructuring EPUB files according to a standardized format.

## Overview

Folian Parser is a command-line tool that restructures EPUB files by:

- Organizing content into a standard EPUB structure
- Moving stylesheets to the `@format` folder
- Importing the Jura font
- Replacing cover images
- Removing Calibre-specific classes from HTML/XHTML files
- Creating properly formatted chapter files

## Installation

```bash
go install github.com/flouciel/folian-parser@latest
```

Or clone the repository and build it manually:

```bash
git clone https://github.com/flouciel/folian-parser.git
cd folian-parser
go build -o folian-parser ./cmd/folian-parser
```

## Usage

```bash
./folian-parser -input input.epub -output output.epub
```

If the output path is not provided, the tool will generate one based on the input path:

```bash
./folian-parser -input input.epub
# Output will be saved as input-fixed.epub
```

## Features

- Automatically extracts metadata from the EPUB file
- Restructures the content according to a standardized format
- Removes Calibre-specific formatting
- Creates a properly formatted titlepage and jacket
- Organizes content into a standard directory structure

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
    ├── jacket.html
    ├── styles
    │   └── stylesheet.css
    ├── titlepage.xhtml
    └── toc.ncx
```

## License

MIT
