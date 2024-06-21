package webp

import (
	"encoding/binary"
	"fmt"
	"io"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"jch-metadata/internal/parser/shared"
	"os"
)

var Parser = parser.Parser{
	Name:      "Webp",
	Container: false,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsWebp(file, startOffset, length)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		chunks, err := GetChunks(file, startOffset, length)
		if err != nil {
			return err
		}
		if action == parser.ShowAction {
			for _, c := range chunks {
				if c.FourC == "EXIF" {
					ifds, err := c.GetExif()
					if err != nil {
						return err
					}
					shared.PrintExif(startOffset > 0, ifds)
				} else if c.FourC == "XMP" {
					xmp, err := c.GetXMP()
					if err != nil {
						return err
					}
					output.PrintHeader(startOffset > 0, "XMP")
					output.PrintMultiline(startOffset > 0, xmp)
					output.Println(startOffset > 0)
				} else if c.FourC == "ICCP" {
					icc, err := c.GetICC()
					if err != nil {
						return err
					}
					shared.PrintICC(startOffset > 0, icc)
				}
			}
		} else if action == parser.ClearAction {
			hasMetadata := false
			for _, c := range chunks {
				if c.IsMetadata() {
					hasMetadata = true
					break
				}
			}
			if !hasMetadata {
				fmt.Println("No metadata found in file!")
				return nil
			}
			fmt.Println("Filtering metadata chunks...")
			err := ClearMetadata(file, chunks)
			if err != nil {
				return err
			}
			fmt.Println("Metadata chunks have been removed!")
		}
		return nil
	},
}

func IsWebp(file *os.File, startOffset int64, length int64) (bool, error) {
	magicBytes := make([]byte, 24)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	if string(magicBytes[0:4]) != "RIFF" {
		return false, nil
	}
	size := binary.LittleEndian.Uint32(magicBytes[4:8])
	if size != uint32(length-8) {
		return false, nil
	}
	if string(magicBytes[8:12]) != "WEBP" {
		return false, nil
	}
	return true, nil
}

func GetChunks(file *os.File, startOffset int64, length int64) ([]Chunk, error) {
	var result []Chunk
	offset := startOffset + 12
	for {
		header := make([]byte, 8)
		_, err := file.ReadAt(header, offset)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header[3] == 0x00 {
			header[3] = 0x20
		}
		chunk := Chunk{
			FourC:   string(header[0:4]),
			Size:    binary.LittleEndian.Uint32(header[4:8]),
			StartAt: offset,
			File:    file,
		}
		result = append(result, chunk)
		offset += 8 + int64(chunk.Size)
		if offset >= length {
			break
		}
	}
	return result, nil
}

func ClearMetadata(file *os.File, chunks []Chunk) error {
	var result = []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}
	for _, c := range chunks {
		if c.IsMetadata() {
			continue
		}
		data, err := c.GetData()
		if err != nil {
			return err
		}
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, c.Size)
		result = append(result, []byte(c.FourC)...)
		result = append(result, length...)
		result = append(result, data...)
	}
	binary.LittleEndian.PutUint32(result[4:8], uint32(len(result)-8))
	tempFile, err := os.CreateTemp("", "jch_metadata_tmp_*")
	if err != nil {
		return err
	}
	_, err = tempFile.Write(result)
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tempFile.Name(), file.Name())
	if err != nil {
		return err
	}
	return nil
}

type Chunk struct {
	FourC   string
	Size    uint32
	StartAt int64
	File    *os.File
}

func (c *Chunk) GetData() ([]byte, error) {
	data := make([]byte, c.Size)
	_, err := c.File.ReadAt(data, c.StartAt+8)
	if err != nil {
		return nil, fmt.Errorf("error reading data at offset %d: %w", c.StartAt+4, err)
	}
	return data, nil
}

func (c *Chunk) GetExif() ([]shared.IFD, error) {
	data, err := c.GetData()
	if err != nil {
		return nil, err
	}
	return shared.ParseExif(data[:]), nil
}

func (c *Chunk) GetXMP() (string, error) {
	data, err := c.GetData()
	if err != nil {
		return "", err
	}
	return string(data[:]), nil
}

func (c *Chunk) GetICC() (*shared.Profile, error) {
	data, err := c.GetData()
	if err != nil {
		return nil, err
	}
	return shared.ParseICC(data[:]), nil
}

func (c *Chunk) IsMetadata() bool {
	return c.FourC == "EXIF" || c.FourC == "XMP " || c.FourC == "ICCP"
}
