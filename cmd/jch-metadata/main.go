package main

import (
	"flag"
	"fmt"
	"jch-metadata/internal/parser"
	"jch-metadata/internal/parser/flac"
	"jch-metadata/internal/parser/mkv"
	"os"
)

var actionArg string
var inputFilename string

func main() {
	parsers := []parser.Parser{
		mkv.Parser,
		flac.Parser,
	}
	flag.StringVar(&inputFilename, "f", "", "Input filename")
	flag.StringVar(&actionArg, "a", "show", "Action to perform: show, clear")
	flag.Parse()
	if inputFilename == "" {
		fmt.Println("Invalid input filename")
		flag.PrintDefaults()
		return
	}
	action, err := parser.ConvertAction(actionArg)
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		return
	}
	fileFlag := os.O_RDONLY
	if action == parser.ClearAction {
		fileFlag = os.O_RDWR
	}
	file, err := os.OpenFile(inputFilename, fileFlag, 644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	parsed := false
	for _, p := range parsers {
		supported, err := p.Support(file)
		if err != nil {
			fmt.Printf("Error checking support from parser %s: %s", p.Name, err)
			continue
		}
		if !supported {
			continue
		}
		fmt.Printf("Processing file type %s...\n", p.Name)
		fmt.Println()
		err = p.Handle(file, action)
		if err != nil {
			fmt.Println("Error handling file:", err)
			return
		}
		parsed = true
		break
	}
	if !parsed {
		fmt.Println("Invalid file format.  The following formats are supported:")
		for _, p := range parsers {
			fmt.Print(p.Name, "   ")
		}
	}
	err = file.Close()
	if err != nil {
		fmt.Println("Failed to close file:", err)
	}
}
