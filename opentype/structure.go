package opentype

import (
	"encoding/binary"
	"os"
)

// Fixed is a 32-bit signed fixed-point number (16.16)
type Fixed int32

// LongDateTime is a Date represented in number of seconds since 12:00 midnight, January 1, 1904. The value is represented as a signed 64-bit integer.
type LongDateTime int64

// Tag is a array of four uint8s (length = 32 bits) used to identify a script, language system, feature, or baseline
type Tag uint32

func (t Tag) String() string {
	bytes := []byte{
		byte(t >> 24 & 0xFF),
		byte(t >> 16 & 0xFF),
		byte(t >> 8 & 0xFF),
		byte(t & 0xFF),
	}
	return string(bytes)
}

const (
	// SfntVersionTrueTypeOpenType : OpenType fonts that contain TrueType outlines.
	SfntVersionTrueTypeOpenType = Tag(0x00010000)
	// SfntVersionCFFOpenType  : OpenType fonts containing CFF data.
	SfntVersionCFFOpenType = Tag(0x4F54544F) // OTTO
	// SfntVersionAppleTrueType : The Apple specification for TrueType font.
	SfntVersionAppleTrueType = Tag(0x74727565) // true
	// SfntVersionAppleType1 : The Apple specification for the old style of PostScript font.
	SfntVersionAppleType1 = Tag(0x74797031) // typ1
	// SfntVersionTTCHeader : The header of TTC format file.
	SfntVersionTTCHeader = Tag(0x74746366) // ttcf
)

func parseSfntVersion(f *os.File, offset int64) (t Tag, err error) {
	f.Seek(offset, 0)
	err = binary.Read(f, binary.BigEndian, &t)
	f.Seek(offset, 0)
	return
}

// OffsetTable is the first table of OpenType font file.
type OffsetTable struct {
	SfntVersion   Tag
	NumTables     uint16
	SearchRange   uint16
	EntrySelector uint16
	RangeShift    uint16
}

func parseOffsetTable(f *os.File) (ot *OffsetTable, err error) {
	ot = &OffsetTable{}
	err = binary.Read(f, binary.BigEndian, ot)
	return
}

// TableRecord is a OpenType table.
type TableRecord struct {
	Tag      Tag
	CheckSum uint32
	Offset   uint32
	Length   uint32
}

func parseTableRecord(f *os.File, numTables uint16) (trs map[string]*TableRecord, err error) {
	trs = make(map[string]*TableRecord)
	for i := uint16(0); i < numTables; i++ {
		tr := &TableRecord{}
		err = binary.Read(f, binary.BigEndian, tr)
		if err != nil {
			return
		}
		trs[tr.Tag.String()] = tr
	}
	return
}
