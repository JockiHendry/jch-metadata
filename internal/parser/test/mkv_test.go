package test

import (
	"jch-metadata/internal/parser/mkv"
	"os"
	"testing"
)

func TestIsMkv(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := mkv.IsMkv(f, 0)
	if err != nil {
		t.Fatalf("Error inspecting file: %s", err)
	}
	if !result {
		t.Fatalf("Result should be true")
	}
	f, err = os.Open("internal/parser/test/test1.webm")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err = mkv.IsMkv(f, 0)
	if err != nil {
		t.Fatalf("Error inspecting file: %s", err)
	}
	if !result {
		t.Fatalf("Result should be true")
	}
}

func TestIsMkv_Unsupported(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.png")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := mkv.IsMkv(f, 0)
	if err != nil {
		t.Fatalf("Error inspecting file: %s", err)
	}
	if result {
		t.Fatalf("Result should be false")
	}
}

func TestGetVSize(t *testing.T) {
	size, offset := mkv.GetVSize([]byte{0x88})
	if size != 8 {
		t.Fatalf("Invalid calculated size, should be 8 but received %d", size)
	}
	if offset != 1 {
		t.Fatalf("Invalid calculated offset, should be 1 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x82})
	if size != 2 {
		t.Fatalf("Invalid calculated size, should be 2 but received %d", size)
	}
	if offset != 1 {
		t.Fatalf("Invalid calculated offset, should be 1 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x40, 0x02})
	if size != 2 {
		t.Fatalf("Invalid calculated size, should be 2 but received %d", size)
	}
	if offset != 2 {
		t.Fatalf("Invalid calculated offset, should be 2 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x20, 0x00, 0x02})
	if size != 2 {
		t.Fatalf("Invalid calculated size, should be 2 but received %d", size)
	}
	if offset != 3 {
		t.Fatalf("Invalid calculated offset, should be 3 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x10, 0x00, 0x00, 0x02})
	if size != 2 {
		t.Fatalf("Invalid calculated size, should be 2 but received %d", size)
	}
	if offset != 4 {
		t.Fatalf("Invalid calculated offset, should be 4 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3a})
	if size != 58 {
		t.Fatalf("Invalid calculated size, should be 58 but received %d", size)
	}
	if offset != 8 {
		t.Fatalf("Invalid calculated offset, should be 8 but received %d", offset)
	}
	size, offset = mkv.GetVSize([]byte{0x00})
	if size != 0 {
		t.Fatalf("Invalid calculated size, should be 0 but received %d", size)
	}
	if offset != 1 {
		t.Fatalf("Invalid calculated offset, should be 1 but received %d", offset)
	}
}

func TestParseFile(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	result, err := mkv.ParseFile(f)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if len(result) != 2 {
		t.Fatalf("Expected 2 elements found but received %d", len(result))
	}
	headerElement := mkv.SearchEBMLElements([]byte{0x1a, 0x45, 0xdf, 0xa3}, result)
	if headerElement == nil {
		t.Fatalf("Can't find header element")
	}
	if headerElement.StartAt != 0 {
		t.Fatalf("Expected header element started at file offset 0 but found %d", headerElement.StartAt)
	}
	if headerElement.Size != 19 {
		t.Fatalf("Expected header element size is 19 but found %d", headerElement.Size)
	}
	if headerElement.DataAt != 5 {
		t.Fatalf("Expected header element data started at file offset 5 but found %d", headerElement.DataAt)
	}
	headerElements, err := headerElement.GetElements()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if len(headerElements) != 3 {
		t.Fatalf("Expected header elements to be 3 but found %d", len(headerElements))
	}
	docTypeElement := mkv.SearchEBMLElements([]byte{0x42, 0x82}, headerElements)
	if docTypeElement == nil {
		t.Fatalf("Can't find docTypeElement")
	}
	if docTypeElement.Size != 8 {
		t.Fatalf("Invalid size for docTypeElement, should be 8 but received %d", docTypeElement.Size)
	}
	docTypeElementValue, err := docTypeElement.StringValue()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if docTypeElementValue != "matroska" {
		t.Fatalf("Invalid value for docTypeElement, should be 'matroska' but received %s", docTypeElementValue)
	}
	docTypeVersion := mkv.SearchEBMLElements([]byte{0x42, 0x87}, headerElements)
	if docTypeVersion == nil {
		t.Fatalf("Can't find docTypeVersion")
	}
	if docTypeVersion.Size != 1 {
		t.Fatalf("Invalid size for docTypeVersion, should be 1 but received %d", docTypeVersion.Size)
	}
	docTypeVersionValue, err := docTypeVersion.UintValue()
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if docTypeVersionValue != 2 {
		t.Fatalf("Invalid value for docTypeVersion, should be 2 but received %d", docTypeVersionValue)
	}
	rootElement := mkv.SearchEBMLElements([]byte{0x18, 0x53, 0x80, 0x67}, result)
	if rootElement == nil {
		t.Fatalf("Can't find root element")
	}
	if rootElement.StartAt != 24 {
		t.Fatalf("Expected root element started at file offset 24 but found %d", rootElement.StartAt)
	}
	if rootElement.Size != 23339305 {
		t.Fatalf("Expected root element size is 23339305 but found %d", rootElement.Size)
	}
	if rootElement.DataAt != 32 {
		t.Fatalf("Expected root element data started at file offset 32 but found %d", rootElement.StartAt)
	}
	if mkv.SearchEBMLElements([]byte{0x99, 0x99}, result) != nil {
		t.Fatalf("Searching non existing element should return nil")
	}
}

func TestGetMetadata(t *testing.T) {
	f, err := os.Open("internal/parser/test/test1.mkv")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	metadata, err := mkv.GetMetadata(f)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	if metadata[0].Info.Filename != "" {
		t.Fatalf("Expecting Filename to be empty but found %s", metadata[0].Info.Filename)
	}
	if metadata[0].Info.Title != "" {
		t.Fatalf("Expecting Title to be empty but found %s", metadata[0].Info.Title)
	}
	if metadata[0].Info.DateUTC.UTC().String() != "2010-08-21 07:23:03 +0000 UTC" {
		t.Fatalf("Expecting DateUTC to be '2010-08-21 07:23:03 +0000 UTC' but found %s", metadata[0].Info.DateUTC.UTC())
	}
	if metadata[0].Info.MuxingApp != "libebml2 v0.10.0 + libmatroska2 v0.10.1" {
		t.Fatalf("Expected MuxingApp to be 'libebml2 v0.10.0 + libmatroska2 v0.10.1' but found %s", metadata[0].Info.MuxingApp)
	}
	if metadata[0].Info.WritingApp != "mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08" {
		t.Fatalf("Expected WritingApp to be 'mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08' but found %s", metadata[0].Info.WritingApp)
	}
}
