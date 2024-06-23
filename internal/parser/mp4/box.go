package mp4

import (
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/output"
	"strings"
	"time"
)

type PrintableBox interface {
	GetType() string
	GetData() ([]byte, error)
	FindNestedBoxByType(boxType string) PrintableBox
	Print()
}

func ConvertBox(box Box) PrintableBox {
	switch box.Type {
	case "ftyp":
		return FileTypeBox{&box}
	case "moov":
		return MoovBox{&box}
	case "mvhd":
		return MvhdBox{&box}
	case "meta":
		return MetaBox{&box}
	case "trak":
		return TrakBox{&box}
	case "tkhd":
		return TrackHeaderBox{&box}
	case "mdia":
		return TrackMediaBox{&box}
	case "mdhd":
		return TrackMediaHeader{MvhdBox{&box}}
	case "hdlr":
		return HandlerBox{&box}
	case "minf":
		return MediaInformationBox{&box}
	case "dinf":
		return DataInformationBox{&box}
	case "udta":
		return UdtaBox{&box}
	default:
		return box
	}
}

type FileTypeBox struct {
	*Box
}

func (b FileTypeBox) GetFileType() (*FileType, error) {
	data, err := b.GetData()
	if err != nil {
		return nil, err
	}
	result := FileType{
		Brand:        string(data[0:4]),
		MinorVersion: binary.BigEndian.Uint32(data[4:8]),
	}
	var compatibleBrands []string
	for i := 8; i < len(data); i += 4 {
		compatibleBrands = append(compatibleBrands, string(data[i:i+4]))
	}
	result.CompatibleBrands = compatibleBrands
	return &result, nil
}

func (b FileTypeBox) Print() {
	fileType, err := b.GetFileType()
	if err != nil {
		output.Printf(false, "Failed to read file type box: %s\n", err)
		return
	}
	output.PrintHeader(false, "File Type")
	output.PrintForm(false, "Brand", fileType.Brand, 35)
	output.PrintForm(false, "Minor Version", fmt.Sprintf("%d", fileType.MinorVersion), 35)
	output.PrintForm(false, "Compatible Brands", strings.Join(fileType.CompatibleBrands, ", "), 35)
	output.Println(false)
}

type MoovBox struct {
	*Box
}

func (b MoovBox) Print() {
	output.PrintHeader(false, "Movie Metadata (moov)")
	output.Println(false)
	boxes := b.GetBoxes()
	for _, b := range boxes {
		b.Print()
	}
}

type FileType struct {
	Brand            string
	MinorVersion     uint32
	CompatibleBrands []string
}

type MvhdBox struct {
	*Box
}

func (b MvhdBox) Print() {
	output.Println(false, "Movie Header (mvhd)")
	result, err := b.GetHeader()
	if err != nil {
		output.Printf(false, "Error parsing movie header: %s\n", err)
	}
	output.PrintForm(false, "Creation Time", result.LocalCreationTime(), 20)
	output.PrintForm(false, "Modification Time", result.LocalModificationTime(), 20)
	output.PrintForm(false, "Duration (sec)", result.DurationString(), 20)
	output.Println(false)
}

func (b MvhdBox) GetHeader() (*Header, error) {
	data, err := b.GetData()
	if err != nil {
		return nil, err
	}
	result := Header{}
	var creationTime uint64
	var modificationTime uint64
	if data[0] == 1 {
		creationTime = binary.BigEndian.Uint64(data[4:12])
		modificationTime = binary.BigEndian.Uint64(data[12:20])
		result.Timescale = binary.BigEndian.Uint32(data[20:24])
		result.Duration = binary.BigEndian.Uint64(data[24:32])
	} else {
		creationTime = uint64(binary.BigEndian.Uint32(data[4:8]))
		modificationTime = uint64(binary.BigEndian.Uint32(data[8:12]))
		result.Timescale = binary.BigEndian.Uint32(data[12:16])
		result.Duration = uint64(binary.BigEndian.Uint32(data[16:20]))
	}
	result.CreationTime = BaseTime.Add(time.Second * time.Duration(creationTime))
	result.ModificationTime = BaseTime.Add(time.Second * time.Duration(modificationTime))
	return &result, nil
}

type Header struct {
	CreationTime     time.Time
	ModificationTime time.Time
	Timescale        uint32
	Duration         uint64
}

func (h Header) LocalCreationTime() string {
	return h.CreationTime.Local().String()
}

func (h Header) LocalModificationTime() string {
	return h.ModificationTime.Local().String()
}

func (h Header) DurationString() string {
	return time.Duration((h.Duration / uint64(h.Timescale)) * uint64(time.Second)).String()
}

type UdtaBox struct {
	*Box
}

func (b UdtaBox) Print() {
	data, err := b.GetData()
	if err != nil {
		fmt.Printf("Error retrieving udta content: %s\n", err)
		return
	}
	output.PrintHeader(false, "User data (udta)")
	output.Println(false)
	offset := 0
	for offset < len(data) {
		length := binary.BigEndian.Uint32(data[offset : offset+4])
		dataValue := data[offset : offset+int(length)]
		output.PrintHexDump(false, dataValue)
		output.Println(false)
		offset += int(length)
	}
	output.Println(false)
}
