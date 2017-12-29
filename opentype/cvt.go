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
