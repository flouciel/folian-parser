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
        <image width="1038" height="1380" xlink:href="images/cover.jpg"/>
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
  <style>
  @font-face {
    font-family: Jura;
    src: url(../fonts/jura.ttf);
  }
  @page {
    margin: 0;
    padding: 0;
  }
  html, body {
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    background: #f8f8f8;
    color: #333;
    font-family: serif;
    overflow: hidden;
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
    position: relative;
  }
  .logo-container {
    margin: 15px auto 50px;
    width: 100px;
    height: 100px;
    display: flex;
    justify-content: center;
    align-items: center;
    position: relative;
  }
  .logo {
    max-width: 70%;
    max-height: 70%;
  }
  .title {
    font-size: 48px;
    font-family: "Jura", monospace;
    font-weight: 200;
    letter-spacing: 4px;
    margin: 20px 0 15px;
    line-height: 1.3;
  }
  .subtitle {
    font-size: 16px;
    color: #777;
    font-weight: 300;
    margin-bottom: 60px;
    font-style: italic;
    letter-spacing: 1px;
  }
  .author {
    font-size: 32px;
    font-family: "Jura", serif;
    color: #555;
    margin-top: auto;
    padding-top: 60px;
    letter-spacing: 2px;
    text-transform: uppercase;
    position: relative;
  }
  .footer {
    font-size: 11px;
    color: #999;
    margin-top: 20px;
    letter-spacing: 1px;
  }
  </style>
</head>
<body>
  <div class="book-cover">
    <div class="logo-container">
      <img src="../images/folian.png" alt="Folian Logo" class="logo"/>
    </div>
    <div class="title">{{BOOK_TITLE}}</div>
    <div class="subtitle">{{BOOK_SUBTITLE}}</div>
    <div class="author">{{BOOK_AUTHOR}}</div>
    <div class="footer">ebook@folian</div>
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
echo "Note: You should replace the placeholder files with actual font and logo files."
echo "Usage: folian-parser -i input.epub -format $FORMAT_DIR"

# Make the script executable
chmod +x create-format-dir.sh
