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

// Tag is table name.
func (prep *Prep) Tag() Tag {
	return String2Tag("prep")
}

// store writes binary expression of this table.
func (prep *Prep) store(w *errWriter) {
	for _, v := range prep.Values {
		w.write(&(v))
	}
	padSpace(w, prep.Length())
}

// CheckSum for this table.
func (prep *Prep) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(prep)
}

// Length returns the size(byte) of this table.
func (prep *Prep) Length() uint32 {
	return uint32(len(prep.Values))
}

// Exists returns true if this is not nil.
func (prep *Prep) Exists() bool {
	return prep != nil
}
