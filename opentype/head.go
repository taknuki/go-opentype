package opentype

import (
	"encoding/binary"
	"os"
)

// Head is a "head" table.
type Head struct {
	MajorVersion       uint16
	MinorVersion       uint16
	FontRevision       Fixed
	CheckSumAdjustment uint32
	MagicNumber        uint32
	Flags              uint16
	UnitsPerEm         uint16
	Created            LongDateTime
	Modified           LongDateTime
	XMin               int16
	YMin               int16
	XMax               int16
	YMax               int16
	MacStyle           uint16
	LowestRecPPEM      uint16
	FontDirectionHint  int16
	IndexToLocFormat   int16
	GlyphDataFormat    int16
}

func parseHead(f *os.File, offset uint32) (h *Head, err error) {
	h = &Head{}
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, h)
	return
}
