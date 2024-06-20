package test

import (
	"bytes"
	"jch-metadata/internal/parser/png"
	"os"
	"testing"
)

func TestIsPNG(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := png.IsPNG(f, 0)
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsPNG_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.flac")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := png.IsPNG(f, 0)
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestGetChunks(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	fileInfo, _ := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file stat")
	}
	chunks, err := png.GetChunks(f, 0, fileInfo.Size())
	if err != nil {
		t.Fatalf("Error getting chunks: %s", err)
	}
	if len(chunks) != 43 {
		t.Fatalf("Unexpected chunks size: %d", len(chunks))
	}
	if !bytes.Equal(chunks[2].ChunkType, []byte{0x74, 0x45, 0x58, 0x74}) {
		t.Fatalf("Invalid chunk type: %v", chunks[2].ChunkType)
	}
	if chunks[2].Length != 25 {
		t.Fatalf("Invalid chunk length: %d", chunks[2].Length)
	}
	if chunks[2].StartAt != 49 {
		t.Fatalf("Invalid start offset: %d", chunks[2].StartAt)
	}
}

func TestGetTextData(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	fileInfo, _ := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file stat")
	}
	result, err := png.GetTextData(f, 0, fileInfo.Size())
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if result["Software"] != "gnome-screenshot" {
		t.Fatalf("Unexpected text data: %s", result["Software"])
	}
	if result["Creation Time"] != "2023-05-20T02:56:29+0700" {
		t.Fatalf("Unexpected text data: %s", result["CreationTime"])
	}
}
