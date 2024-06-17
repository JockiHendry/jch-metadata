package test

import (
	"jch-metadata/internal/parser/flac"
	"os"
	"testing"
)

func TestIsFlac(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.flac")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := flac.IsFLAC(f, 0)
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsFlac_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := flac.IsFLAC(f, 0)
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestFlacMetadata(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.flac")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	metadata, err := flac.GetMetadata(f, 0)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if len(metadata) != 4 {
		t.Fatalf("Expecting 4 metadata but found %d instead", len(metadata))
	}
	if metadata[0].Type != 0 {
		t.Fatalf("Expecting first metadata type to be 0 (STREAMINFO) but found %d", metadata[0].Type)
	}
	if metadata[0].StartAt != 4 {
		t.Fatalf("Expecting first metadata to be started at offset 4 but found %d", metadata[0].StartAt)
	}
	if metadata[0].Length != 34 {
		t.Fatalf("Expecting first metadata size to be 34 but found %d", metadata[0].Length)
	}
	if metadata[1].Type != 3 {
		t.Fatalf("Expecting second metadata type to be 3 (SEEKTABLE) but found %d", metadata[1].Type)
	}
	if metadata[1].StartAt != 42 {
		t.Fatalf("Expecting second metadata to be started at offset 42 but found %d", metadata[1].StartAt)
	}
	if metadata[1].Length != 18 {
		t.Fatalf("Expecting second metadata size to be 18 but found %d", metadata[1].Length)
	}
	if metadata[2].Type != 4 {
		t.Fatalf("Expecting third metadata type to be 4 (VORBIS_COMMENT) but found %d", metadata[2].Type)
	}
	if metadata[2].StartAt != 64 {
		t.Fatalf("Expecting third metadata to be started at offset 64 but found %d", metadata[2].StartAt)
	}
	if metadata[2].Length != 68 {
		t.Fatalf("Expecting third metadata size to be 68 but found %d", metadata[2].Length)
	}
	if metadata[3].Type != 1 {
		t.Fatalf("Expecting third metadata type to be 1 (PADDING) but found %d", metadata[3].Type)
	}
	if metadata[3].StartAt != 136 {
		t.Fatalf("Expecting third metadata to be started at offset 136 but found %d", metadata[3].StartAt)
	}
	if metadata[3].Length != 8192 {
		t.Fatalf("Expecting third metadata size to be 8192 but found %d", metadata[3].Length)
	}

	vorbisComment, err := metadata[2].GetVorbisComment()
	if err != nil {
		t.Fatalf("Failed to retrieve Vorbis Comment: %s", err)
	}
	if vorbisComment.VendorString != "reference libFLAC 1.3.2 20170101" {
		t.Fatalf("Unexpected vendor string: %s", vorbisComment.VendorString)
	}
	if vorbisComment.UserComment[0] != "Comment=Processed by SoX" {
		t.Fatalf("Unexpected user comment: %s", vorbisComment.UserComment[0])
	}
}
