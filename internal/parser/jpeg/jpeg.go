package jpeg

import (
	"bytes"
	"fmt"
	"io"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
	"sort"
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
			if len(metadata.EXIFTags) > 0 {
				output.PrintHeader(startOffset > 0, "EXIF Tags")
				tags := make([]uint16, len(metadata.EXIFTags))
				i := 0
				for k := range metadata.EXIFTags {
					tags[i] = k
					i++
				}
				sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
				for _, t := range tags {
					output.PrintForm(startOffset > 0, fmt.Sprintf("0x%04X", t), metadata.EXIFTags[t], 10)
				}
				output.Println(startOffset > 0)
			}

			output.PrintHeader(startOffset > 0, "XMP")
			for _, s := range metadata.XMP {
				output.PrintMultiline(startOffset > 0, s)
				output.Println(startOffset > 0)
			}

			for _, m := range metadata.UnsupportedMarkers {
				output.PrintHeader(startOffset > 0, "Application Segment 0x%04X", m.GetMarker())
				output.PrintHexDump(startOffset > 0, m.Raw.Bytes())
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

func FindApplicationMarkers(file *os.File, startOffset int64) ([]ApplicationMarker, error) {
	offset := startOffset
	data := make([]byte, 512)
	marked := byte(0)
	var appMarker ApplicationMarker
	var result []ApplicationMarker
	for {
		_, err := file.ReadAt(data, offset)
		if err == io.EOF {
			break
		}
		for _, b := range data {
			if b == 0xFF {
				marked = b
				continue
			} else if marked == 0xFF {
				if b == 0x00 {
					marked = 0
					continue
				}
				marked = b
				if appMarker.IsValid() {
					result = append(result, appMarker)
				}
				appMarker = ApplicationMarker{}
				appMarker.Raw.Grow(512)
				appMarker.Raw.Write([]byte{0xFF, b})
				continue
			} else if marked >= 0xE0 && marked <= 0xEF {
				appMarker.Raw.WriteByte(b)
			}
		}
		offset += 512
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
			result.JFIFThumbnail = m.Raw.Bytes()[16] > 0 && m.Raw.Bytes()[17] > 0
		} else if m.IsJFXXSegment() {
			result.JFXXThumbnail = true
		} else if m.IsEXIFSegment() {
			result.EXIFTags = m.GetEXIFValues()
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

type ApplicationMarker struct {
	Raw bytes.Buffer
}

func (m *ApplicationMarker) IsValid() bool {
	if m.Raw.Len() <= 2 {
		return false
	}
	marker := m.GetMarker()
	return marker[1] >= 0xE0 && marker[1] <= 0xEF
}

func (m *ApplicationMarker) IsJFIFSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE0}) {
		return false
	}
	raw := m.Raw.Bytes()
	if len(raw) < 9 {
		return false
	}
	if string(raw[4:8]) != "JFIF" {
		return false
	}
	return true
}

func (m *ApplicationMarker) IsJFXXSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE0}) {
		return false
	}
	raw := m.Raw.Bytes()
	if len(raw) < 9 {
		return false
	}
	return string(raw[4:8]) == "JFXX"
}

func (m *ApplicationMarker) IsEXIFSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE1}) {
		return false
	}
	raw := m.Raw.Bytes()
	if len(raw) < 11 {
		return false
	}
	return bytes.Equal(raw[4:10], []byte{0x45, 0x78, 0x69, 0x66, 0x00, 0x00})
}

func (m *ApplicationMarker) IsICCProfileSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE2}) {
		return false
	}
	raw := m.Raw.Bytes()
	return string(raw[4:15]) == "ICC_PROFILE"
}

func (m *ApplicationMarker) IsXMPSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE1}) {
		return false
	}
	raw := m.Raw.Bytes()
	return string(raw[4:32]) == "http://ns.adobe.com/xap/1.0/"
}

func (m *ApplicationMarker) IsExtendedXMPSegment() bool {
	if !bytes.Equal(m.GetMarker(), []byte{0xFF, 0xE1}) {
		return false
	}
	raw := m.Raw.Bytes()
	return string(raw[4:38]) == "http://ns.adobe.com/xmp/extension/"
}

func (m *ApplicationMarker) GetEXIFValues() map[uint16]string {
	raw := m.Raw.Bytes()
	return ParseExif(raw[10:])
}

func (m *ApplicationMarker) GetICCProfile() Profile {
	raw := m.Raw.Bytes()
	if raw[16] != 1 {
		return Profile{}
	}
	return ParseICC(raw[18:])
}

func (m *ApplicationMarker) GetXMP() string {
	raw := m.Raw.Bytes()
	return string(raw[33:])
}

func (m *ApplicationMarker) GetExtendedXMP() string {
	raw := m.Raw.Bytes()
	return string(raw[79:])
}

func (m *ApplicationMarker) GetMarker() []byte {
	marker := m.Raw.Bytes()
	return marker[0:2]
}

type Metadata struct {
	JFIFThumbnail      bool
	JFXXThumbnail      bool
	EXIFTags           map[uint16]string
	ICCProfile         Profile
	UnsupportedMarkers []ApplicationMarker
	XMP                []string
}
