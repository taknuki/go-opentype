package opentype

import "io"

// Builder is a font file builder.
type Builder struct {
	writer       *fontWriter
	OffsetTable  *OffsetTable
	tableRecords map[string]*TableRecord
	Head         *Head
	Name         *Name
	Hhea         *Hhea
	Maxp         *Maxp
	Hmtx         *Hmtx
	Cvt          *Cvt
	Fpgm         *Fpgm
	Prep         *Prep
	Loca         *Loca
	Glyf         *Glyf
}

// NewBuilder creates Builder using parsed font.
func NewBuilder(font *Font) *Builder {
	b := &Builder{
		tableRecords: make(map[string]*TableRecord, len(font.tableRecords)),
		Head:         font.Head,
		Name:         font.Name,
		Hhea:         font.Hhea,
		Maxp:         font.Maxp,
		Hmtx:         font.Hmtx,
		Cvt:          font.Cvt,
		Fpgm:         font.Fpgm,
		Prep:         font.Prep,
		Loca:         font.Loca,
		Glyf:         font.Glyf,
	}
	for key, value := range font.tableRecords {
		b.tableRecords[key] = value
	}
	return b
}

// Filter deletes opentype tables that is not in the list from the target of Builder.
func (b *Builder) Filter(list []string) {
	new := make(map[string]*TableRecord, len(list))
	for _, tag := range list {
		tr, ok := b.tableRecords[tag]
		if ok {
			new[tag] = tr
		}
	}
	b.tableRecords = new
}

// Build creates new font file.
func (b *Builder) Build(writer io.Writer) (err error) {
	b.writer = newFontWriter(writer, b.numTables())
	b.OffsetTable = createOffsetTable(SfntVersionTrueTypeOpenType, b.numTables())
	offset := b.OffsetTable.Length() + TableRecordLength*(uint32)(b.numTables())
	b.Head.CheckSumAdjustment = 0
	offset, err = b.replaceTableRecord(b.Head, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Name, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Hhea, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Maxp, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Hmtx, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Cvt, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Fpgm, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Prep, offset)
	if err != nil {
		return
	}
	offset, err = b.replaceTableRecord(b.Loca, offset)
	if err != nil {
		return
	}
	_, err = b.replaceTableRecord(b.Glyf, offset)
	if err != nil {
		return
	}
	return b.writer.write(b)
}

func (b *Builder) replaceTableRecord(t Table, offset uint32) (nextOffset uint32, err error) {
	tr, err := createTableRecord(t, offset)
	if err != nil {
		return
	}
	b.tableRecords[t.Tag().String()] = tr
	nextOffset = offset + padLength(tr.Length)
	b.writer.append(tr, t)
	return
}

func (b *Builder) numTables() uint16 {
	return uint16(len(b.tableRecords))
}

type fontWriter struct {
	writer       *errWriter
	tableRecords []*TableRecord
	tables       []Table
}

func newFontWriter(w io.Writer, numTables uint16) *fontWriter {
	return &fontWriter{
		writer:       newErrWriter(w),
		tableRecords: make([]*TableRecord, 0, numTables),
		tables:       make([]Table, 0, numTables),
	}
}

func (w *fontWriter) append(tr *TableRecord, t Table) {
	w.tableRecords = append(w.tableRecords, tr)
	w.tables = append(w.tables, t)
}

func (w *fontWriter) write(b *Builder) (err error) {
	w.writer.write(b.OffsetTable)
	for _, tr := range w.tableRecords {
		w.writer.write(tr)
	}
	for _, t := range w.tables {
		t.store(w.writer)
	}
	return w.writer.err
}
