#!/bin/bash

# Create a format directory with the necessary files for Folian Parser

# Default format directory path
FORMAT_DIR="format"

# Check if a custom path was provided
if [ $# -eq 1 ]; then
    FORMAT_DIR="$1"
fi

echo "Creating format directory at: $FORMAT_DIR"

# Create the format directory if it doesn't exist
mkdir -p "$FORMAT_DIR"

# Create a basic titlepage.xhtml template
cat > "$FORMAT_DIR/titlepage.xhtml" << 'EOF'
<?xml version='1.0' encoding='utf-8'?>
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
            height: 100%;
            width: 100%;
            margin: 0;
            padding: 0;
        }
        body {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        svg {
            max-width: 100%;
            max-height: 100%;
        }
    </style>
</head>
<body>
    <svg xmlns="http://www.w3.org/2000/svg"
         xmlns:xlink="http://www.w3.org/1999/xlink"
         version="1.1"
         viewBox="0 0 1038 1380"
         preserveAspectRatio="xMidYMid meet">
        <image width="1038" height="1380" xlink:href="images/cover.jpeg"/>
    </svg>
</body>
</html>
EOF

# Create a basic jacket.xhtml template
cat > "$FORMAT_DIR/jacket.xhtml" << 'EOF'
<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>
  <title>{{BOOK_TITLE}}</title>
  <link href="styles/stylesheet.css" rel="stylesheet" type="text/css"/>
    <style type="text/css">
      @font-face {
        font-family: "Jura";
        src: url("fonts/jura.ttf");
        font-weight: normal;
        font-style: normal;
      }

      body {
        margin: 5%;
        padding: 0;
        background-color: #ffffff;
        font-family: "Jura", serif;
        color: #333;
        text-align: center;
      }

      .book-cover {
        margin: 0 auto;
        max-width: 100%;
        text-align: center;
        padding: 2em 0;
        min-height: 90vh; /* Make the cover take up most of the viewport height */
        display: flex;
        flex-direction: column;
        justify-content: space-between;
      }

      .logo {
        width: 100px;
        height: auto;
        max-width: 100%;
        margin-bottom: 2em;
      }

      .title {
        font-size: 2em;
        font-family: "Jura", sans-serif;
        text-align: center;
        letter-spacing: 2px;
        margin: 1em 0 0.5em;
        font-weight: bold;
      }

      .subtitle {
        font-size: 1.2em;
        color: #555;
        font-style: italic;
        letter-spacing: 1px;
        margin-bottom: 1.5em;
      }

      .author {
        font-size: 1.2em;
        color: #555;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin: 0;
        padding-bottom: 0.3em;
      }

      .content {
        margin: 2em 0;
        flex-grow: 1;
      }

      .bottom-content {
        margin-top: auto; /* Push to the bottom when using flexbox */
        padding-bottom: 1em; /* Add some padding at the bottom */
      }

      .footer {
        font-size: 0.8em;
        color: #777;
        letter-spacing: 1px;
        text-align: center;
        margin-top: 0.5em; /* Reduced space between author and footer */
      }

      /* For e-readers that don't support flexbox */
      @supports not (display: flex) {
        .book-cover {
          position: relative;
          min-height: 90vh;
        }

        .bottom-content {
          position: absolute;
          bottom: 1em;
          left: 0;
          right: 0;
        }

        .content {
          margin-bottom: 8em; /* Ensure content doesn't overlap with bottom content in fallback mode */
        }
      }

      /* Media queries for better responsiveness */
      @media screen and (max-width: 600px) {
        .title {
          font-size: 1.5em;
        }

        .subtitle, .author {
          font-size: 1em;
        }

        .logo {
          width: 80px;
        }
      }
    </style>
  </head>
<body class="cover">
  <div class="book-cover">
    <div class="content">
      <div class="logo-container">
        <img src="../images/folian.png" alt="Folian Logo" class="logo"/>
      </div>
      <h1 class="title">{{BOOK_TITLE}}</h1>
      <div class="subtitle">{{BOOK_SUBTITLE}}</div>
    </div>
    <div class="bottom-content">
      <div class="author">{{BOOK_AUTHOR}}</div>
      <div class="footer">ebook@folian</div>
    </div>
  </div>
  </body>
</html>
EOF

# Create a placeholder for the Folian logo
echo "Please replace this with an actual PNG logo file" > "$FORMAT_DIR/folian.png"

# Create a nav.xhtml template
cat > "$FORMAT_DIR/nav.xhtml" << 'EOF'
<?xml version='1.0' encoding='utf-8'?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>{{BOOK_TITLE}}</title>
  <link href="styles/stylesheet.css" rel="stylesheet" type="text/css"/>
  <style type="text/css">
  li {
    list-style-type: none;
    padding-left: 2em;
    margin-left: 0;
  }
  a {
    text-decoration: none;
  }
  a:hover {
    color: red;
  }
  </style>
</head>
<body>
  <nav epub:type="toc" id="toc">
    <h2>Table of Contents</h2>
    <ol>
      {{TOC_ENTRIES}}
    </ol>
  </nav>
</body>
</html>
EOF

# Create a placeholder for the Jura font
echo "Please replace this with the actual Jura TTF font file" > "$FORMAT_DIR/jura.ttf"

# Create a default stylesheet.css
cat > "$FORMAT_DIR/stylesheet.css" << 'EOF'
@page {
  margin-bottom: 5pt;
  margin-top: 5pt;
}
@font-face {
  font-family: Jura;
  src: url(../fonts/jura.ttf);
}
/* Styles for Folian books */
h1 {
  text-align: center;
  font-size: 2.5em;
  font-family: "Jura", serif;
  margin: 3em auto 0 auto;
}
h2 {
  text-align: left;
  text-indent: 5%;
  font-size: 1.7em;
  font-family: "Jura", serif;
  margin: 2em auto 1em auto;
}
h3 {
  text-align: center;
  font-size: 1.5em;
  font-family: serif;
  margin: 1em auto 1em auto;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
}
th, td {
  padding: 0.5em;
  border: 1px solid #ccc;
  text-align: left;
  vertical-align: top;
}
th {
  background-color: #f0f0f0;
}
p {
  text-indent: 5%;
  text-align: justify;
  line-height: 1.5;
}
p.nonindent {
  text-indent: 0;
}
p.sans-serif {
  font-family: sans-serif;
}
p.center {
  text-indent: 0;
  text-align: center;
}
p.right {
  text-indent: 0;
  text-align: right;
}
p.poem {
  text-indent: 0;
  margin-left: 15%;
  font-size: 0.9em;
}
p.serif {
  font-family: serif;
}
hr {
  border: 1px solid;
  width: 40%;
  border-radius: 1px;
  margin-top: 2em;
  margin-bottom: 2em;
}
blockquote {
  margin: 2em 0 2em 5%;
}
blockquote > p {
  text-indent: 0;
  font-family: monospace;
}
p.pagebreak {
  display: block;
  page-break-after: always;
}
p.title {
  font-size: 1.7em;
}
p.author {
  font-size: 1.5em;
}
p.series {
  font-size: 0.9em;
}
p.quote {
  margin-left: 5%;
}
div.box {
  margin: 2em 5% 2em 5%;
}
div.box > p {
  font-family: sans-serif;
  font-size: 0.8em;
}
div.computer > p {
  font-family: monospace;
  font-size: 0.9em;
}
div.info {
  margin-top: 2em;
}
div.info p {
  text-indent: 0;
  text-align: center;
  font-family: serif;
  line-height: 1.5;
}
sup {
  font-size: 80%;
}
a {
  text-decoration: none;
}
a:hover {
  color: red;
}
td {
  font-family: monospace;
  font-size: 0.9em;
}
ul {
  list-style-type: disc;
  /* classic bullet point */
  padding-left: 1.5rem;
  /* spacing from the left */
  margin-bottom: 1rem;
}
li {
  margin-bottom: 0.5rem;
  /* space between list items */
  font-size: 1rem;
  color: #333;
}
EOF

echo "Format directory created successfully at: $FORMAT_DIR"
echo ""
echo "Next steps:"
echo "1. Replace placeholder files with actual files:"
echo "   - Replace $FORMAT_DIR/jura.ttf with the actual Jura font file"
echo "   - Replace $FORMAT_DIR/folian.png with your actual logo"
echo ""
echo "2. Run the Folian Parser tool:"
echo "   ./folian-parser -i your-book.epub -o your-book-fixed.epub"
echo ""
echo "For custom format directory:"
echo "   ./folian-parser -i your-book.epub -o your-book-fixed.epub -f $FORMAT_DIR"

# Make the script executable
chmod +x create-format-dir.sh
