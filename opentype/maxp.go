package opentype

import (
	"encoding/binary"
	"io"
	"os"
)

// Maxp is a "maxp" table.
// This table establishes the memory requirements for this font.
type Maxp struct {
	Version Fixed
	// The number of glyphs in the font.
	NumGlyphs uint16
	// Maximum points in a non-composite glyph.
	MaxPoints uint16
	// Maximum contours in a non-composite glyph.
	MaxContours uint16
	// Maximum points in a composite glyph.
	MaxCompositePoints uint16
	// Maximum contours in a composite glyph.
	MaxCompositeContours uint16
	// 1 if instructions do not use the twilight zone (Z0), or 2 if instructions do use Z0; should be set to 2 in most cases.
	MaxZones uint16
	// Maximum points used in Z0.
	MaxTwilightPoints uint16
	// Number of Storage Area locations.
	MaxStorage uint16
	// Number of FDEFs, equal to the highest function number + 1.
	MaxFunctionDefs uint16
	// Number of IDEFs.
	MaxInstructionDefs uint16
	// Maximum stack depth.
	MaxStackElements uint16
	// Maximum byte count for glyph instructions.
	MaxSizeOfInstructions uint16
	// Maximum number of components referenced at “top level” for any composite glyph.
	MaxComponentElements uint16
	// Maximum levels of recursion; 1 for simple components.
	MaxComponentDepth uint16
}

func parseMaxp(f *os.File, offset uint32) (m *Maxp, err error) {
	m = &Maxp{}
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, &(m.Version))
	if err != nil {
		return
	}
	// version 0.5
	if 0x00005000 == m.Version {
		err = binary.Read(f, binary.BigEndian, &(m.NumGlyphs))
		return
	}
	// version 1.0
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, m)
	return
}

// Tag is table name.
func (m *Maxp) Tag() Tag {
	return String2Tag("maxp")
}

// Store writes binary expression of this table.
func (m *Maxp) Store(w io.Writer) (err error) {
	err = bWrite(w, m)
	if err != nil {
		return
	}
	return padSpace(w, m.Length())
}

// CheckSum for this table.
func (m *Maxp) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(m)
}

// Length returns the size(byte) of this table.
func (m *Maxp) Length() uint32 {
	// version 0.5
	if 0x00005000 == m.Version {
		return uint32(6)
	}
	return uint32(32)
}
