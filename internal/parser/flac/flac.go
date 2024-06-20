package flac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
)

var Parser = parser.Parser{
	Name:      "FLAC",
	Container: false,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsFLAC(file, startOffset)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		metadata, err := GetMetadata(file, startOffset)
		if err != nil {
			return err
		}
		if action == parser.ShowAction {
			found := false
			for _, m := range metadata {
				if m.Type == 4 {
					comment, err := m.GetVorbisComment()
					if err != nil {
						return err
					}
					output.PrintForm(startOffset > 0, "Vendor String", comment.VendorString, 13)
					output.PrintHeader(startOffset > 0, "User Comments")
					for _, c := range comment.UserComment {
						output.Printf(startOffset > 0, " %s\n", c)
					}
					output.Println(startOffset > 0)
					found = true
				}
			}
			if !found {
				output.Println(startOffset > 0, "Vorbis comment metadata not found!")
			}
		} else if action == parser.ClearAction {
			found := false
			for _, m := range metadata {
				if m.Type == 4 {
					fmt.Println("Converting Vorbis comment into padding...")
					err = m.ConvertToPadding()
					if err != nil {
						return err
					}
					fmt.Println("Vorbis comment has been converted into padding!")
					found = true
				}
			}
			if !found {
				fmt.Println("Vorbis comment metadata not found!")
			}
		} else {
			fmt.Printf("Unssuported action: %s\n", action)
		}
		return nil
	},
}

func IsFLAC(file *os.File, startOffset int64) (bool, error) {
	magicBytes := make([]byte, 4)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0x66, 0x4C, 0x61, 0x43}), nil
}

func GetMetadata(file *os.File, offset int64) ([]Metadata, error) {
	var result []Metadata
	offset = offset + int64(4)
	for {
		header := make([]byte, 4)
		_, err := file.ReadAt(header, offset)
		if err != nil {
			return result, fmt.Errorf("failed to read file: %w", err)
		}
		metadata := Metadata{
			StartAt: offset,
			Type:    header[0] & 0x7F,
			Length:  binary.BigEndian.Uint32([]byte{0, header[1], header[2], header[3]}),
			File:    file,
			Last:    false,
		}
		result = append(result, metadata)
		offset += int64(metadata.Length) + 4
		if header[0]>>7 == 1 {
			metadata.Last = true
			break
		}
	}
	return result, nil
}

type Metadata struct {
	StartAt int64
	Type    byte
	Length  uint32
	File    *os.File
	Last    bool
}

func (m *Metadata) GetVorbisComment() (*VorbisComment, error) {
	if m.Type != 4 {
		return nil, fmt.Errorf("this metadata type %d doesn't contain Vorbis Comment", m.Type)
	}
	data := make([]byte, m.Length)
	_, err := m.File.ReadAt(data, m.StartAt+4)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vendorLength := binary.LittleEndian.Uint32(data[0:4])
	result := VorbisComment{
		VendorString: string(data[4 : 4+vendorLength]),
	}
	offset := 4 + vendorLength
	numOfUserComments := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	result.UserComment = make([]string, numOfUserComments)
	for i := 0; i < numOfUserComments; i++ {
		commentLength := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
		comment := string(data[offset : offset+commentLength])
		offset += commentLength
		result.UserComment[i] = comment
	}
	return &result, nil
}

func (m *Metadata) ConvertToPadding() error {
	data := make([]byte, 6+m.Length)
	_, err := m.File.ReadAt(data, m.StartAt)
	if err != nil {
		return err
	}
	if m.Last {
		data[0] = 0x81
	} else {
		data[0] = 0x01
	}
	for i := 4; i < len(data); i++ {
		data[i] = 0
	}
	_, err = m.File.WriteAt(data, m.StartAt)
	if err != nil {
		return err
	}
	return nil
}

type VorbisComment struct {
	VendorString string
	UserComment  []string
}
