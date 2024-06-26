package mkv

import (
	"bytes"
	"fmt"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Parser = parser.Parser{
	Name:      "Mkv (Matroska)",
	Container: true,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsMkv(file, startOffset)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		if action == parser.ShowAction {
			metadata, err := GetMetadata(file)
			if err != nil {
				return err
			}
			for _, m := range metadata {
				err := Show(m, file, parsers)
				if err != nil {
					return err
				}
			}
		} else if action == parser.ClearAction {
			err := ClearMetadata(file)
			if err != nil {
				return err
			}
			output.Println(false, "Metadata cleared")
		} else if action == parser.ExtractAction {
			attachmentElement, err := GetElementFromSeek(file, []byte{0x19, 0x41, 0xA4, 0x69})
			if err != nil {
				return err
			}
			if attachmentElement == nil {
				fmt.Println("No attachment to extract!")
				return nil
			}
			attachments := NewAttachments(attachmentElement)
			for _, a := range attachments {
				output.Printf(false, "Extracting attachment %s...", a.Name)
				err = ExtractAttachment(file, a)
				if err != nil {
					return err
				}
			}
		} else {
			output.Printf(false, "Unsupported action: %s\n", action)
		}
		return nil
	},
}

func Show(m Metadata, file *os.File, parsers []parser.Parser) error {
	output.PrintHeader(false, "Info")
	output.PrintForm(false, "Filename", m.Info.Filename, 13)
	var dateStr string
	if m.Info.DateUTC.Year() <= 1970 {
		dateStr = ""
	} else {
		dateStr = m.Info.DateUTC.String()
	}
	output.PrintForm(false, "Date", dateStr, 13)
	output.PrintForm(false, "Title", m.Info.Title, 13)
	output.PrintForm(false, "Muxing App", m.Info.MuxingApp, 13)
	output.PrintForm(false, "Writing App", m.Info.WritingApp, 13)
	output.Println(false)

	for _, t := range m.Tracks {
		output.PrintHeader(false, "Track %d", t.Number)
		output.PrintForm(false, "Name", t.Name, 13)
		output.PrintForm(false, "Type", GetTrackType(t.Type), 13)
		output.PrintForm(false, "Language", t.Language, 13)
		output.Println(false)
	}

	for _, a := range m.Attachments {
		output.PrintHeader(false, "Attachment")
		output.PrintForm(false, "Name", a.Name, 13)
		output.PrintForm(false, "Media Type", a.MediaType, 13)
		output.PrintForm(false, "Description", a.Description, 13)
		output.Println(false)
		parsed, err := parser.StartParsing(parsers, file, parser.ShowAction, a.DataAt, a.Size)
		if err != nil {
			return fmt.Errorf("error while processing attachment [%s]: %w", a.Name, err)
		}
		if !parsed {
			output.Println(true, "Unsupported file type")
		}
		output.Println(false)
	}

	for _, t := range m.Tags {
		if t.Name == "BPS" {
			continue
		}
		output.PrintHeader(false, "Tag")
		output.PrintForm(false, "Name", t.Name, 13)
		output.PrintForm(false, "Target Type", t.TargetType, 13)
		output.PrintForm(false, "Language", t.Language, 13)
		output.PrintForm(false, "Value", t.Value, 13)
		output.Println(false)
	}

	return nil
}

func IsMkv(file *os.File, startOffset int64) (bool, error) {
	magicBytes := make([]byte, 4)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0x1A, 0x45, 0xDF, 0xA3}), nil
}

func ParseFile(file *os.File) ([]EBMLElement, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file stat: %w", err)
	}
	fileSize := fileInfo.Size()
	return GetEBMLElements(file, 0, fileSize, 9999)
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

func GetMetadata(file *os.File) ([]Metadata, error) {
	var result []Metadata
	metadata := Metadata{}
	infoElement, err := GetElementFromSeek(file, []byte{0x15, 0x49, 0xA9, 0x66})
	if err != nil {
		return nil, err
	}
	infoElements := infoElement.GetElements()
	metadata.Info.Filename = GetStringValue([]byte{0x73, 0x84}, infoElements)
	metadata.Info.Title = GetStringValue([]byte{0x7B, 0xA9}, infoElements)
	metadata.Info.DateUTC = GetDateValue([]byte{0x44, 0x61}, infoElements)
	metadata.Info.MuxingApp = GetStringValue([]byte{0x4D, 0x80}, infoElements)
	metadata.Info.WritingApp = GetStringValue([]byte{0x57, 0x41}, infoElements)

	var tracks []Track
	trackElement, err := GetElementFromSeek(file, []byte{0x16, 0x54, 0xAE, 0x6B})
	if err != nil {
		return nil, err
	}
	if trackElement != nil {
		trackElements := trackElement.GetElements()
		for _, t := range trackElements {
			if bytes.Equal(t.ElementID, []byte{0xAE}) {
				children := t.GetElements()
				track := Track{
					Number:   GetUInt64Value([]byte{0xD7}, children),
					Name:     GetStringValue([]byte{0x53, 0x6E}, children),
					Type:     GetUInt64Value([]byte{0x83}, children),
					Language: GetStringValue([]byte{0x22, 0xB5, 0x9C}, children),
				}
				tracks = append(tracks, track)
			}
		}
	}
	metadata.Tracks = tracks

	var attachments []Attachment
	attachmentElement, err := GetElementFromSeek(file, []byte{0x19, 0x41, 0xA4, 0x69})
	if err != nil {
		return nil, err
	}
	if attachmentElement != nil {
		attachments = NewAttachments(attachmentElement)
	}
	metadata.Attachments = attachments

	var tags []Tag
	tagElement, err := GetElementFromSeek(file, []byte{0x12, 0x54, 0xC3, 0x67})
	if err != nil {
		return nil, err
	}
	if tagElement != nil {
		tagElements := tagElement.GetElements()
		for _, t := range tagElements {
			if bytes.Equal(t.ElementID, []byte{0x73, 0x73}) {
				children := t.GetElements()
				tag := Tag{
					TargetType: GetStringValue([]byte{0x63, 0xCA}, children),
				}
				simpleTag := SearchEBMLElements([]byte{0x67, 0xC8}, children)
				if simpleTag != nil {
					simpleTags := simpleTag.GetElements()
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
		e := v.GetElements()

		output.Println(false, "Removing all values from Info elements...")
		infoElement := SearchEBMLElements([]byte{0x15, 0x49, 0xA9, 0x66}, e)
		infoElements := infoElement.GetElements()
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

func GetElementFromSeek(file *os.File, elementId []byte) (*EBMLElement, error) {
	elements, err := ParseFile(file)
	if err != nil {
		return nil, err
	}
	rootElement := SearchEBMLElements([]byte{0x18, 0x53, 0x80, 0x67}, elements)
	seekElements := rootElement.FindFirstElement([]byte{0x11, 0x4D, 0x9B, 0x74}, nil)
	var seekResult *EBMLElement
	for _, seekElement := range seekElements.GetElements() {
		search := seekElement.FindFirstElement([]byte{0x53, 0xAB}, elementId)
		if search != &EmptyElement {
			seekResult = &seekElement
			break
		}
	}
	if seekResult == nil {
		return nil, nil
	}
	offset, err := seekResult.FindFirstElement([]byte{0x53, 0xAC}, nil).UintValue()
	if err != nil {
		return nil, err
	}
	result, err := GetEBMLElements(file, rootElement.DataAt+int64(offset), rootElement.DataAt+int64(rootElement.Size), 1)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func NewAttachments(attachmentElement *EBMLElement) []Attachment {
	var attachments []Attachment
	attachmentsElements := attachmentElement.GetElements()
	for i, a := range attachmentsElements {
		if bytes.Equal(a.ElementID, []byte{0x61, 0xA7}) {
			name, _ := a.FindFirstElement([]byte{0x46, 0x6E}, nil).StringValue()
			descriptionElement := a.FindFirstElement([]byte{0x46, 0x7E}, nil)
			description := ""
			if descriptionElement != nil {
				description, _ = descriptionElement.StringValue()
			}
			mediaType, _ := a.FindFirstElement([]byte{0x46, 0x60}, nil).StringValue()
			attachment := Attachment{
				Index:       i,
				Name:        name,
				Description: description,
				MediaType:   mediaType,
			}
			attachmentData := a.FindFirstElement([]byte{0x46, 0x5C}, nil)
			attachment.DataAt = attachmentData.DataAt
			attachment.Size = int64(attachmentData.Size)
			attachments = append(attachments, attachment)
		}
	}
	return attachments
}

func ExtractAttachment(file *os.File, attachment Attachment) error {
	data := make([]byte, attachment.Size)
	_, err := file.ReadAt(data, attachment.DataAt)
	if err != nil {
		return err
	}
	err = os.MkdirAll("output", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}
	ext := "raw"
	if attachment.MediaType != "" {
		m := strings.Split(attachment.MediaType, "/")
		if len(m) >= 2 {
			ext = m[1]
		}
	}
	basename := filepath.Base(strings.TrimSuffix(file.Name(), ".mkv"))
	filename := filepath.Join("output", fmt.Sprintf("%s_attachment_%02d.%s", basename, attachment.Index, ext))
	err = os.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing attachment %02d: %w", attachment.Index, err)
	}
	output.Printf(false, "Attachment has been extracted to %s\n", filename)
	return nil
}

type Metadata struct {
	Info struct {
		Filename   string
		DateUTC    time.Time
		Title      string
		MuxingApp  string
		WritingApp string
	}
	Tracks      []Track
	Attachments []Attachment
	Tags        []Tag
}

type Track struct {
	Number   uint64
	Name     string
	Type     uint64
	Language string
}

type Attachment struct {
	Index       int
	Name        string
	MediaType   string
	Description string
	DataAt      int64
	Size        int64
}

type Tag struct {
	Name       string
	TargetType string
	Language   string
	Value      string
}
