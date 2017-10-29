package opentype

import (
	"fmt"
	"os"
)

// Font is the opentype font.
type Font struct {
	offsetTable *OffsetTable
	CMap        *CMap
}

// ParseFont returns the Font instance from the font file.
func ParseFont(f *os.File) (*Font, error) {
	return parseFont(f, 0)
}

func parseFont(f *os.File, offset int64) (*Font, error) {
	sfntVersion, err := parseSfntVersion(f, offset)
	if err != nil {
		return nil, err
	}
	switch sfntVersion {
	case SfntVersionTrueTypeOpenType:
		return parseTrueTypeFont(f)
	case SfntVersionAppleTrueType:
		return parseTrueTypeFont(f)
	case SfntVersionCFFOpenType:
		return parseCFFFont(f)
	default:
		return nil, fmt.Errorf("%s is not supported SFNT Version", sfntVersion)
	}
}

func parseTrueTypeFont(f *os.File) (font *Font, err error) {
	font, err = parseCommonTable(f)
	// TODO TrueType specific
	return
}

func parseCFFFont(f *os.File) (font *Font, err error) {
	font, err = parseCommonTable(f)
	// TODO CFF specific
	return
}

func parseCommonTable(f *os.File) (font *Font, err error) {
	font = &Font{}
	font.offsetTable, err = parseOffsetTable(f)
	if err != nil {
		return
	}
	trs, err := parseTableRecord(f, font.offsetTable.NumTables)
	if err != nil {
		return
	}
	tr, err := getTableRecord(trs, "cmap")
	if err != nil {
		return
	}
	font.CMap, err = parseCMap(f, tr)
	if err != nil {
		return
	}
	return
}
