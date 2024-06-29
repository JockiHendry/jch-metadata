package main

import (
	"flag"
	"fmt"
	"io/fs"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"jch-metadata/internal/parser/elf"
	"jch-metadata/internal/parser/flac"
	"jch-metadata/internal/parser/jpeg"
	"jch-metadata/internal/parser/mkv"
	"jch-metadata/internal/parser/mp4"
	"jch-metadata/internal/parser/png"
	"jch-metadata/internal/parser/webp"
	"os"
	"path/filepath"
)

var actionArg string
var inputFilename string

var parsers = []parser.Parser{
	flac.Parser,
	jpeg.Parser,
	png.Parser,
	webp.Parser,
	mkv.Parser,
	mp4.Parser,
	elf.Parser,
}

func parseFile(fileName string, action parser.Action) {
	fileFlag := os.O_RDONLY
	if action == parser.ClearAction {
		fileFlag = os.O_RDWR
	}
	fmt.Printf("Opening file \033[7m%s\033[27m\n", fileName)
	file, err := os.OpenFile(fileName, fileFlag, 644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file:", err)
		}
	}(file)
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error retrieving file stat:", err)
	}
	parsed, err := parser.StartParsing(parsers, file, action, 0, fileInfo.Size())
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
}

func parseBatchedFile(fileName string, action parser.Action, fileSize int64) bool {
	fileFlag := os.O_RDONLY
	if action == parser.ClearAction {
		fileFlag = os.O_RDWR
	}
	file, err := os.OpenFile(fileName, fileFlag, 644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file:", err)
		}
	}(file)
	supported := false
	for _, p := range parsers {
		supported, err = p.Support(file, 0, fileSize)
		if err != nil {
			continue
		}
		if supported {
			break
		}
	}
	if !supported {
		return false
	}
	parseFile(fileName, action)
	fmt.Println()
	fmt.Println()
	return true
}

func main() {
	output.Setup()
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
	fileStat, err := os.Stat(inputFilename)
	if err != nil {
		fmt.Printf("Error retrieving information for %s: %s\n", inputFilename, err)
		return
	}
	if fileStat.Mode().IsDir() {
		supportedFiles := 0
		totalFiles := 0
		err = filepath.WalkDir(inputFilename, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Encountered error while processing file: %s\n", err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			fileInfo, err := d.Info()
			if err != nil {
				fmt.Printf("Encountered error while getting file info: %s\n", err)
				return nil
			}
			supported := parseBatchedFile(path, action, fileInfo.Size())
			if supported {
				supportedFiles += 1
			}
			totalFiles += 1
			return nil
		})
		if err != nil {
			fmt.Printf("Encountered error while listing directory: %s\n", err)
		}
		fmt.Printf("Recursively parsed %d files out of %d files in folder %s\n", supportedFiles, totalFiles, inputFilename)
	} else {
		parseFile(inputFilename, action)
	}
}
