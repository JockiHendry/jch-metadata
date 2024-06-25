package mkv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type EBMLElement struct {
	ElementID []byte
	Size      uint64
	StartAt   int64
	DataAt    int64
	File      *os.File
	Elements  []EBMLElement
}

var EmptyElement = EBMLElement{}

func (e *EBMLElement) FindFirstElement(elementId []byte, value []byte) *EBMLElement {
	fileOffset := e.DataAt
	if e.Elements != nil {
		for _, child := range e.Elements {
			if bytes.Equal(child.ElementID, elementId) {
				if value == nil {
					return &child
				} else {
					data, _ := child.GetBytes()
					if bytes.Equal(data, value) {
						return &child
					}
				}
			}
			fileOffset = child.DataAt + int64(child.Size)
		}
	}
	limit := e.DataAt + int64(e.Size)
	for fileOffset < limit {
		var element *EBMLElement
		var err error
		element, fileOffset, err = NewEBMLElement(e.File, fileOffset)
		if err != nil {
			return nil
		}
		if bytes.Equal(element.ElementID, elementId) {
			if value == nil {
				e.Elements = append(e.Elements, *element)
				return element
			} else {
				data, _ := element.GetBytes()
				if bytes.Equal(data, value) {
					e.Elements = append(e.Elements, *element)
					return element
				}
			}
		}
	}
	return &EmptyElement
}

func (e *EBMLElement) GetElements() []EBMLElement {
	if e.Elements != nil {
		return e.Elements
	} else {
		elements, _ := GetEBMLElements(e.File, e.DataAt, e.DataAt+int64(e.Size), 9999)
		e.Elements = elements
		return elements
	}
}

func (e *EBMLElement) GetBytes() ([]byte, error) {
	data := make([]byte, e.Size)
	_, err := e.File.ReadAt(data, e.DataAt)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func (e *EBMLElement) StringValue() (string, error) {
	if e.Size == 0 {
		return "", nil
	}
	if data, err := e.GetBytes(); err == nil {
		return string(data), nil
	} else {
		return "", err
	}
}

func (e *EBMLElement) UintValue() (uint64, error) {
	if e.Size == 0 {
		return 0, nil
	}
	if data, err := e.GetBytes(); err == nil {
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
	} else {
		return 0, err
	}
}

func (e *EBMLElement) DateValue() (time.Time, error) {
	if e.Size == 0 {
		return time.Unix(0, 0), nil
	}
	if data, err := e.GetBytes(); err == nil {
		dateValue := binary.BigEndian.Uint64(data)
		if dateValue == 0 {
			return time.Unix(0, 0), nil
		}
		return time.Unix(978307200, int64(dateValue)), nil
	} else {
		return time.Unix(0, 0), fmt.Errorf("failed to read file: %w", err)
	}
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

func GetEBMLElements(file *os.File, fileOffset int64, limit int64, count int) ([]EBMLElement, error) {
	var result []EBMLElement
	c := 0

	for fileOffset < limit && c < count {
		var element *EBMLElement
		var err error
		element, fileOffset, err = NewEBMLElement(file, fileOffset)
		if err != nil {
			return nil, err
		}
		result = append(result, *element)
		c += 1
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

func NewEBMLElement(file *os.File, fileOffset int64) (*EBMLElement, int64, error) {
	elementIdData := make([]byte, 8)
	_, err := file.ReadAt(elementIdData, fileOffset)
	if err != nil {
		return nil, fileOffset, fmt.Errorf("failed to read file: %w", err)
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
		return nil, fileOffset, fmt.Errorf("failed to read file: %w", err)
	}
	size, offset := GetVSize(sizeData[:])
	element.Size = size
	fileOffset += int64(offset)
	element.DataAt = fileOffset

	return &element, fileOffset + int64(size), nil
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
