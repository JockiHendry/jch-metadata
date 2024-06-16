package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

var MkvParser = Parser{
	Name: "Mkv (Matroska)",
	Support: func(file *os.File) (bool, error) {
		return IsMkv(file)
	},
	Handle: func(file *os.File, action Action) error {
		metadata, err := GetMetadata(file)
		if err != nil {
			return err
		}
		if action == ShowAction {
			for _, m := range metadata {
				Show(m)
			}
		} else if action == ClearAction {
			err = ClearMetadata(file)
			if err != nil {
				return err
			}
			fmt.Println("Metadata cleared")
		} else {
			fmt.Printf("Unsupported action: %s\n", action)
		}
		return nil
	},
}

func Show(m MkvMetadata) {
	fmt.Println("Info")
	fmt.Println("====")
	fmt.Printf("Filename    : %s\n", m.Info.Filename)
	var dateStr string
	if m.Info.DateUTC.Year() <= 1970 {
		dateStr = ""
	} else {
		dateStr = m.Info.DateUTC.String()
	}
	fmt.Printf("Date        : %s\n", dateStr)
	fmt.Printf("Title       : %s\n", m.Info.Title)
	fmt.Printf("Muxing App  : %s\n", m.Info.MuxingApp)
	fmt.Printf("Writing App : %s\n", m.Info.WritingApp)
	fmt.Println()

	for _, t := range m.Tracks {
		fmt.Printf("Track %d\n", t.Number)
		fmt.Println("=========")
		fmt.Printf("Name     : %s\n", t.Name)
		fmt.Printf("Type     : %s\n", GetTrackType(t.Type))
		fmt.Printf("Language : %s\n", t.Language)
		fmt.Println()
	}

	for _, a := range m.Attachments {
		fmt.Println("Attachment")
		fmt.Println("==========")
		fmt.Printf("Name        : %s\n", a.Name)
		fmt.Printf("Media Type  : %s\n", a.MediaType)
		fmt.Printf("Description : %s\n", a.Description)
		fmt.Println()
	}

	for _, t := range m.Tags {
		if t.Name == "BPS" {
			continue
		}
		fmt.Println("Tag")
		fmt.Println("===")
		fmt.Printf("Name        : %s\n", t.Name)
		fmt.Printf("Target Type : %s\n", t.TargetType)
		fmt.Printf("Language    : %s\n", t.Language)
		fmt.Printf("Value       : %s\n", t.Value)
		fmt.Println()
	}
}

func IsMkv(file *os.File) (bool, error) {
	magicBytes := make([]byte, 4)
	n, err := file.ReadAt(magicBytes, 0)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	if n != 4 {
		return false, fmt.Errorf("unexpected bytes returned: %d", n)
	}
	return bytes.Equal(magicBytes, []byte{0x1A, 0x45, 0xDF, 0xA3}), nil
}

func GetVSize(v []byte) (uint64, int) {
	if v[0] == 0 {
		return 0, 1
	}
	i := 0
	for {
		if v[0]&(byte(128)>>i) == (byte(128) >> i) {
			break
		}
		i = i + 1
	}
	i = i + 1
	startIndex := 8 - i
	result := make([]byte, 8)
	result[startIndex] = v[0] & (byte(0xFF) >> i)
	startIndex++
	for j := 1; j < i; j++ {
		result[startIndex] = v[j]
		startIndex++
	}
	return binary.BigEndian.Uint64(result), i
}

type EBMLElement struct {
	ElementID []byte
	Size      uint64
	StartAt   int64
	DataAt    int64
	File      *os.File
}

func (e *EBMLElement) GetElements() ([]EBMLElement, error) {
	return GetEBMLElements(e.File, e.DataAt, e.DataAt+int64(e.Size))
}

func (e *EBMLElement) StringValue() (string, error) {
	if e.Size == 0 {
		return "", nil
	}
	data := make([]byte, e.Size)
	_, err := e.File.ReadAt(data, e.DataAt)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

func (e *EBMLElement) UintValue() (uint64, error) {
	if e.Size == 0 {
		return 0, nil
	}
	data := make([]byte, e.Size)
	_, err := e.File.ReadAt(data, e.DataAt)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}
	tmp := make([]byte, 8)
	startIndex := 0
	for i := 0; i < 8-len(data); i++ {
		tmp[startIndex] = 0
		startIndex++
	}
	for i := 0; i < len(data); i++ {
		tmp[startIndex] = data[i]
		startIndex++
	}
	return binary.BigEndian.Uint64(tmp), nil
}

func (e *EBMLElement) DateValue() (time.Time, error) {
	if e.Size == 0 {
		return time.Unix(0, 0), nil
	}
	data := make([]byte, 8)
	_, err := e.File.ReadAt(data, e.DataAt)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to read file: %w", err)
	}
	dateValue := binary.BigEndian.Uint64(data)
	if dateValue == 0 {
		return time.Unix(0, 0), nil
	}
	return time.Unix(978307200, int64(dateValue)), nil
}

func (e *EBMLElement) ClearValue() error {
	if e.Size == 0 {
		return nil
	}
	data := make([]byte, e.Size)
	for i := range data {
		data[i] = 0
	}
	_, err := e.File.WriteAt(data, e.DataAt)
	if err != nil {
		return err
	}
	return nil
}

func GetEBMLElements(file *os.File, fileOffset int64, limit int64) ([]EBMLElement, error) {
	var result []EBMLElement

	for {
		// Retrieve elementID
		elementIdData := make([]byte, 8)
		_, err := file.ReadAt(elementIdData, fileOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		_, offset := GetVSize(elementIdData[:])
		element := EBMLElement{
			ElementID: elementIdData[0 : 0+offset],
			StartAt:   fileOffset,
			File:      file,
		}
		fileOffset += int64(offset)

		// Retrieve elementSize
		sizeData := make([]byte, 8)
		_, err = file.ReadAt(sizeData, fileOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		size, offset := GetVSize(sizeData[:])
		element.Size = size
		fileOffset += int64(offset)
		element.DataAt = fileOffset

		result = append(result, element)
		fileOffset += int64(size)

		if fileOffset >= limit {
			break
		}
	}
	return result, nil
}

func SearchEBMLElements(elementID []byte, elements []EBMLElement) *EBMLElement {
	for _, v := range elements {
		if bytes.Equal(v.ElementID, elementID) {
			return &v
		}
	}
	return nil
}

func ParseFile(file *os.File) ([]EBMLElement, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file stat: %w", err)
	}
	fileSize := fileInfo.Size()
	return GetEBMLElements(file, 0, fileSize)
}

func GetStringValue(elementId []byte, elements []EBMLElement) string {
	element := SearchEBMLElements(elementId, elements)
	if element == nil {
		return ""
	}
	v, _ := element.StringValue()
	return v
}

func GetUInt64Value(elementId []byte, elements []EBMLElement) uint64 {
	element := SearchEBMLElements(elementId, elements)
	if element == nil {
		return 0
	}
	v, _ := element.UintValue()
	return v
}

func GetDateValue(elementId []byte, elements []EBMLElement) time.Time {
	element := SearchEBMLElements(elementId, elements)
	if element == nil {
		return time.Unix(0, 0)
	}
	v, _ := element.DateValue()
	return v
}

func ClearValue(elementId []byte, elements []EBMLElement) error {
	element := SearchEBMLElements(elementId, elements)
	if element == nil {
		return nil
	}
	return element.ClearValue()
}
func GetTrackType(value uint64) string {
	switch value {
	case 1:
		return "video"
	case 2:
		return "audio"
	case 3:
		return "complex"
	case 16:
		return "logo"
	case 17:
		return "subtitle"
	case 18:
		return "buttons"
	case 32:
		return "control"
	case 33:
		return "metadata"
	default:
		return "unknown"
	}
}

func GetMetadata(file *os.File) ([]MkvMetadata, error) {
	var result []MkvMetadata
	elements, err := ParseFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file stat: %w", err)
	}
	for _, v := range elements {
		if !bytes.Equal(v.ElementID, []byte{0x18, 0x53, 0x80, 0x67}) {
			continue
		}
		metadata := MkvMetadata{}
		e, _ := v.GetElements()

		infoElement := SearchEBMLElements([]byte{0x15, 0x49, 0xA9, 0x66}, e)
		infoElements, _ := infoElement.GetElements()
		metadata.Info.Filename = GetStringValue([]byte{0x73, 0x84}, infoElements)
		metadata.Info.Title = GetStringValue([]byte{0x7B, 0xA9}, infoElements)
		metadata.Info.DateUTC = GetDateValue([]byte{0x44, 0x61}, infoElements)
		metadata.Info.MuxingApp = GetStringValue([]byte{0x4D, 0x80}, infoElements)
		metadata.Info.WritingApp = GetStringValue([]byte{0x57, 0x41}, infoElements)

		var tracks []MkvTrack
		trackElement := SearchEBMLElements([]byte{0x16, 0x54, 0xAE, 0x6B}, e)
		if trackElement != nil {
			trackElements, _ := trackElement.GetElements()
			for _, t := range trackElements {
				if bytes.Equal(t.ElementID, []byte{0xAE}) {
					elements, _ := t.GetElements()
					track := MkvTrack{
						Number:   GetUInt64Value([]byte{0xD7}, elements),
						Name:     GetStringValue([]byte{0x53, 0x6E}, elements),
						Type:     GetUInt64Value([]byte{0x83}, elements),
						Language: GetStringValue([]byte{0x22, 0xB5, 0x9C}, elements),
					}
					tracks = append(tracks, track)
				}
			}
		}
		metadata.Tracks = tracks

		var attachments []MkvAttachment
		attachmentElement := SearchEBMLElements([]byte{0x19, 0x41, 0xA4, 0x69}, e)
		if attachmentElement != nil {
			attachmentsElements, _ := attachmentElement.GetElements()
			for _, a := range attachmentsElements {
				if bytes.Equal(a.ElementID, []byte{0x61, 0xA7}) {
					elements, _ := a.GetElements()
					attachment := MkvAttachment{
						Name:        GetStringValue([]byte{0x46, 0x6E}, elements),
						Description: GetStringValue([]byte{0x46, 0x7E}, elements),
						MediaType:   GetStringValue([]byte{0x46, 0x60}, elements),
					}
					attachments = append(attachments, attachment)
				}
			}
		}
		metadata.Attachments = attachments

		var tags []MkvTag
		tagElement := SearchEBMLElements([]byte{0x12, 0x54, 0xC3, 0x67}, e)
		if tagElement != nil {
			tagElements, _ := tagElement.GetElements()
			for _, t := range tagElements {
				if bytes.Equal(t.ElementID, []byte{0x73, 0x73}) {
					elements, _ := t.GetElements()
					tag := MkvTag{
						TargetType: GetStringValue([]byte{0x63, 0xCA}, elements),
					}
					simpleTag := SearchEBMLElements([]byte{0x67, 0xC8}, elements)
					if simpleTag != nil {
						simpleTags, _ := simpleTag.GetElements()
						tag.Name = GetStringValue([]byte{0x45, 0xA3}, simpleTags)
						tag.Language = GetStringValue([]byte{0x44, 0x7A}, simpleTags)
						tag.Value = GetStringValue([]byte{0x44, 0x87}, simpleTags)
						tags = append(tags, tag)
					}

				}
			}
		}
		metadata.Tags = tags

		result = append(result, metadata)
	}
	return result, nil
}

func ClearMetadata(file *os.File) error {
	elements, err := ParseFile(file)
	if err != nil {
		return err
	}
	for _, v := range elements {
		if !bytes.Equal(v.ElementID, []byte{0x18, 0x53, 0x80, 0x67}) {
			continue
		}
		e, _ := v.GetElements()

		fmt.Println("Removing all values from Info elements...")
		infoElement := SearchEBMLElements([]byte{0x15, 0x49, 0xA9, 0x66}, e)
		infoElements, _ := infoElement.GetElements()
		err := ClearValue([]byte{0x73, 0x84}, infoElements)
		if err != nil {
			return err
		}
		err = ClearValue([]byte{0x7B, 0xA9}, infoElements)
		if err != nil {
			return err
		}
		err = ClearValue([]byte{0x44, 0x61}, infoElements)
		if err != nil {
			return err
		}
		err = ClearValue([]byte{0x4D, 0x80}, infoElements)
		if err != nil {
			return err
		}
		err = ClearValue([]byte{0x57, 0x41}, infoElements)
		if err != nil {
			return err
		}
	}
	return nil
}

type MkvMetadata struct {
	Info struct {
		Filename   string
		DateUTC    time.Time
		Title      string
		MuxingApp  string
		WritingApp string
	}
	Tracks      []MkvTrack
	Attachments []MkvAttachment
	Tags        []MkvTag
}

type MkvTrack struct {
	Number   uint64
	Name     string
	Type     uint64
	Language string
}

type MkvAttachment struct {
	Name        string
	MediaType   string
	Description string
}

type MkvTag struct {
	Name       string
	TargetType string
	Language   string
	Value      string
}
