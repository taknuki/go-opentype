package opentype

import (
	"fmt"
	"io"
)

// Builder is a font file builder.
type Builder struct {
	tables []Table
}

// NewBuilder creates Builder.
func NewBuilder() *Builder {
	return &Builder{
		tables: make([]Table, 0),
	}
}

// WithTable set table on target of Builder.
// If table is nil, Builder ignores it or removes the own table that has the same tag of it.
func (b *Builder) WithTable(t Table) *Builder {
	tables := make([]Table, 0, len(b.tables))
	for _, cur := range b.tables {
		if cur.Tag() != t.Tag() {
			tables = append(tables, cur)
		}
	}
	if t != nil {
		tables = append(tables, t)
	}
	b.tables = tables
	return b
}

// WithTables set tables on target of Builder.
func (b *Builder) WithTables(tables []Table) *Builder {
	for _, t := range tables {
		b.WithTable(t)
	}
	return b
}

// Build creates new font file.
func (b *Builder) Build(writer io.Writer) (err error) {
	numTables := len(b.tables)
	offsetTable := createOffsetTable(SfntVersionTrueTypeOpenType, uint16(numTables))
	offset := offsetTable.Length() + TableRecordLength*uint32(numTables)
	tableRecords := make(map[string]*TableRecord, numTables)
	for _, t := range b.tables {
		tr, err := createTableRecord(t, offset)
		if err != nil {
			return fmt.Errorf("failed to create TableRecord: %s cause: %s", t.Tag(), err)
		}
		tableRecords[t.Tag().String()] = tr
		offset += padLength(tr.Length)
	}
	w := newErrWriter(writer)
	w.write(offsetTable)
	for _, t := range b.tables {
		w.write(tableRecords[t.Tag().String()])
	}
	for _, t := range b.tables {
		t.store(w)
	}
	return w.errorf("failed to create font file: %s")
}
