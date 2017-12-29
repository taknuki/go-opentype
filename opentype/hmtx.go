package opentype

import (
	"encoding/binary"
	"os"
)

// LongHorMetric is a paired advance width and left side bearing values for each glyph.
type LongHorMetric struct {
	// Advance width, in font design units.
	AdvanceWidth uint16
	// Glyph left side bearing, in font design units.
	Lsb int16
}

// Hmtx is a "hmtx" table.
// The horizontal metrics ('hmtx') table provides glyph advance widths and left side bearings.
type Hmtx struct {
	// Paired advance width and left side bearing values for each glyph. Records are indexed by glyph ID.
	HMetrics []*LongHorMetric
	// Left side bearings for glyph IDs greater than or equal to numberOfHMetrics.
	LeftSideBearings []int16
}

func parseHmtx(f *os.File, offset uint32, numGlyphs, numberOfHMetrics uint16) (h *Hmtx, err error) {
	h = &Hmtx{}
	f.Seek(int64(offset), 0)
	h.HMetrics = make([]*LongHorMetric, int(numberOfHMetrics))
	for i := 0; i < int(numberOfHMetrics); i++ {
		m := &LongHorMetric{}
		err = binary.Read(f, binary.BigEndian, m)
		if err != nil {
			return
		}
		h.HMetrics[i] = m
	}
	h.LeftSideBearings = make([]int16, int(numGlyphs-numberOfHMetrics))
	for i := 0; i < int(numGlyphs-numberOfHMetrics); i++ {
		err = binary.Read(f, binary.BigEndian, &(h.LeftSideBearings[i]))
		if err != nil {
			return
		}
	}
	return
}
