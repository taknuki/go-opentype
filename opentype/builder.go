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

// AddTable to Builder. If table is nil, that is ignored and returns false.
func (b *Builder) AddTable(t Table) bool {
	if t != nil {
		b.tables = append(b.tables, t)
		return true
	}
	return false
}

// AddTables to Builder. If table is nil, that is ignored.
func (b *Builder) AddTables(tables []Table) {
	for _, t := range tables {
		b.AddTable(t)
	}
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
	return fmt.Errorf("failed to create font file: %s", w.err)
}
