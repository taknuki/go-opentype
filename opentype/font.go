package opentype

import (
	"fmt"
	"os"
)

// Font is the opentype font.
type Font struct {
	Name *Name
	CMap *CMap
	Head *Head
	Hhea *Hhea
	Maxp *Maxp
	Hmtx *Hmtx
	Cvt  *Cvt
	Fpgm *Fpgm
	Prep *Prep
	Loca *Loca
	Glyf *Glyf
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

func parseTrueTypeFont(f *os.File) (*Font, error) {
	return parseOpenTypeTable(f, func(font *Font, p *optionalFontParser) {
		p.parse("cvt ", true, func(tr *TableRecord) (err error) {
			font.Cvt, err = parseCvt(f, tr.Offset, tr.Length)
			return err
		})
		p.parse("fpgm", true, func(tr *TableRecord) (err error) {
			font.Fpgm, err = parseFpgm(f, tr.Offset, tr.Length)
			return err
		})
		p.parse("prep", true, func(tr *TableRecord) (err error) {
			font.Prep, err = parsePrep(f, tr.Offset, tr.Length)
			return err
		})
		p.parse("loca", false, func(tr *TableRecord) (err error) {
			err = tableRequired(font.Maxp, font.Head)
			if err != nil {
				return err
			}
			font.Loca, err = parseLoca(f, tr.Offset, font.Maxp.NumGlyphs, font.Head.IndexToLocFormat)
			return err
		})
		p.parse("glyf", false, func(tr *TableRecord) (err error) {
			err = tableRequired(font.Loca)
			if err != nil {
				return err
			}
			font.Glyf, err = parseGlyf(f, tr.Offset, tr.Length, font.Loca)
			return err
		})
	})
}

func parseCFFFont(f *os.File) (*Font, error) {
	return parseOpenTypeTable(f, func(font *Font, p *optionalFontParser) {
		// TODO CFF specific
	})
}

func parseOpenTypeTable(f *os.File, parser func(font *Font, p *optionalFontParser)) (font *Font, err error) {
	font = &Font{}
	offsetTable, err := parseOffsetTable(f)
	if err != nil {
		return
	}
	tableRecords, err := parseTableRecord(f, offsetTable.NumTables)
	if err != nil {
		return
	}
	p := newOptionalFontParser(tableRecords)
	p.parse("name", false, func(tr *TableRecord) error {
		font.Name, err = parseName(f, tr.Offset)
		return err
	})
	p.parse("head", false, func(tr *TableRecord) error {
		font.Head, err = parseHead(f, tr.Offset, tr.CheckSum)
		return err
	})
	p.parse("hhea", false, func(tr *TableRecord) error {
		font.Hhea, err = parseHhea(f, tr.Offset)
		return err
	})
	p.parse("maxp", false, func(tr *TableRecord) error {
		font.Maxp, err = parseMaxp(f, tr.Offset)
		return err
	})
	p.parse("hmtx", false, func(tr *TableRecord) error {
		err = tableRequired(font.Maxp, font.Hhea)
		if err != nil {
			return err
		}
		font.Hmtx, err = parseHmtx(f, tr.Offset, font.Maxp.NumGlyphs, font.Hhea.NumberOfHMetrics)
		return err
	})
	p.parse("cmap", true, func(tr *TableRecord) error {
		font.CMap, err = parseCMap(f, tr.Offset)
		return err
	})
	parser(font, p)
	return font, p.err()
}

// Tables are OpenType tables that are not nil.
func (font *Font) Tables() []Table {
	tables := []Table{
		font.Head,
		font.Name,
		font.Hhea,
		font.Maxp,
		font.Hmtx,
		font.Cvt,
		font.Fpgm,
		font.Prep,
		font.Loca,
		font.Glyf,
	}
	ret := make([]Table, 0, len(tables))
	for _, t := range tables {
		if t != nil {
			ret = append(ret, t)
		}
	}
	return ret
}

// FilterGlyf creates new Font with filtered glyf.
// You should set filter[0] = 0, that points to the “missing character”, or this method inserts it.
func (font *Font) FilterGlyf(filter []uint16) (*Font, error) {
	new := &Font{
		Name: font.Name,
		CMap: font.CMap,
		Head: font.Head,
		Hhea: font.Hhea,
		Maxp: font.Maxp,
		Hmtx: font.Hmtx,
		Cvt:  font.Cvt,
		Fpgm: font.Fpgm,
		Prep: font.Prep,
		Loca: font.Loca,
		Glyf: font.Glyf,
	}
	f := make([]uint16, 0, len(filter)+1)
	maxGID := uint16(0)
	if 0 != filter[0] {
		f = append(f, filter[0])
	}
	for _, gid := range filter {
		f = append(f, gid)
		if maxGID < gid {
			maxGID = gid
		}
	}
	if maxGID > font.Maxp.NumGlyphs-1 {
		return nil, fmt.Errorf("filtering glyph failed: request(%d) exceeds maximum glyph id(%d)", maxGID, font.Maxp.NumGlyphs-1)
	}
	new.Glyf = new.Glyf.filter(f)
	new.Loca = new.Glyf.generateLoca()
	new.Hmtx = new.Hmtx.filter(f)
	new.Maxp.NumGlyphs = uint16(len(f))
	new.Hhea.NumberOfHMetrics = uint16(len(new.Hmtx.HMetrics))
	new.Head.IndexToLocFormat = new.Loca.indexToLocFormat
	return new, nil
}
