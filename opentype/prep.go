package opentype

import (
	"encoding/binary"
	"os"
)

// Prep is a "prep" table.
// The Control Value Program consists of a set of TrueType instructions that will be executed whenever the font or point size or transformation matrix change and before each glyph is interpreted.
type Prep struct {
	Values []uint8
}

func parsePrep(f *os.File, offset, length uint32) (prep *Prep, err error) {
	prep = &Prep{
		Values: make([]uint8, length),
	}
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, prep.Values)
	if err != nil {
		return
	}
	return
}
