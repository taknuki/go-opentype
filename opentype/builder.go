package opentype

import (
	"fmt"
	"io"
)

// Builder is a font file builder.
type Builder struct {
	writer       *writer
	OffsetTable  *OffsetTable
	tableRecords map[string]*TableRecord
	Head         *Head
	Name         *Name
}

// NewBuilder creates Builder using parsed font.
func NewBuilder(font *Font) *Builder {
	b := &Builder{
		tableRecords: make(map[string]*TableRecord, len(font.tableRecords)),
		Head:         font.Head,
		Name:         font.Name,
	}
	for key, value := range font.tableRecords {
		b.tableRecords[key] = value
	}
	return b
}

// Build creates new font file.
func (b *Builder) Build(writer io.Writer) (err error) {
	for key := range b.tableRecords {
		if "head" != key && "name" != key {
			delete(b.tableRecords, key)
		}
	}
	b.writer = newWriter(writer, b.numTables())
	b.OffsetTable = createOffsetTable(SfntVersionTrueTypeOpenType, b.numTables())
	offset := OffsetTableLength + TableRecordLength*(uint32)(b.numTables())
	b.Head.CheckSumAdjustment = 0
	offset, err = b.replaceTableRecord(b.Head, offset)
	fmt.Println(offset)
	offset, err = b.replaceTableRecord(b.Name, offset)
	fmt.Println(offset)
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

type writer struct {
	writer       io.Writer
	tableRecords []*TableRecord
	tables       []Table
}

func newWriter(w io.Writer, numTables uint16) *writer {
	return &writer{
		writer:       w,
		tableRecords: make([]*TableRecord, 0, numTables),
		tables:       make([]Table, 0, numTables),
	}
}

func (w *writer) append(tr *TableRecord, t Table) {
	w.tableRecords = append(w.tableRecords, tr)
	w.tables = append(w.tables, t)
}

func (w *writer) write(b *Builder) (err error) {
	err = w.bw(b.OffsetTable)
	if err != nil {
		return
	}
	for _, tr := range w.tableRecords {
		err = w.bw(tr)
		if err != nil {
			return
		}
	}
	for _, t := range w.tables {
		err = t.Store(w.writer)
		if err != nil {
			return
		}
	}
	return
}

func (w *writer) bw(data interface{}) error {
	return bWrite(w.writer, data)
}
