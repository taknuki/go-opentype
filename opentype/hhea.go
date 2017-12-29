package opentype

import (
	"encoding/binary"
	"os"
)

// Hhea is a "hhea" table.
// This table contains information for horizontal layout.
type Hhea struct {
	MajorVersion        uint16
	MinorVersion        uint16
	Ascender            int16
	Descender           int16
	LineGap             int16
	AdvanceWidthMax     uint16
	MinLeftSideBearing  int16
	MinRightSideBearing int16
	XMaxExtent          int16
	CaretSlopeRise      int16
	CaretSlopeRun       int16
	CaretOffset         int16
	Reserved1           int16
	Reserved2           int16
	Reserved3           int16
	Reserved4           int16
	MetricDataFormat    int16
	NumberOfHMetrics    uint16
}

func parseHhea(f *os.File, offset uint32) (h *Hhea, err error) {
	h = &Hhea{}
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, h)
	return
}
