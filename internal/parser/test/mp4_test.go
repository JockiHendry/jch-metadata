package test

import (
	"fmt"
	"jch-metadata/internal/parser/mp4"
	"os"
	"testing"
)

func TestIsMP4(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mp4")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := mp4.IsMP4(f)
	if err != nil {
		t.Fatalf("Error inspecting file: %s", err)
	}
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsMP4_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := mp4.IsMP4(f)
	if err != nil {
		t.Fatalf("Error inspecting file: %s", err)
	}
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestGetBoxes(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mp4")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	boxes, err := mp4.GetBoxes(f, 0, stat.Size())
	if err != nil {
		t.Fatalf("Error getting boxes: %s", err)
	}
	if len(boxes) != 3 {
		t.Fatalf("Unexpected number of boxes: %d", len(boxes))
	}
	boxes0 := boxes[0].(mp4.FileTypeBox)
	if boxes0.Type != "ftyp" {
		t.Fatalf("Unexpected box type: %s", boxes0.Type)
	}
	if boxes0.Size != 28 {
		t.Fatalf("Unexpected box size: %d", boxes0.Size)
	}
	if boxes0.StartOffset != 0 {
		t.Fatalf("Unexpected start offset: %d", boxes0.StartOffset)
	}
	boxes1 := boxes[1].(mp4.MoovBox)
	if boxes1.Type != "moov" {
		t.Fatalf("Unexpected box type: %s", boxes1.Type)
	}
	if boxes1.Size != 3937 {
		t.Fatalf("Unexpected box size: %d", boxes1.Size)
	}
	if boxes1.StartOffset != 28 {
		t.Fatalf("Unexpected start offset: %d", boxes1.StartOffset)
	}
	boxes2 := boxes[2].(mp4.Box)
	if boxes2.Type != "mdat" {
		t.Fatalf("Unexpected box type: %s", boxes2.Type)
	}
	if boxes2.Size != 1293680 {
		t.Fatalf("Unexpected box size: %d", boxes2.Size)
	}
	if boxes2.StartOffset != 3965 {
		t.Fatalf("Unexpected start offset: %d", boxes2.StartOffset)
	}
	fmt.Println(boxes)
}

func TestGetFileType(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mp4")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	boxes, err := mp4.GetBoxes(f, 0, stat.Size())
	if err != nil {
		t.Fatalf("Error getting boxes: %s", err)
	}
	fileTypeBox := boxes[0].(mp4.FileTypeBox)
	if fileTypeBox.Size != 28 {
		t.Fatalf("Unexpected box size: %d", fileTypeBox.Size)
	}
}

func TestGetMovieHeader(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mp4")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	boxes, err := mp4.GetBoxes(f, 0, stat.Size())
	if err != nil {
		t.Fatalf("Error getting boxes: %s", err)
	}
	boxes = boxes[1].(mp4.MoovBox).GetBoxes()
	if err != nil {
		t.Fatalf("Error getting boxes: %s", err)
	}
	header, err := boxes[0].(mp4.MvhdBox).GetHeader()
	if err != nil {
		t.Fatalf("Error getting movie header: %s", err)
	}
	if header.CreationTime.UTC().String() != "2003-03-03 10:03:03 +0000 UTC" {
		t.Fatalf("Invalid creationg time: %s", header.CreationTime.UTC().String())
	}
	if header.ModificationTime.UTC().String() != "2003-03-03 10:03:03 +0000 UTC" {
		t.Fatalf("Invalid creationg time: %s", header.ModificationTime.UTC().String())
	}
	if header.Timescale != 25000 {
		t.Fatalf("Invalid timescale: %d", header.Timescale)
	}
	if header.Duration != 100000 {
		t.Fatalf("Invalid duration: %d", header.Duration)
	}
}

func TestMdta(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mp4")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	boxes, err := mp4.GetBoxes(f, 0, stat.Size())
	if err != nil {
		t.Fatalf("Error getting boxes: %s", err)
	}
	metaBox := boxes[1].(mp4.MoovBox).FindNestedBoxByType("meta").(mp4.MetaBox)
	result, err := metaBox.GetMdta()
	if err != nil {
		t.Fatalf("Error retreiving metadata: %s", err)
	}
	if len(result) != 7 {
		t.Fatalf("Invalid number of keys: %d", len(result))
	}
	if result["com.apple.quicktime.creationdate"].String() != "2005-05-05T12:05:05+0200" {
		t.Fatalf("Invalid value for key: %s", result["com.apple.quicktime.creationdate"].String())
	}
	if result["com.apple.quicktime.location.ISO6709"].String() != "-36.6101-066.91515+119.900/" {
		t.Fatalf("Invalid value for key: %s", result["com.apple.quicktime.location.ISO6709"].String())
	}
}
