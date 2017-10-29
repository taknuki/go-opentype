package opentype

import (
	"encoding/binary"
	"os"
)

// IsFontCollection returns true if the file is FontCollection format.
func IsFontCollection(f *os.File) (bool, error) {
	sfntVersion, err := parseSfntVersion(f, 0)
	if err != nil {
		return false, err
	}
	if sfntVersion == SfntVersionTTCHeader {
		return true, nil
	}
	return false, nil
}

// FontCollection packages multiple OpenType fonts in a single file structure.
type FontCollection struct {
	header *ttcHeader
	Fonts  []*Font
}

// ParseFontCollections returns the FontCollection instance from the font file.
func ParseFontCollections(f *os.File) (fc *FontCollection, err error) {
	fc = &FontCollection{}
	fc.header, err = parseTTCHeader(f)
	if err != nil {
		return
	}
	fc.Fonts = make([]*Font, fc.header.numFonts)
	for i, o := range fc.header.offsetTable {
		fc.Fonts[i], err = parseFont(f, int64(o))
		if err != nil {
			return
		}
	}
	return
}

// ttcHeader is The header of TTC format file.
type ttcHeader struct {
	sfntVersion  Tag
	majorVersion uint16
	minorVersion uint16
	offsetTable  []uint32
	numFonts     uint32
}

func parseTTCHeader(f *os.File) (h *ttcHeader, err error) {
	h = &ttcHeader{}
	err = binary.Read(f, binary.BigEndian, &(h.sfntVersion))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(h.majorVersion))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(h.minorVersion))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(h.numFonts))
	if err != nil {
		return
	}
	h.offsetTable = make([]uint32, h.numFonts)
	for i := uint32(0); i < h.numFonts; i++ {
		err = binary.Read(f, binary.BigEndian, &(h.offsetTable[i]))
		if err != nil {
			return
		}
	}
	return
}
