package opentype

import (
	"encoding/binary"
	"os"
)

// Fpgm is a "fpgm" table.
// This table is similar to the CVT Program, except that it is only run once, when the font is first used. It is used only for FDEFs and IDEFs.
type Fpgm struct {
	Values []uint8
}

func parseFpgm(f *os.File, offset, length uint32) (fpgm *Fpgm, err error) {
	fpgm = &Fpgm{
		Values: make([]uint8, length),
	}
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, fpgm.Values)
	if err != nil {
		return
	}
	return
}
