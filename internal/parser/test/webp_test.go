package test

import (
	"jch-metadata/internal/parser/webp"
	"os"
	"testing"
)

func TestIsWebP(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.webp")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	fileInfo, _ := f.Stat()
	result, err := webp.IsWebp(f, 0, fileInfo.Size())
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsWebP_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	fileInfo, _ := f.Stat()
	result, err := webp.IsWebp(f, 0, fileInfo.Size())
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestWebpGetChunks(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.webp")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	fileInfo, _ := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file stat")
	}
	chunks, err := webp.GetChunks(f, 0, fileInfo.Size())
	if err != nil {
		t.Fatalf("Error getting chunks: %s", err)
	}
	if len(chunks) != 4 {
		t.Fatalf("Unexpected chunks size: %d", len(chunks))
	}
	if chunks[0].StartAt != 12 {
		t.Fatalf("Unexpected StartAt: %d", chunks[0].StartAt)
	}
	if chunks[0].FourC != "VP8X" {
		t.Fatalf("Unexpected FourC: %s", chunks[0].FourC)
	}
	if chunks[0].Size != 10 {
		t.Fatalf("Unexpected Size: %d", chunks[0].Size)
	}
	if chunks[1].StartAt != 30 {
		t.Fatalf("Unexpected StartAt: %d", chunks[1].StartAt)
	}
	if chunks[1].FourC != "VP8 " {
		t.Fatalf("Unexpected FourC: %s", chunks[1].FourC)
	}
	if chunks[1].Size != 23010 {
		t.Fatalf("Unexpected Size: %d", chunks[1].Size)
	}
	if chunks[2].StartAt != 23048 {
		t.Fatalf("Unexpected StartAt: %d", chunks[2].StartAt)
	}
	if chunks[2].FourC != "EXIF" {
		t.Fatalf("Unexpected FourC: %s", chunks[2].FourC)
	}
	if chunks[2].Size != 6422 {
		t.Fatalf("Unexpected Size: %d", chunks[2].Size)
	}
	if chunks[3].StartAt != 29478 {
		t.Fatalf("Unexpected StartAt: %d", chunks[3].StartAt)
	}
	if chunks[3].FourC != "XMP " {
		t.Fatalf("Unexpected FourC: %s", chunks[3].FourC)
	}
	if chunks[3].Size != 325 {
		t.Fatalf("Unexpected Size: %d", chunks[3].Size)
	}

	ifds, err := chunks[2].GetExif()
	if err != nil {
		t.Fatalf("Error reading EXIF data: %s", err)
	}
	if len(ifds) != 3 {
		t.Fatalf("Invalid IFD size: %d", len(ifds))
	}
	if ifds[0].StartOffset != 8 {
		t.Fatalf("Invalid start offset: %d", ifds[0].StartOffset)
	}
	if len(ifds[0].Tags) != 10 {
		t.Fatalf("Invalid number of tags: %d", len(ifds[0].Tags))
	}
	if ifds[0].Tags[0x10F] != "Samsung" {
		t.Fatalf("Invalid tags: 0x10F => %s", ifds[0].Tags[0x10F])
	}
	if ifds[0].Tags[0x110] != "Galaxy Nexus" {
		t.Fatalf("Invalid tags: 0x110 => %s", ifds[0].Tags[0x110])
	}
	if ifds[0].Tags[0x112] != "65536" {
		t.Fatalf("Invalid tags: 0x112 => %s", ifds[0].Tags[0x112])
	}
	if ifds[1].StartOffset != 210 {
		t.Fatalf("Invalid start offset: %d", ifds[1].StartOffset)
	}
	if len(ifds[1].Tags) != 38 {
		t.Fatalf("Invalid number of tags: %d", len(ifds[1].Tags))
	}
	if ifds[1].Tags[0x9004] != "2013:10:21 15:19:01" {
		t.Fatalf("Invalid tags: 0x9004 => %s", ifds[1].Tags[0x9004])
	}
	if ifds[1].Tags[0xA420] != "d95be34ec8879acecb4922b372f51e81" {
		t.Fatalf("Invalid tags: 0xA420 => %s", ifds[1].Tags[0xA420])
	}
	if ifds[2].StartOffset != 1008 {
		t.Fatalf("Invalid start offset: %d", ifds[2].StartOffset)
	}
	if len(ifds[2].Tags) != 6 {
		t.Fatalf("Invalid number of tags: %d", len(ifds[2].Tags))
	}
	if ifds[2].Tags[0x103] != "393216" {
		t.Fatalf("Invalid tags: 0x103 => %s", ifds[2].Tags[0x103])
	}

	xmp, err := chunks[3].GetXMP()
	if err != nil {
		t.Fatalf("Error reading XMP data: %s", err)
	}
	if len(xmp) != 325 {
		t.Fatalf("Invalid XMP value")
	}
}
