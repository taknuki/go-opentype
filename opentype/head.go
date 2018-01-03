package opentype

import (
	"encoding/binary"
	"fmt"
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

func parseHead(f *os.File, offset, checkSum uint32) (h *Head, err error) {
	h = &Head{}
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, h)
	if err == nil {
		actual, err := h.CheckSum()
		if err == nil && checkSum != actual {
			err = fmt.Errorf("Table head has invalid checksum, expected:%d actual:%d", checkSum, actual)
		}
	}
	return
}

// Tag is table name.
func (h *Head) Tag() Tag {
	return String2Tag("head")
}

// store writes binary expression of this table.
func (h *Head) store(w *errWriter) {
	w.write(h)
	padSpace(w, h.Length())
}

// CheckSum for this table.
func (h *Head) CheckSum() (checkSum uint32, err error) {
	bk := h.CheckSumAdjustment
	h.CheckSumAdjustment = 0
	checkSum, err = simpleCheckSum(h)
	h.CheckSumAdjustment = bk
	return
}

// Length returns the size(byte) of this table.
func (h *Head) Length() uint32 {
	return uint32(54)
}
