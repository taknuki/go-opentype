package opentype

import (
	"encoding/binary"
	"io"
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

// Tag is table name.
func (fpgm *Fpgm) Tag() Tag {
	return String2Tag("fpgm")
}

// Store writes binary expression of this table.
func (fpgm *Fpgm) Store(w io.Writer) (err error) {
	for _, v := range fpgm.Values {
		err = bWrite(w, &(v))
		if err != nil {
			return
		}
	}
	return padSpace(w, fpgm.Length())
}

// CheckSum for this table.
func (fpgm *Fpgm) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(fpgm)
}

// Length returns the size(byte) of this table.
func (fpgm *Fpgm) Length() uint32 {
	return uint32(len(fpgm.Values))
}
