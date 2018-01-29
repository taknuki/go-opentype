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
// Left side bearings: The horizontal distance from the base position to the left edge of the glyph's bounding box.
// Adovance width: The horizontal distance from the current base position and the next base position.
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

func (h *Hmtx) filter(f []uint16) *Hmtx {
	th := uint16(len(h.HMetrics))
	numHM := 0
	for i, gid := range f {
		if gid < th {
			numHM = i + 1
		}
	}
	new := &Hmtx{
		HMetrics:         make([]*LongHorMetric, 0, numHM),
		LeftSideBearings: make([]int16, 0, len(f)-numHM),
	}
	for i, gid := range f {
		if i < numHM {
			if gid < th {
				new.HMetrics = append(new.HMetrics, h.HMetrics[gid])
			} else {
				new.HMetrics = append(new.HMetrics, &LongHorMetric{
					AdvanceWidth: 0,
					Lsb:          h.LeftSideBearings[gid-th],
				})
			}
		} else {
			new.LeftSideBearings = append(new.LeftSideBearings, h.LeftSideBearings[gid-th])
		}
	}
	return new
}

// Tag is table name.
func (h *Hmtx) Tag() Tag {
	return String2Tag("hmtx")
}

// store writes binary expression of this table.
func (h *Hmtx) store(w *errWriter) {
	for _, hm := range h.HMetrics {
		w.write(hm)
	}
	for _, lsb := range h.LeftSideBearings {
		w.write(&(lsb))
	}
	padSpace(w, h.Length())
}

// CheckSum for this table.
func (h *Hmtx) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(h)
}

// Length returns the size(byte) of this table.
func (h *Hmtx) Length() uint32 {
	return uint32(4*len(h.HMetrics) + 2*len(h.LeftSideBearings))
}

// Exists returns true if this is not nil.
func (h *Hmtx) Exists() bool {
	return h != nil
}
