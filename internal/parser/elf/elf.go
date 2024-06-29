package elf

import (
	"bytes"
	"debug/buildinfo"
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"fmt"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"math/rand"
	"os"
	"sort"
)

var Parser = parser.Parser{
	Name:      "ELF (Executable)",
	Container: false,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsELF(file, startOffset)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		if action == parser.ShowAction {
			err := ShowMetadata(file)
			if err != nil {
				return err
			}
		} else if action == parser.ClearAction {
			err := ClearMetadata(file)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func IsELF(file *os.File, startOffset int64) (bool, error) {
	magicBytes := make([]byte, 4)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0x7F, 0x45, 0x4C, 0x46}), nil
}

func ShowMetadata(file *os.File) error {
	dwarfFiles, err := GetDWARFFiles(file)
	if err != nil {
		return err
	}
	if dwarfFiles != nil {
		output.PrintHeader(false, "DWARF Source Files")
		for _, f := range dwarfFiles {
			output.Println(false, f)
		}
		output.Println(false)
	}
	pclntabFiles, err := GetPclntabFiles(file)
	if err != nil {
		return err
	}
	if pclntabFiles != nil {
		output.PrintHeader(false, "Go Symbol Table Files")
		for _, f := range pclntabFiles {
			output.Println(false, f)
		}
		output.Println(false)
	}
	buildId, err := GetBuildId(file)
	if err != nil {
		return err
	}
	if buildId != "" {
		output.PrintHeader(false, "Go Build ID")
		output.Println(false, buildId)
		output.Println(false)
	}
	buildInfo := GetBuildInfo(file)
	if buildInfo != nil {
		output.PrintHeader(false, "Go Build Info")
		output.Println(false, buildInfo)
	}
	return nil
}

func GetDWARFFiles(file *os.File) ([]string, error) {
	var result = make(map[string]bool)
	elfFile, err := elf.NewFile(file)
	if err != nil {
		return nil, fmt.Errorf("error opening ELF file: %w", err)
	}
	d, err := elfFile.DWARF()
	if err != nil {
		return nil, nil
	}
	reader := d.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			return nil, fmt.Errorf("error reading next DWARF entry: %w", err)
		} else if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			reader.SkipChildren()
			continue
		}
		lineReader, err := d.LineReader(entry)
		if err != nil {
			return nil, fmt.Errorf("error getting line reader: %w", err)
		} else if lineReader == nil {
			continue
		}

		files := lineReader.Files()
		for _, f := range files {
			if f == nil {
				continue
			}
			_, found := result[f.Name]
			if !found {
				result[f.Name] = true
			}
		}
	}
	files := make([]string, 0, len(result))
	for k := range result {
		files = append(files, k)
	}
	sort.Strings(files)
	return files, nil
}

func GetPclntabFiles(file *os.File) ([]string, error) {
	elfFile, err := elf.NewFile(file)
	if err != nil {
		return nil, fmt.Errorf("error opening ELF file: %w", err)
	}
	goSymTabSection := elfFile.Section(".gosymtab")
	if goSymTabSection == nil {
		return nil, nil
	}
	goSymTabData, err := goSymTabSection.Data()
	if err != nil {
		return nil, fmt.Errorf("error reading .gosymtab data: %w", err)
	}
	pclntab := elfFile.Section(".gopclntab")
	if pclntab == nil {
		return nil, fmt.Errorf(".gopclntab not found")
	}
	pclntabData, err := pclntab.Data()
	if err != nil {
		return nil, fmt.Errorf("error reading .gopclntab data: %w", err)
	}
	pcln := gosym.NewLineTable(pclntabData, elfFile.Section(".text").Addr)
	table, err := gosym.NewTable(goSymTabData, pcln)
	if err != nil {
		return nil, fmt.Errorf("error parsing gosymtab: %w", err)
	}
	var files []string
	for k := range table.Files {
		files = append(files, k)
	}
	sort.Strings(files)
	return files, nil
}

func GetBuildId(file *os.File) (string, error) {
	elfFile, err := elf.NewFile(file)
	if err != nil {
		return "", fmt.Errorf("error opening ELF file: %w", err)
	}
	goBuildIdSection := elfFile.Section(".note.go.buildid")
	if goBuildIdSection == nil {
		return "", nil
	}
	data, err := goBuildIdSection.Data()
	if err != nil {
		return "", fmt.Errorf("error reading .note.go.buildid data: %w", err)
	}
	return string(data[16 : len(data)-1]), nil
}

func GetBuildInfo(file *os.File) *buildinfo.BuildInfo {
	info, err := buildinfo.Read(file)
	if err != nil {
		return nil
	}
	return info
}

func ClearMetadata(file *os.File) error {
	elfFile, err := elf.NewFile(file)
	if err != nil {
		return fmt.Errorf("error opening ELF file: %w", err)
	}
	pclntabFiles, err := GetPclntabFiles(file)
	if err != nil {
		return fmt.Errorf("error reading .gopclntab section: %w", err)
	}
	if pclntabFiles != nil {
		fmt.Println("Found .gopclntab section.  Obfuscating file names...")
		pclntab := elfFile.Section(".gopclntab")
		data, err := pclntab.Data()
		if err != nil {
			return fmt.Errorf("error reading .gopclntab data: %w", err)
		}
		for _, f := range pclntabFiles {
			ObfuscateString(data, f)
		}
		_, err = file.WriteAt(data, int64(pclntab.Offset))
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
		fmt.Println("Obfuscated file names has been saved!")
	}
	goBuildIdSection := elfFile.Section(".note.go.buildid")
	if goBuildIdSection != nil {
		fmt.Println("Found .note.go.buildid.  Obfuscating value...")
		data, err := goBuildIdSection.Data()
		if err != nil {
			return fmt.Errorf("erro reading .note.go.buildid data: %w", err)
		}
		for i := 16; i < len(data)-1; i++ {
			if data[i] != '/' {
				data[i] = byte(97 + rand.Intn(26))
			}
		}
		_, err = file.WriteAt(data, int64(goBuildIdSection.Offset))
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
		fmt.Println("Obfuscated Build ID has been saved!")
	}
	return nil
}

func ObfuscateString(source []byte, target string) {
	targetLength := len(target)
	obfuscated := make([]byte, targetLength)
	for i := 0; i < targetLength; i++ {
		obfuscated[i] = byte(97 + rand.Intn(26))
	}
	end := len(source) - targetLength
	for i := 0; i < end; i++ {
		if string(source[i:i+targetLength]) == target {
			copy(source[i:i+targetLength], obfuscated)
		}
	}
}
