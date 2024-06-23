package mp4

import (
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/output"
)

type TrakBox struct {
	*Box
}

func (b TrakBox) Print() {
	output.PrintHeader(false, "Track (trak)")
	for _, box := range b.GetBoxes() {
		if box.GetType() == "edts" {
			continue
		}
		box.Print()
	}
	output.Println(false)
}

type TrackHeaderBox struct {
	*Box
}

func (b TrackHeaderBox) Print() {
	data, err := b.GetData()
	if err != nil {
		output.Printf(false, "Error retrieving track header: %s\n", err)
		return
	}
	var trackId uint32
	if data[0] == 1 {
		trackId = binary.BigEndian.Uint32(data[20:24])
	} else {
		trackId = binary.BigEndian.Uint32(data[12:16])
	}
	var flag string
	if data[3]&0x1 == 0 {
		flag = "Disabled"
	} else if data[3]&0x1 == 1 {
		flag = "Enabled"
	}
	if data[3]&0x2 == 1 {
		flag += ", Used"
	}
	if data[3]&0x4 == 1 {
		flag += ", Preview"
	}
	output.PrintForm(false, "Track ID", fmt.Sprintf("%d", trackId), 20)
	output.PrintForm(false, "Flag", flag, 20)
}

type TrackMediaBox struct {
	*Box
}

func (b TrackMediaBox) Print() {
	for _, box := range b.GetBoxes() {
		box.Print()
	}
}

type TrackMediaHeader struct {
	MvhdBox
}

func (b TrackMediaHeader) Print() {
	header, err := b.GetHeader()
	if err != nil {
		output.Printf(false, "Error retrieving media header: %s\n", err)
		return
	}
	output.PrintForm(false, "Creation Time", header.LocalCreationTime(), 20)
	output.PrintForm(false, "Modification Time", header.LocalModificationTime(), 20)
	output.PrintForm(false, "Duration", header.DurationString(), 20)
}

type Media struct {
	*Header
	Handler string
}

type MediaInformationBox struct {
	*Box
}

func (b MediaInformationBox) Print() {
	for _, box := range b.GetBoxes() {
		if box.GetType() == "nmhd" || box.GetType() == "stbl" || box.GetType() == "smhd" || box.GetType() == "vmhd" {
			continue
		}
		box.Print()
	}
}

type DataInformationBox struct {
	*Box
}

func (b DataInformationBox) Print() {
	for _, box := range b.GetBoxes() {
		data, err := box.GetData()
		if err != nil {
			output.Printf(false, "Error retrieving data information: %s", err)
		}
		if box.GetType() == "dref" {
			numberOfEntries := binary.BigEndian.Uint32(data[4:8])
			offset := 8
			for i := 0; i < int(numberOfEntries); i++ {
				size := binary.BigEndian.Uint32(data[offset : offset+4])
				refType := string(data[offset+4 : offset+8])
				value := string(data[offset+12 : offset+int(size)])
				output.PrintForm(false, "Ref", fmt.Sprintf("%s %s", refType, value), 20)
				offset += int(size)
			}
		} else {
			box.Print()
		}

	}
}
