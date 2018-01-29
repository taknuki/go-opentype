package opentype

import (
	"encoding/binary"
	"os"
)

// Cvt is a "cvt" table.
// This table contains a list of values that can be referenced by instructions.
type Cvt struct {
	Values []int16
}

func parseCvt(f *os.File, offset, length uint32) (c *Cvt, err error) {
	size := length / 2
	c = &Cvt{
		Values: make([]int16, size),
	}
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, c.Values)
	if err != nil {
		return
	}
	return
}

// Tag is table name.
func (c *Cvt) Tag() Tag {
	return String2Tag("cvt ")
}

// store writes binary expression of this table.
func (c *Cvt) store(w *errWriter) {
	for _, v := range c.Values {
		w.write(&(v))
	}
	padSpace(w, c.Length())
}

// CheckSum for this table.
func (c *Cvt) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(c)
}

// Length returns the size(byte) of this table.
func (c *Cvt) Length() uint32 {
	return uint32(2 * len(c.Values))
}

// Exists returns true if this is not nil.
func (c *Cvt) Exists() bool {
	return c != nil
}
