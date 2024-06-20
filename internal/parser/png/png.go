package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
)

var Parser = parser.Parser{
	Name:      "PNG",
	Container: false,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsPNG(file, startOffset)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		if action == parser.ShowAction {
			textData, err := GetTextData(file, startOffset, length)
			if err != nil {
				return err
			}
			if len(textData) == 0 {
				output.Println(startOffset > 0, "Textual data not found!")
				return nil
			}
			width := 13
			for k := range textData {
				if len(k) > width {
					width = len(k)
				}
			}
			for k, v := range textData {
				output.PrintForm(startOffset > 0, k, v, width)
			}
		} else if action == parser.ClearAction {
			textData, err := GetTextData(file, startOffset, length)
			if err != nil {
				return err
			}
			if len(textData) == 0 {
				fmt.Println("There is no textual data to remove!")
				return nil
			}
			err = RemoveTextData(file)
			if err != nil {
				return err
			}
			fmt.Println("Textual data has been removed!")
		} else {
			fmt.Printf("Unssuported action: %s\n", action)
		}
		return nil
	},
}

func IsPNG(file *os.File, startOffset int64) (bool, error) {
	magicBytes := make([]byte, 8)
	_, err := file.ReadAt(magicBytes, startOffset)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.Equal(magicBytes, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0xA, 0x1A, 0x0A}), nil
}

func GetChunks(file *os.File, startOffset int64, length int64) ([]Chunk, error) {
	var result []Chunk
	offset := startOffset + int64(8)
	for {
		header := make([]byte, 8)
		_, err := file.ReadAt(header, offset)
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}
		chunk := Chunk{
			Length:    binary.BigEndian.Uint32(header[0:4]),
			ChunkType: header[4:8],
			StartAt:   offset,
			File:      file,
		}
		result = append(result, chunk)
		offset += 12 + int64(chunk.Length)
		if offset >= (startOffset + length) {
			break
		}
	}
	return result, nil
}

func GetTextData(file *os.File, startOffset int64, length int64) (map[string]string, error) {
	result := make(map[string]string)
	chunks, err := GetChunks(file, startOffset, length)
	if err != nil {
		return nil, err
	}
	for _, chunk := range chunks {
		if bytes.Equal(chunk.ChunkType, []byte{0x74, 0x45, 0x58, 0x74}) {
			keyword, value := chunk.ParseText()
			result[keyword] = value
		}
	}
	return result, nil
}

func RemoveTextData(file *os.File) error {
	fileInfo, _ := file.Stat()
	chunks, err := GetChunks(file, 0, fileInfo.Size())
	if err != nil {
		return err
	}
	tempFile, err := os.CreateTemp("", "jch_metadata_tmp_*")
	if err != nil {
		return err
	}
	_, err = tempFile.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0xA, 0x1A, 0x0A})
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		if bytes.Equal(chunk.ChunkType, []byte{0x74, 0x45, 0x58, 0x74}) {
			continue
		}
		chunkData := make([]byte, chunk.Length+12)
		_, err = file.ReadAt(chunkData, chunk.StartAt)
		if err != nil {
			return err
		}
		_, err = tempFile.Write(chunkData)
		if err != nil {
			return err
		}
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
	Length    uint32
	ChunkType []byte
	StartAt   int64
	File      *os.File
}

func (c *Chunk) ParseText() (string, string) {
	data := make([]byte, c.Length)
	_, _ = c.File.ReadAt(data, c.StartAt+8)
	i := 0
	for ; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
	}
	return string(data[0:i]), string(data[i+1:])
}
