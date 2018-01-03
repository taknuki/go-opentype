package opentype

import (
	"fmt"
	"os"
)

// Font is the opentype font.
type Font struct {
	offsetTable  *OffsetTable
	tableRecords map[string]*TableRecord
	Name         *Name
	CMap         *CMap
	Head         *Head
	Hhea         *Hhea
	Maxp         *Maxp
	Hmtx         *Hmtx
	Cvt          *Cvt
	Fpgm         *Fpgm
	Prep         *Prep
	Loca         *Loca
	Glyf         *Glyf
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
	if err != nil {
		return
	}
	p := newOptionalFontParser(font.tableRecords)
	p.parse("cvt ", func(tr *TableRecord) error {
		font.Cvt, err = parseCvt(f, tr.Offset, tr.Length)
		return err
	})
	p.parse("fpgm", func(tr *TableRecord) error {
		font.Fpgm, err = parseFpgm(f, tr.Offset, tr.Length)
		return err
	})
	p.parse("prep", func(tr *TableRecord) error {
		font.Prep, err = parsePrep(f, tr.Offset, tr.Length)
		return err
	})
	p.parse("loca", func(tr *TableRecord) error {
		err = tableRequired(font.Maxp, font.Head)
		if err != nil {
			return err
		}
		font.Loca, err = parseLoca(f, tr.Offset, font.Maxp.NumGlyphs, font.Head.IndexToLocFormat)
		return err
	})
	p.parse("glyf", func(tr *TableRecord) error {
		err = tableRequired(font.Loca)
		if err != nil {
			return err
		}
		font.Glyf, err = parseGlyf(f, tr.Offset, tr.Length, font.Loca)
		return err
	})
	err = p.err()
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
	p := newOptionalFontParser(font.tableRecords)
	p.parse("name", func(tr *TableRecord) error {
		font.Name, err = parseName(f, tr.Offset)
		return err
	})
	p.parse("head", func(tr *TableRecord) error {
		font.Head, err = parseHead(f, tr.Offset, tr.CheckSum)
		return err
	})
	p.parse("hhea", func(tr *TableRecord) error {
		font.Hhea, err = parseHhea(f, tr.Offset)
		return err
	})
	p.parse("maxp", func(tr *TableRecord) error {
		font.Maxp, err = parseMaxp(f, tr.Offset)
		return err
	})
	p.parse("hmtx", func(tr *TableRecord) error {
		// missed := make([]string, 0)
		// if font.Maxp == nil {
		// 	missed = append(missed, "maxp")
		// }
		// if font.Hhea == nil {
		// 	missed = append(missed, "hhea")
		// }
		// if len(missed) > 0 {
		// 	return fmt.Errorf("requires %s", strings.Join(missed, ","))
		// }
		err = tableRequired(font.Maxp, font.Hhea)
		if err != nil {
			return err
		}
		font.Hmtx, err = parseHmtx(f, tr.Offset, font.Maxp.NumGlyphs, font.Hhea.NumberOfHMetrics)
		return err
	})
	p.parse("cmap", func(tr *TableRecord) error {
		font.CMap, err = parseCMap(f, tr.Offset)
		return err
	})
	err = p.err()
	return
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
