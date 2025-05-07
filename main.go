package main

import (
	"os"

	"github.com/flouciel/folian-parser/cmd/folian-parser"
)

func main() {
	// Just call the main function from the cmd/folian-parser package
	folianparser.Main()
	os.Exit(0)
}
