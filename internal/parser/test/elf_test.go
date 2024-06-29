package test

import (
	debugElf "debug/elf"
	"jch-metadata/internal/parser/elf"
	"os"
	"testing"
)

func TestIsELF(t *testing.T) {
	f, err := os.Open("internal/parser/test/test")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := elf.IsELF(f, 0)
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsELF_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := elf.IsELF(f, 0)
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestGetDWARFFiles(t *testing.T) {
	f, err := os.Open("internal/parser/test/test")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	files, err := elf.GetDWARFFiles(f)
	if err != nil {
		t.Fatalf("Error reading DWARF source files: %s", err)
	}
	if len(files) != 415 {
		t.Fatalf("Unexpected number of files: %d", len(files))
	}
}

func TestGetPclntab(t *testing.T) {
	f, err := os.Open("internal/parser/test/test")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	files, err := elf.GetPclntabFiles(f)
	if len(files) != 187 {
		t.Fatalf("Unexpected number of files: %d", len(files))
	}
}

func TestGetBuildID(t *testing.T) {
	f, err := os.Open("internal/parser/test/test")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	buildId, err := elf.GetBuildId(f)
	if err != nil {
		t.Fatalf("Failed to get Build ID: %s", err)
	}
	if buildId != "fOqpebvfm4cCEk2BfjBL/bKOqjm8SStyrXVTKjXgJ/l48vQKVF2zyAF8A8Ds6N/eCOMgPIS1nMYAh_LZ88C" {
		t.Fatalf("Invalid Build ID: %s", buildId)
	}
}

func TestObfuscateString(t *testing.T) {
	f, err := debugElf.Open("internal/parser/test/test")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	data, err := f.Section(".gopclntab").Data()
	if err != nil {
		t.Fatalf("Error reading .gopclntab section: %s", err)
	}
	target := "/tmp/test/main.go"
	end := len(data) - len(target)
	found := false
	for i := 0; i < end; i++ {
		if string(data[i:i+len(target)]) == target {
			found = true
		}
	}
	if !found {
		t.Fatalf("String to obfuscate not found in .gopclntab")
	}
	elf.ObfuscateString(data, "/tmp/test/main.go")
	found = false
	for i := 0; i < end; i++ {
		if string(data[i:i+len(target)]) == target {
			found = true
		}
	}
	if found {
		t.Fatalf("The string should be obfuscated")
	}
}
