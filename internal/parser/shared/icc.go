package shared

import (
	"bytes"
	"encoding/binary"
	"jch-metadata/internal/output"
)

func ParseICC(raw []byte) *Profile {
	result := Profile{
		CmmType:            string(raw[4:8]),
		ProfileClass:       string(raw[12:16]),
		PrimaryPlatform:    string(raw[40:44]),
		DeviceManufacturer: string(raw[48:52]),
		DeviceModel:        string(raw[52:56]),
		ProfileCreator:     string(raw[80:83]),
	}
	tagCount := binary.BigEndian.Uint32(raw[128:132])
	offset := 132
	for i := 0; i < int(tagCount); i++ {
		if bytes.Equal(raw[offset:offset+4], []byte{0x63, 0x70, 0x72, 0x74}) {
			dataOffset := binary.BigEndian.Uint32(raw[offset+4 : offset+8])
			size := binary.BigEndian.Uint32(raw[offset+8 : offset+12])
			if string(raw[dataOffset:dataOffset+4]) == "text" {
				result.Copyright = string(raw[dataOffset+8 : dataOffset+size-1])
			} else if string(raw[dataOffset:dataOffset+4]) == "mluc" {
				result.Copyright = string(raw[dataOffset+28 : dataOffset+size-1])
			}
			break
		}
		offset += 12
	}
	return &result
}

type Profile struct {
	CmmType            string
	ProfileClass       string
	PrimaryPlatform    string
	DeviceManufacturer string
	DeviceModel        string
	ProfileCreator     string
	Copyright          string
}

func PrintICC(indented bool, profile *Profile) {
	if profile == nil {
		return
	}
	output.PrintHeader(indented, "ICC Profile")
	output.PrintForm(indented, "CMM Type", profile.CmmType, 18)
	output.PrintForm(indented, "Profile Class", profile.ProfileClass, 18)
	output.PrintForm(indented, "Primary Platform", profile.PrimaryPlatform, 18)
	output.PrintForm(indented, "Dev Manufacturer", profile.DeviceManufacturer, 18)
	output.PrintForm(indented, "Dev Model", profile.DeviceModel, 18)
	output.PrintForm(indented, "Profile Creator", profile.ProfileCreator, 18)
	output.PrintForm(indented, "Copyright", profile.Copyright, 18)
}
