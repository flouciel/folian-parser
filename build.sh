#!/bin/bash

# Build the folian-parser tool
go build -o folian-parser ./cmd/folian-parser

echo "Build complete. The folian-parser tool is ready to use."
echo "Usage: ./folian-parser -input input.epub -output output.epub"
