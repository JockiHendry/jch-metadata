package shared

import (
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/output"
	"sort"
)

type IFD struct {
	StartOffset uint32
	Tags        map[uint16]string
}

func ParseExif(raw []byte) []IFD {
	if string(raw[0:4]) != "Exif" {
		return nil
	}
	if raw[4] != 0x00 && raw[5] != 0x00 {
		return nil
	}
	byteOrder := ByteOrder{
		byteOrder: raw[6:8],
	}
	if byteOrder.getUint16(raw[8:10]) != 0x002A {
		return nil
	}
	offsetIFD := byteOrder.getUint32(raw[10:14])
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
	i := ifd.StartOffset + 6
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
				start := valueOffset + 6
				end := start + valueCount - 1
				if int(end) > len(raw) {
					continue
				}
				ifd.Tags[tagId] = string(raw[start:end])
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

func PrintExif(indented bool, ifds []IFD) {
	if ifds == nil || len(ifds) == 0 {
		return
	}
	for _, ifd := range ifds {
		output.PrintHeader(indented, "EXIF IFD Offset 0x%0X", ifd.StartOffset)
		tags := make([]uint16, len(ifd.Tags))
		i := 0
		for k := range ifd.Tags {
			tags[i] = k
			i++
		}
		sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
		for _, t := range tags {
			output.PrintForm(indented, fmt.Sprintf("0x%04X", t), ifd.Tags[t], 10)
		}
		output.Println(indented)
	}
}
