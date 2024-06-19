package main

import (
	"flag"
	"fmt"
	"jch-metadata/internal/parser"
	"jch-metadata/internal/parser/flac"
	"jch-metadata/internal/parser/jpeg"
	"jch-metadata/internal/parser/mkv"
	"jch-metadata/internal/parser/png"
	"os"
)

var actionArg string
var inputFilename string

var parsers = []parser.Parser{
	flac.Parser,
	mkv.Parser,
	png.Parser,
	jpeg.Parser,
}

func main() {
	flag.StringVar(&inputFilename, "f", "", "Input filename")
	flag.StringVar(&actionArg, "a", "show", "Action to perform: show, clear, extract")
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
	fmt.Printf("Opening file %s\n", inputFilename)
	file, err := os.OpenFile(inputFilename, fileFlag, 644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	parsed, err := parser.StartParsing(parsers, file, action, 0)
	if err != nil {
		fmt.Println("Error handling file:", err)
		return
	}
	if !parsed {
		fmt.Println("Invalid file format.  The following formats are supported:")
		for _, p := range parsers {
			fmt.Print(p.Name, "   ")
		}
		fmt.Println()
	}
	err = file.Close()
	if err != nil {
		fmt.Println("Failed to close file:", err)
	}
}
