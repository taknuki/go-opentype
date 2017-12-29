package opentype

import (
	"fmt"
	"os"
)

// Font is the opentype font.
type Font struct {
	offsetTable  *OffsetTable
	tableRecords map[string]*TableRecord
	CMap         *CMap
	Head         *Head
	Hhea         *Hhea
	Maxp         *Maxp
	Hmtx         *Hmtx
}

func (f *Font) getTableRecord(tag string) (*TableRecord, error) {
	tr, ok := f.tableRecords[tag]
	if !ok {
		return nil, fmt.Errorf("%s table recors is not found", tag)
	}
	return tr, nil
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
	font.tableRecords, err = parseTableRecord(f, font.offsetTable.NumTables)
	if err != nil {
		return
	}
	cmap, err := font.getTableRecord("cmap")
	if err != nil {
		return
	}
	font.CMap, err = parseCMap(f, cmap.Offset)
	if err != nil {
		return
	}
	head, err := font.getTableRecord("head")
	if err != nil {
		return
	}
	font.Head, err = parseHead(f, head.Offset)
	if err != nil {
		return
	}
	hhea, err := font.getTableRecord("hhea")
	if err != nil {
		return
	}
	font.Hhea, err = parseHhea(f, hhea.Offset)
	if err != nil {
		return
	}
	maxp, err := font.getTableRecord("maxp")
	if err != nil {
		return
	}
	font.Maxp, err = parseMaxp(f, maxp.Offset)
	if err != nil {
		return
	}
	hmtx, err := font.getTableRecord("hmtx")
	if err != nil {
		return
	}
	font.Hmtx, err = parseHmtx(f, hmtx.Offset, font.Maxp.NumGlyphs, font.Hhea.NumberOfHMetrics)
	if err != nil {
		return
	}
	return
}
