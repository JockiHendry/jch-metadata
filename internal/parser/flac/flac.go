package flac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/parser"
	"os"
)

var Parser = parser.Parser{
	Name: "FLAC",
	Support: func(file *os.File) (bool, error) {
		return IsFLAC(file)
	},
	Handle: func(file *os.File, action parser.Action) error {
		metadata, err := GetMetadata(file)
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
					fmt.Printf("Vendor String  :  %s\n", comment.VendorString)
					fmt.Println("User Comments  :")
					for _, c := range comment.UserComment {
						fmt.Printf(" %s\n", c)
					}
					fmt.Println()
					found = true
				}
			}
			if !found {
				fmt.Println("Vorbis comment metadata not found!")
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

func IsFLAC(file *os.File) (bool, error) {
	magicBytes := make([]byte, 4)
	_, err := file.ReadAt(magicBytes, 0)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0x66, 0x4C, 0x61, 0x43}), nil
}

func GetMetadata(file *os.File) ([]Metadata, error) {
	var result []Metadata
	offset := int64(4)
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
