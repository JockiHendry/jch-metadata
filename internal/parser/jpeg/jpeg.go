package jpeg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var Parser = parser.Parser{
	Name:      "JPEG",
	Container: false,
	Support: func(file *os.File, startOffset int64) (bool, error) {
		return IsJPEG(file, startOffset)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, parsers []parser.Parser) error {
		if action == parser.ShowAction {
			metadata, err := ParseFile(file, startOffset)
			if err != nil {
				return err
			}
			output.PrintHeader(startOffset > 0, "JFIF Segments")
			output.PrintForm(startOffset > 0, "Has JFIF Thumbnail", fmt.Sprintf("%v", metadata.JFIFThumbnail), 20)
			output.PrintForm(startOffset > 0, "Has JFXX Thumbnail", fmt.Sprintf("%v", metadata.JFXXThumbnail), 20)
			output.Println(startOffset > 0)
			if len(metadata.IFDs) > 0 {
				for _, ifd := range metadata.IFDs {
					output.PrintHeader(startOffset > 0, "EXIF IFD Offset 0x%0X", ifd.StartOffset)
					tags := make([]uint16, len(ifd.Tags))
					i := 0
					for k := range ifd.Tags {
						tags[i] = k
						i++
					}
					sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
					for _, t := range tags {
						output.PrintForm(startOffset > 0, fmt.Sprintf("0x%04X", t), ifd.Tags[t], 10)
					}
					output.Println(startOffset > 0)
				}
			}

			output.PrintHeader(startOffset > 0, "XMP")
			for _, s := range metadata.XMP {
				output.PrintMultiline(startOffset > 0, s)
				output.Println(startOffset > 0)
			}

			for _, m := range metadata.UnsupportedMarkers {
				output.PrintHeader(startOffset > 0, "Application Segment 0x%04X", m.Marker)
				output.PrintHexDump(startOffset > 0, m.Raw)
				output.Println(startOffset > 0)
			}
			output.Println(startOffset > 0)
			output.PrintHeader(startOffset > 0, "ICC Profile")
			output.PrintForm(startOffset > 0, "CMM Type", metadata.ICCProfile.CmmType, 18)
			output.PrintForm(startOffset > 0, "Profile Class", metadata.ICCProfile.ProfileClass, 18)
			output.PrintForm(startOffset > 0, "Primary Platform", metadata.ICCProfile.PrimaryPlatform, 18)
			output.PrintForm(startOffset > 0, "Dev Manufacturer", metadata.ICCProfile.DeviceManufacturer, 18)
			output.PrintForm(startOffset > 0, "Dev Model", metadata.ICCProfile.DeviceModel, 18)
			output.PrintForm(startOffset > 0, "Profile Creator", metadata.ICCProfile.ProfileCreator, 18)
			output.PrintForm(startOffset > 0, "Copyright", metadata.ICCProfile.Copyright, 18)
		} else if action == parser.ExtractAction {
			thumbnailData, err := ExtractThumbnail(file, startOffset)
			if err != nil {
				return fmt.Errorf("error extracting thumbnail: %w", err)
			}
			if thumbnailData == nil {
				output.Println(startOffset > 0, "No thumbnail to extract")
			}
			err = os.MkdirAll("output", os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating output directory: %w", err)
			}
			ext := filepath.Ext(file.Name())
			basename := filepath.Base(strings.TrimSuffix(file.Name(), ext))
			filename := filepath.Join("output", basename+"_thumbnail.jpeg")
			err = os.WriteFile(filename, thumbnailData, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error writing thumbnail: %w", err)
			}
			output.Printf(startOffset > 0, "Thumbnail has been extracted to %s\n", filename)
		}
		return nil
	},
}

func IsJPEG(file *os.File, startOffset int64) (bool, error) {
	magicBytes := make([]byte, 3)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0xFF, 0xD8, 0xFF}), nil
}

func FindApplicationMarkers(file *os.File, startOffset int64) ([]ApplicationSegment, error) {
	offset := startOffset
	var result []ApplicationSegment
	appMarker := ApplicationSegment{
		Marker: make([]byte, 2),
	}
	for {
		_, err := file.ReadAt(appMarker.Marker, offset)
		if err == io.EOF {
			break
		}
		if appMarker.Marker[0] == 0xFF && appMarker.Marker[1] >= 0xE0 && appMarker.Marker[1] <= 0xEF {
			lengthRaw := make([]byte, 2)
			_, err := file.ReadAt(lengthRaw, offset+2)
			if err != nil {
				return nil, err
			}
			appMarker.Length = binary.BigEndian.Uint16(lengthRaw)
			appMarker.Raw = make([]byte, appMarker.Length+2)
			_, err = file.ReadAt(appMarker.Raw, offset)
			if err != nil {
				return nil, err
			}
			result = append(result, appMarker)
			offset += int64(2 + appMarker.Length)
			appMarker = ApplicationSegment{
				Marker: make([]byte, 2),
			}
		} else {
			offset += 2
		}
	}
	return result, nil
}

func ParseFile(file *os.File, startOffset int64) (*Metadata, error) {
	markers, err := FindApplicationMarkers(file, startOffset)
	if err != nil {
		return nil, err
	}
	result := Metadata{
		JFIFThumbnail: false,
		JFXXThumbnail: false,
	}
	for _, m := range markers {
		if m.IsJFIFSegment() {
			result.JFIFThumbnail = m.Raw[16] > 0 && m.Raw[17] > 0
		} else if m.IsJFXXSegment() {
			result.JFXXThumbnail = true
		} else if m.IsEXIFSegment() {
			result.IFDs = m.GetIFDs()
		} else if m.IsICCProfileSegment() {
			result.ICCProfile = m.GetICCProfile()
		} else if m.IsXMPSegment() {
			result.XMP = append(result.XMP, m.GetXMP())
		} else if m.IsExtendedXMPSegment() {
			result.XMP = append(result.XMP, m.GetExtendedXMP())
		} else {
			result.UnsupportedMarkers = append(result.UnsupportedMarkers, m)
		}
	}
	return &result, nil
}

func ExtractThumbnail(file *os.File, startOffset int64) ([]byte, error) {
	markers, err := FindApplicationMarkers(file, startOffset)
	if err != nil {
		return nil, err
	}
	var exifSegment *ApplicationSegment
	for _, m := range markers {
		if m.IsEXIFSegment() {
			exifSegment = &m
			break
		}
	}
	if exifSegment == nil {
		return nil, nil
	}
	ifds := exifSegment.GetIFDs()
	for _, ifd := range ifds {
		if ifd.Tags[0x0103] == "6" {
			offsetStr, exists := ifd.Tags[0x0201]
			if !exists {
				continue
			}
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse thumbnail offset [%s]: %w", offsetStr, err)
			}
			sizeStr, exists := ifd.Tags[0x0202]
			if !exists {
				continue
			}
			size, err := strconv.Atoi(sizeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse thumbnail size [%s]: %w", sizeStr, err)
			}
			start := offset + 10
			return exifSegment.Raw[start : start+size], nil
		}
	}
	return nil, nil
}

type ApplicationSegment struct {
	StartOffset int64
	Marker      []byte
	Length      uint16
	Raw         []byte
}

func (m *ApplicationSegment) IsJFIFSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE0}) {
		return false
	}
	if m.Length < 9 {
		return false
	}
	if string(m.Raw[4:8]) != "JFIF" {
		return false
	}
	return true
}

func (m *ApplicationSegment) IsJFXXSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE0}) {
		return false
	}
	if m.Length < 9 {
		return false
	}
	return string(m.Raw[4:8]) == "JFXX"
}

func (m *ApplicationSegment) IsEXIFSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE1}) {
		return false
	}
	if m.Length < 11 {
		return false
	}
	return bytes.Equal(m.Raw[4:10], []byte{0x45, 0x78, 0x69, 0x66, 0x00, 0x00})
}

func (m *ApplicationSegment) IsICCProfileSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE2}) {
		return false
	}
	return string(m.Raw[4:15]) == "ICC_PROFILE"
}

func (m *ApplicationSegment) IsXMPSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE1}) {
		return false
	}
	return string(m.Raw[4:32]) == "http://ns.adobe.com/xap/1.0/"
}

func (m *ApplicationSegment) IsExtendedXMPSegment() bool {
	if !bytes.Equal(m.Marker, []byte{0xFF, 0xE1}) {
		return false
	}
	return string(m.Raw[4:38]) == "http://ns.adobe.com/xmp/extension/"
}

func (m *ApplicationSegment) GetIFDs() []IFD {
	return ParseExif(m.Raw[10:])
}

func (m *ApplicationSegment) GetICCProfile() Profile {
	if m.Raw[16] != 1 {
		return Profile{}
	}
	return ParseICC(m.Raw[18:])
}

func (m *ApplicationSegment) GetXMP() string {
	return string(m.Raw[33:])
}

func (m *ApplicationSegment) GetExtendedXMP() string {
	return string(m.Raw[79:])
}

type Metadata struct {
	JFIFThumbnail      bool
	JFXXThumbnail      bool
	IFDs               []IFD
	ICCProfile         Profile
	UnsupportedMarkers []ApplicationSegment
	XMP                []string
}