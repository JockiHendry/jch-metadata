package test

import (
	"bytes"
	"jch-metadata/internal/parser/jpeg"
	"os"
	"testing"
)

func TestIsJPEG(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.jpeg")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.IsJPEG(f, 0)
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsJPEG_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.IsJPEG(f, 0)
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestFindApplicationMarkers(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.jpeg")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.FindApplicationMarkers(f, 0)
	if len(result) != 4 {
		t.Fatalf("Unexpected result size: %d", len(result))
	}
	if !bytes.Equal(result[0].GetMarker(), []byte{0xFF, 0xE0}) {
		t.Fatalf("Unexpected marker: %v", result[0].GetMarker())
	}
	if !bytes.Equal(result[1].GetMarker(), []byte{0xFF, 0xE1}) {
		t.Fatalf("Unexpected marker: %v", result[1].GetMarker())
	}
	if !bytes.Equal(result[2].GetMarker(), []byte{0xFF, 0xE0}) {
		t.Fatalf("Unexpected marker: %v", result[2].GetMarker())
	}
	if !bytes.Equal(result[3].GetMarker(), []byte{0xFF, 0xE2}) {
		t.Fatalf("Unexpected marker: %v", result[3].GetMarker())
	}
}

func TestParseJPEG(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.jpeg")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.ParseFile(f, 0)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if result.JFIFThumbnail {
		t.Fatalf("Image doesn't have JFIF Thumbnail")
	}
	if result.JFXXThumbnail {
		t.Fatalf("Image doesn't have JFXX Thumbnail")
	}
	if len(result.EXIFTags) != 40 {
		t.Fatalf("Unexpected number of EXIF tags: %d", len(result.EXIFTags))
	}
	if result.EXIFTags[0x110] != "Canon EOS 40D" {
		t.Fatalf("Unexpected value for tag ID: %s", result.EXIFTags[0x110])
	}
	if result.EXIFTags[0x131] != "GIMP 2.4.5" {
		t.Fatalf("Unexpected value for tag ID: %s", result.EXIFTags[0x131])
	}
	if len(result.UnsupportedMarkers) != 0 {
		t.Fatalf("Unexpected number of unsupported markers: %d", len(result.UnsupportedMarkers))
	}
	if result.ICCProfile.ProfileCreator != "HP " {
		t.Fatalf("Unexpected ICC Profile creator: %s", result.ICCProfile.ProfileCreator)
	}
	if result.ICCProfile.ProfileClass != "mntr" {
		t.Fatalf("Unexpected ICC Profile class: %s", result.ICCProfile.ProfileClass)
	}
	if result.ICCProfile.DeviceModel != "sRGB" {
		t.Fatalf("Unexpected device model: %s", result.ICCProfile.DeviceModel)
	}
	if result.ICCProfile.DeviceManufacturer != "IEC " {
		t.Fatalf("Unexpected device manufacturer: %s", result.ICCProfile.DeviceManufacturer)
	}
	if result.ICCProfile.PrimaryPlatform != "MSFT" {
		t.Fatalf("Unexpected primary platform: %s", result.ICCProfile.PrimaryPlatform)
	}
	if result.ICCProfile.CmmType != "Lino" {
		t.Fatalf("Unexpected CMM type: %s", result.ICCProfile.CmmType)
	}
	if result.ICCProfile.Copyright != "Copyright (c) 1998 Hewlett-Packard Company" {
		t.Fatalf("Unexpected Copyright: %s", result.ICCProfile.Copyright)
	}
}

func TestParseXMP(t *testing.T) {
	f, err := os.Open("internal/parser/test/test2.jpeg")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.ParseFile(f, 0)
	if len(result.XMP) != 2 {
		t.Fatalf("Invalid XMP chunks: %d", len(result.XMP))
	}
	if len(result.XMP[0]) != 838 {
		t.Fatalf("Invalid size for first XMP chunk: %d", len(result.XMP[0]))
	}
	if len(result.XMP[1]) != 52041 {
		t.Fatalf("Invalid size for second XMP chunk: %d", len(result.XMP[0]))
	}
}
