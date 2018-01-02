package opentype

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// Fixed is a 32-bit signed fixed-point number (16.16)
type Fixed int32

// LongDateTime is a Date represented in number of seconds since 12:00 midnight, January 1, 1904. The value is represented as a signed 64-bit integer.
type LongDateTime int64

func (t LongDateTime) value() time.Time {
	return time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(t) * time.Second)
}

// Tag is a array of four uint8s (length = 32 bits) used to identify a script, language system, feature, or baseline
type Tag uint32

func (t Tag) String() string {
	bytes := []byte{
		byte(t >> 24 & 0xFF),
		byte(t >> 16 & 0xFF),
		byte(t >> 8 & 0xFF),
		byte(t & 0xFF),
	}
	return string(bytes)
}

// String2Tag creates Tag from string.
func String2Tag(s string) Tag {
	bytes := []byte(s)
	return Tag(uint32(bytes[0])<<24 + uint32(bytes[1])<<16 + uint32(bytes[2])<<8 + uint32(bytes[3]))
}

const (
	// SfntVersionTrueTypeOpenType : OpenType fonts that contain TrueType outlines.
	SfntVersionTrueTypeOpenType = Tag(0x00010000)
	// SfntVersionCFFOpenType  : OpenType fonts containing CFF data.
	SfntVersionCFFOpenType = Tag(0x4F54544F) // OTTO
	// SfntVersionAppleTrueType : The Apple specification for TrueType font.
	SfntVersionAppleTrueType = Tag(0x74727565) // true
	// SfntVersionAppleType1 : The Apple specification for the old style of PostScript font.
	SfntVersionAppleType1 = Tag(0x74797031) // typ1
	// SfntVersionTTCHeader : The header of TTC format file.
	SfntVersionTTCHeader = Tag(0x74746366) // ttcf
)

func parseSfntVersion(f *os.File, offset int64) (t Tag, err error) {
	f.Seek(offset, 0)
	err = binary.Read(f, binary.BigEndian, &t)
	f.Seek(offset, 0)
	return
}

// OffsetTable is the first table of OpenType font file.
type OffsetTable struct {
	SfntVersion Tag
	// Number of tables.
	NumTables uint16
	// (Maximum power of 2 <= numTables) x 16.
	SearchRange uint16
	// Log2(maximum power of 2 <= numTables).
	EntrySelector uint16
	// NumTables x 16-searchRange.
	RangeShift uint16
}

func parseOffsetTable(f *os.File) (ot *OffsetTable, err error) {
	ot = &OffsetTable{}
	err = binary.Read(f, binary.BigEndian, ot)
	return
}

func createOffsetTable(sfntVersion Tag, numTables uint16) *OffsetTable {
	ot := &OffsetTable{
		SfntVersion: sfntVersion,
		NumTables:   numTables,
	}
	ot.refreshField()
	return ot
}

const (
	// OffsetTableLength : size of OffsetTable
	OffsetTableLength = uint32(12)
	// TableRecordLength : size of TableRecord
	TableRecordLength = uint32(16)
)

func (ot *OffsetTable) refreshField() {
	es := uint16(0)
	sr := uint16(1)
	for {
		if 1<<(es+1) > ot.NumTables {
			break
		}
		es = es + 1
		sr = sr * 2
	}
	ot.SearchRange = sr * 16
	ot.EntrySelector = es
	ot.RangeShift = ot.NumTables*16 - ot.SearchRange
}

// TableRecord is a OpenType table record.
type TableRecord struct {
	Tag      Tag
	CheckSum uint32
	// Offset from beginning of TrueType font file.
	Offset uint32
	Length uint32
}

func parseTableRecord(f *os.File, numTables uint16) (trs map[string]*TableRecord, err error) {
	trs = make(map[string]*TableRecord)
	for i := uint16(0); i < numTables; i++ {
		tr := &TableRecord{}
		err = binary.Read(f, binary.BigEndian, tr)
		if err != nil {
			return
		}
		trs[tr.Tag.String()] = tr
	}
	for _, tr := range trs {
		err = tr.validate(f)
		if err != nil {
			return
		}
	}
	return
}

func createTableRecord(t Table, offset uint32) (tr *TableRecord, err error) {
	checkSum, err := t.CheckSum()
	if err != nil {
		return
	}
	tr = &TableRecord{
		Tag:      t.Tag(),
		CheckSum: checkSum,
		Offset:   offset,
		Length:   t.Length(),
	}
	return
}

func (tr *TableRecord) validate(f *os.File) (err error) {
	if "head" == tr.Tag.String() {
		return
	}
	_, err = f.Seek((int64)(tr.Offset), 0)
	if err != nil {
		return
	}
	checkSum, err := calcCheckSum(f, tr.Length)
	if checkSum != tr.CheckSum {
		err = fmt.Errorf("Table %s has invalid checksum, expected:%d actual:%d", tr.Tag, tr.CheckSum, checkSum)
	}
	return
}

// Table is a OpenType table.
type Table interface {
	Tag() Tag
	Length() uint32
	CheckSum() (uint32, error)
	store(w *errWriter)
}

func simpleCheckSum(t Table) (checkSum uint32, err error) {
	b := bytes.NewBuffer([]byte{})
	w := newErrWriter(b)
	t.store(w)
	if w.hasErr() {
		err = w.err
		return
	}
	return calcCheckSum(b, t.Length())
}

func calcCheckSum(r io.Reader, length uint32) (checkSum uint32, err error) {
	blocks := make([]uint32, padBlocks(length))
	err = binary.Read(r, binary.BigEndian, blocks)
	if err != nil {
		return
	}
	for _, block := range blocks {
		checkSum += block
	}
	return
}

func padBlocks(length uint32) uint32 {
	q := length / 4
	r := length % 4
	if 0 == r {
		return q
	}
	return q + 1
}

func padLength(length uint32) uint32 {
	return padBlocks(length) * 4
}

// All tables must begin on four-byte boundries, and any remaining space between tables is padded with zeros.
func padSpace(w *errWriter, length uint32) {
	pad := padLength(length) - length
	w.write(make([]uint8, pad))
}
