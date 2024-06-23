package mp4

import (
	"encoding/binary"
	"fmt"
	"jch-metadata/internal/output"
)

type MetaBox struct {
	*Box
}

func (b MetaBox) Print() {
	output.Println(false, "Metadata (meta)")
	handlerBox := b.FindNestedBoxByType("hdlr")
	if handlerBox == nil {
		output.Printf(false, "Failed to retrieve handler for meta box\n")
		return
	}
	handler, err := handlerBox.(HandlerBox).GetHandler()
	if err != nil {
		output.Printf(false, "Failed to retrieve handler for meta box: %s\n", err)
		return
	}
	if handler.Type == "mdta" {
		mdta, err := b.GetMdta()
		if err != nil {
			output.Printf(false, "Failed to parse mdta: %s\n", err)
		}
		width := 30
		for k := range mdta {
			if len(k) > width {
				width = len(k)
			}
		}
		for k, v := range mdta {
			output.PrintForm(false, k, v.String(), width)
		}
		output.Println(false)
	} else {
		output.Printf(false, "Skip parsing unsupported handler %s\n", handler.Type)
	}

	boxes := b.GetBoxes()
	if boxes == nil {
		output.Println(false, "Empty meta box")
	} else {
		for _, b := range boxes {
			if b.GetType() != "mdta" && b.GetType() != "hdlr" && b.GetType() != "keys" && b.GetType() != "ilst" {
				output.Printf(false, "Skip parsing box %s\n", b.GetType())
			}
		}
	}
	output.Println(false)
}

func (b MetaBox) GetMdta() (map[string]KeyValue, error) {
	boxes := b.GetBoxes()
	if boxes == nil {
		return nil, fmt.Errorf("empty meta box")
	}
	result := make(map[string]KeyValue)
	keys, err := b.GetKeys()
	if err != nil {
		return nil, err
	}
	values, err := b.GetValues()
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		result[keys[v.Index-1]] = v
	}
	return result, nil
}

type Handler struct {
	Type string
	Name string
}

func (b MetaBox) GetKeys() ([]string, error) {
	box := b.FindNestedBoxByType("keys")
	if box == nil {
		return nil, fmt.Errorf("keys box not found")
	}
	data, err := box.GetData()
	if err != nil {
		return nil, err
	}
	entryCount := int(binary.BigEndian.Uint32(data[4:8]))
	startOffset := 8
	var result []string
	for i := 1; i <= entryCount; i++ {
		size := binary.BigEndian.Uint32(data[startOffset : startOffset+4])
		start := startOffset + 8
		end := start + int(size) - 8
		value := string(data[start:end])
		result = append(result, value)
		startOffset += int(size)
	}
	return result, nil
}

func (b MetaBox) GetValues() ([]KeyValue, error) {
	box := b.FindNestedBoxByType("ilst")
	if box == nil {
		return nil, fmt.Errorf("ilst box not found")
	}
	data, err := box.GetData()
	if err != nil {
		return nil, err
	}
	var result []KeyValue
	for i := 0; i < len(data); {
		size := binary.BigEndian.Uint32(data[i : i+4])
		dataSize := binary.BigEndian.Uint32(data[i+8 : i+12])
		if string(data[i+12:i+16]) != "data" {
			return nil, fmt.Errorf("invalid data identifier: %s", string(data[i+12:i+16]))
		}
		keyValue := KeyValue{
			Index:           binary.BigEndian.Uint32(data[i+4 : i+8]),
			Type:            binary.BigEndian.Uint32(data[i+16 : i+20]),
			LocaleIndicator: binary.BigEndian.Uint32(data[i+20 : i+24]),
			Value:           data[i+24 : i+24+int(dataSize)-16],
		}
		result = append(result, keyValue)
		i += int(size)
	}
	return result, nil
}

type KeyValue struct {
	Index           uint32
	Type            uint32
	LocaleIndicator uint32
	Value           []byte
}

func (k KeyValue) String() string {
	if k.Type <= 5 {
		return string(k.Value[:])
	} else {
		return fmt.Sprintf("%v", binary.BigEndian.Uint32(k.Value[:]))
	}
}

type HandlerBox struct {
	*Box
}

func (b HandlerBox) GetHandler() (*Handler, error) {
	data, err := b.GetData()
	if err != nil {
		return nil, err
	}
	return &Handler{
		Type: string(data[8:12]),
		Name: string(data[24:]),
	}, nil
}

func (b HandlerBox) Print() {
	handler, err := b.GetHandler()
	if err != nil {
		output.Printf(false, "Error retrieving handler data: %s\n", err)
		return
	}
	output.PrintForm(false, "Handler", fmt.Sprintf("%s (%s)", handler.Type, handler.Name), 20)
}
