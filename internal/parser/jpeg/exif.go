package jpeg

import (
	"encoding/binary"
	"fmt"
)

type IFD struct {
	StartOffset uint32
	Tags        map[uint16]string
}

func ParseExif(raw []byte) []IFD {
	byteOrder := ByteOrder{
		byteOrder: raw[0:2],
	}
	offsetIFD := byteOrder.getUint32(raw[4:8])
	var result []IFD
	ifd := IFD{}
	for {
		ifd.StartOffset = offsetIFD
		links, next := ParseIFD(&ifd, raw, byteOrder)
		result = append(result, ifd)
		for _, link := range links {
			ifd = IFD{
				StartOffset: link,
			}
			_, _ = ParseIFD(&ifd, raw, byteOrder)
			result = append(result, ifd)
		}
		if next == 0 {
			break
		}
		offsetIFD = next
		ifd = IFD{}
	}
	return result
}

func ParseIFD(ifd *IFD, raw []byte, byteOrder ByteOrder) ([]uint32, uint32) {
	i := ifd.StartOffset
	numberOfTags := int(byteOrder.getUint16(raw[i : i+2]))
	i += 2
	links := make([]uint32, 0)
	ifd.Tags = make(map[uint16]string)
	for c := 0; c < numberOfTags; c++ {
		tagId := byteOrder.getUint16(raw[i : i+2])
		i += 2
		tagType := byteOrder.getUint16(raw[i : i+2])
		i += 2
		valueCount := byteOrder.getUint32(raw[i : i+4])
		i += 4
		valueOffset := byteOrder.getUint32(raw[i : i+4])
		i += 4
		if tagId == 0x8769 {
			links = append(links, valueOffset)
		} else {
			if tagType == 2 || tagType == 129 {
				end := valueOffset + valueCount - 1
				if int(end) > len(raw) {
					continue
				}
				ifd.Tags[tagId] = string(raw[valueOffset:end])
			} else {
				ifd.Tags[tagId] = fmt.Sprintf("%v", valueOffset)
			}
		}
	}
	next := byteOrder.getUint32(raw[i : i+4])
	return links, next
}

type ByteOrder struct {
	byteOrder []byte
}

func (o *ByteOrder) getUint16(value []byte) uint16 {
	if o.byteOrder[0] == 0x49 && o.byteOrder[1] == 0x49 {
		return binary.LittleEndian.Uint16(value)
	} else if o.byteOrder[0] == 0x4D && o.byteOrder[1] == 0x4D {
		return binary.BigEndian.Uint16(value)
	}
	fmt.Printf("Invalid byte order: %v", o.byteOrder)
	return 0
}

func (o *ByteOrder) getUint32(value []byte) uint32 {
	if o.byteOrder[0] == 0x49 && o.byteOrder[1] == 0x49 {
		return binary.LittleEndian.Uint32(value)
	} else if o.byteOrder[0] == 0x4D && o.byteOrder[1] == 0x4D {
		return binary.BigEndian.Uint32(value)
	}
	fmt.Printf("Invalid byte order: %v", o.byteOrder)
	return 0
}