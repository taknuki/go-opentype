package opentype

import (
	"encoding/binary"
	"os"
)

// Loca is a "loca" table.
// The indexToLoc table stores the offsets to the locations of the glyphs in the font, relative to the beginning of the glyphData table.
type Loca struct {
	indexToLocFormat int16
	offsetsShort     []uint16
	offsetsLong      []uint32
}

// IsShort returns true if this table is short version.
func (l *Loca) IsShort() bool {
	return 0 == l.indexToLocFormat
}

// IsLong returns true if this table is long version.
func (l *Loca) IsLong() bool {
	return !l.IsShort()
}

// Len returns the length of the index.
func (l *Loca) Len() int {
	if l.IsShort() {
		return len(l.offsetsShort)
	}
	return len(l.offsetsLong)
}

// Get returns the offset.
func (l *Loca) Get(i int) (offset uint32) {
	if l.IsShort() {
		return 2 * (uint32)(l.offsetsShort[i])
	}
	return l.offsetsLong[i]
}

func parseLoca(f *os.File, offset uint32, numGlyphs uint16, indexToLocFormat int16) (l *Loca, err error) {
	size := numGlyphs + 1
	l = &Loca{
		indexToLocFormat: indexToLocFormat,
	}
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	if l.IsShort() {
		l.offsetsShort = make([]uint16, size)
		err = binary.Read(f, binary.BigEndian, l.offsetsShort)
	} else {
		l.offsetsLong = make([]uint32, size)
		err = binary.Read(f, binary.BigEndian, l.offsetsLong)
	}
	return
}