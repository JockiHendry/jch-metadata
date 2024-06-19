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
	if len(result) != 3 {
		t.Fatalf("Unexpected result size: %d", len(result))
	}
	if !bytes.Equal(result[0].Marker, []byte{0xFF, 0xE0}) {
		t.Fatalf("Unexpected marker: %v", result[0].Marker)
	}
	if !bytes.Equal(result[1].Marker, []byte{0xFF, 0xE1}) {
		t.Fatalf("Unexpected marker: %v", result[1].Marker)
	}
	if !bytes.Equal(result[2].Marker, []byte{0xFF, 0xE2}) {
		t.Fatalf("Unexpected marker: %v", result[3].Marker)
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
	if len(result.IFDs) != 3 {
		t.Fatalf("Unexpected number of IFDs: %d", len(result.IFDs))
	}
	if len(result.IFDs[0].Tags) != 10 {
		t.Fatalf("Unexpected number of EXIF tags in IFD 0: %d", len(result.IFDs[0].Tags))
	}
	if len(result.IFDs[1].Tags) != 27 {
		t.Fatalf("Unexpected number of EXIF tags in IFD 1: %d", len(result.IFDs[1].Tags))
	}
	if len(result.IFDs[2].Tags) != 6 {
		t.Fatalf("Unexpected number of EXIF tags in IFD 2: %d", len(result.IFDs[2].Tags))
	}
	if result.IFDs[0].Tags[0x110] != "Canon EOS 40D" {
		t.Fatalf("Unexpected value for tag ID: %s", result.IFDs[0].Tags[0x110])
	}
	if result.IFDs[0].Tags[0x131] != "GIMP 2.4.5" {
		t.Fatalf("Unexpected value for tag ID: %s", result.IFDs[0].Tags[0x131])
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
	if len(result.XMP) != 22 {
		t.Fatalf("Invalid XMP chunks: %d", len(result.XMP))
	}
	if len(result.XMP[0]) != 838 {
		t.Fatalf("Invalid size for first XMP chunk: %d", len(result.XMP[0]))
	}
	if len(result.XMP[1]) != 65383 {
		t.Fatalf("Invalid size for second XMP chunk: %d", len(result.XMP[1]))
	}
}

func TestExtract(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.jpeg")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := jpeg.ExtractThumbnail(f, 0)
	if err != nil {
		t.Fatalf("Error extracting thumbnail: %s", err)
	}
	if len(result) == 0 {
		t.Fatalf("Unexpected thumbnail size: %d", len(result))
	}
	if result[0] != 0xFF && result[1] != 0xD8 {
		t.Fatalf("Invalid thumbnail")
	}
}
