package opentype

import (
	"encoding/binary"
	"os"
)

// Glyf is a "glyf" table.
// This table contains information that describes the glyphs in the font in the TrueType outline format.
type Glyf struct {
	data [][]byte
}

func parseGlyf(f *os.File, offset, length uint32, l *Loca) (g *Glyf, err error) {
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	g = &Glyf{
		data: make([][]byte, l.Len()),
	}
	for i := 0; i < l.Len(); i++ {
		first := l.Get(i)
		var last uint32
		if i >= l.Len()-1 {
			last = length
		} else {
			last = l.Get(i + 1)
		}
		g.data[i] = make([]byte, last-first)
		err = binary.Read(f, binary.BigEndian, g.data[i])
		if err != nil {
			return
		}
	}
	return
}

func (g *Glyf) filter(f []uint16) (new *Glyf) {
	new = &Glyf{
		data: make([][]byte, len(f)),
	}
	for i, gid := range f {
		new.data[i] = g.data[gid]
	}
	return
}

func (g *Glyf) generateLoca() (l *Loca) {
	l = &Loca{
		indexToLocFormat: 1,
		offsetsLong:      make([]uint32, len(g.data)+1),
	}
	offset := uint32(0)
	for i, d := range g.data {
		l.offsetsLong[i] = offset
		offset += uint32(len(d))
	}
	// extra entry
	l.offsetsLong[len(g.data)] = offset
	return
}

// Tag is table name.
func (g *Glyf) Tag() Tag {
	return String2Tag("glyf")
}

// store writes binary expression of this table.
func (g *Glyf) store(w *errWriter) {
	for _, d := range g.data {
		w.writeBin(d)
	}
	padSpace(w, g.Length())
}

// CheckSum for this table.
func (g *Glyf) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(g)
}

// Length returns the size(byte) of this table.
func (g *Glyf) Length() uint32 {
	l := uint32(0)
	for _, d := range g.data {
		l += uint32(len(d))
	}
	return l
}

// Exists returns true if this is not nil.
func (g *Glyf) Exists() bool {
	return g != nil
}
