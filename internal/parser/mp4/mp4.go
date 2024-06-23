package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
	"jch-metadata/internal/output"
	"jch-metadata/internal/parser"
	"os"
	"time"
)

var BaseTime = time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)

var Parser = parser.Parser{
	Name:      "MP4",
	Container: true,
	Support: func(file *os.File, startOffset int64, length int64) (bool, error) {
		return IsMP4(file)
	},
	Handle: func(file *os.File, action parser.Action, startOffset int64, length int64, parsers []parser.Parser) error {
		if action == parser.ShowAction {
			boxes, err := GetBoxes(file, startOffset, length)
			if err != nil {
				return err
			}
			for _, box := range boxes {
				box.Print()
			}
		} else if action == parser.ClearAction {
			err := ClearMetadata(file, startOffset, length)
			if err != nil {
				return err
			}
			output.Println(false, "Metadata has been cleared!")
		}
		return nil
	},
}

func IsMP4(file *os.File) (bool, error) {
	magicBytes := make([]byte, 4)
	_, err := file.ReadAt(magicBytes, 4)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	return string(magicBytes[:]) == "ftyp", nil
}

func ClearMetadata(file *os.File, offset int64, length int64) error {
	output.Println(false, "Turning moov.meta box into free space...")
	boxes, err := GetBoxes(file, offset, length)
	if err != nil {
		return err
	}
	var moovBox MoovBox
	for _, b := range boxes {
		if b.GetType() == "moov" {
			moovBox = b.(MoovBox)
			break
		}
	}
	if moovBox.Size == 0 {
		output.Println(false, "Can't find moov box!")
		return nil
	}
	meta := moovBox.FindNestedBoxByType("meta")
	if meta == nil {
		output.Println(false, "Can't find meta box!")
		return nil
	}
	metaOffset, metaData, err := meta.(MetaBox).ToFreeBox()
	if err != nil {
		return err
	}
	_, err = file.WriteAt(metaData, metaOffset)
	if err != nil {
		return fmt.Errorf("error writing changes to file: %w", err)
	}
	return nil
}

func GetBoxes(file *os.File, startOffset int64, length int64) ([]PrintableBox, error) {
	var result []PrintableBox
	stat, _ := file.Stat()
	fileSize := stat.Size()
	for i := startOffset; ; {
		box := Box{
			StartOffset: i,
			File:        file,
		}
		header := make([]byte, 8)
		_, err := file.ReadAt(header, i)
		if err == io.EOF {
			break
		}
		if i >= startOffset+length-8 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading box size: %w", err)
		}
		i += 8
		box.Size = uint64(binary.BigEndian.Uint32(header[0:4]))
		if int64(box.Size) >= fileSize {
			return nil, fmt.Errorf("invalid box size: %d", box.Size)
		}
		box.Type = string(header[4:8])
		for _, c := range box.Type {
			if byte(c) < 97 || byte(c) > 122 {
				return nil, fmt.Errorf("invalid box type: %s", box.Type)
			}
		}
		if box.Size == 1 {
			largeSize := make([]byte, 8)
			_, err := file.ReadAt(largeSize, i)
			if err != nil {
				return nil, fmt.Errorf("error reading box large size: %w", err)
			}
			i += 8
			box.Size = binary.BigEndian.Uint64(largeSize)
		} else if box.Type == "uuid" {
			userType := make([]byte, 16)
			_, err := file.ReadAt(userType, i)
			if err != nil {
				return nil, fmt.Errorf("error reading user type uuid: %w", err)
			}
			i += 16
			box.UUID = string(userType)
		}
		box.StartData = i
		result = append(result, ConvertBox(box))
		if box.Size == 0 {
			break
		}
		i += int64(box.Size) - (i - box.StartOffset)
	}
	return result, nil
}

type Box struct {
	Type        string
	UUID        string
	StartOffset int64
	StartData   int64
	Size        uint64
	File        *os.File
}

func (b Box) GetType() string {
	return b.Type
}

func (b Box) GetData() ([]byte, error) {
	data := make([]byte, b.Size-uint64(b.StartData-b.StartOffset))
	_, err := b.File.ReadAt(data, b.StartData)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading box %s data: %w", b.Type, err)
	}
	return data, nil
}

func (b Box) GetBoxes() []PrintableBox {
	boxes, err := GetBoxes(b.File, b.StartData, int64(b.Size))
	if err != nil {
		return nil
	}
	return boxes
}

func (b Box) FindNestedBoxByType(boxType string) PrintableBox {
	boxes := b.GetBoxes()
	if boxes == nil {
		return nil
	}
	for _, nestedBox := range boxes {
		if nestedBox.GetType() == boxType {
			return nestedBox
		}
	}
	return nil
}

func (b Box) Print() {
	output.Printf(false, "Skip parsing box type %s\n", b.Type)
	boxes := b.GetBoxes()
	if boxes != nil {
		for _, b := range boxes {
			b.Print()
		}
	}
	output.Println(false)
}

func (b Box) ToFreeBox() (int64, []byte, error) {
	data := make([]byte, b.Size)
	_, err := b.File.ReadAt(data, b.StartOffset)
	if err != nil {
		return 0, nil, fmt.Errorf("error reading data: %w", err)
	}
	data[4] = 'f'
	data[5] = 'r'
	data[6] = 'e'
	data[7] = 'e'
	for i := 8; i < len(data); i++ {
		data[i] = 0
	}
	return b.StartOffset, data, nil
}
