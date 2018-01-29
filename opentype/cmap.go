package opentype

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// CMap is a "cmap" table.
type CMap struct {
	Header          *CMapHeader
	EncodingRecords []*EncodingRecord
}

func parseCMap(f *os.File, offset uint32) (cm *CMap, err error) {
	cm = &CMap{}
	cm.Header = &CMapHeader{}
	f.Seek(int64(offset), 0)
	err = binary.Read(f, binary.BigEndian, cm.Header)
	if err != nil {
		return
	}
	cm.EncodingRecords = make([]*EncodingRecord, int(cm.Header.NumTables))
	for i := 0; i < int(cm.Header.NumTables); i++ {
		er := &EncodingRecord{}
		err = binary.Read(f, binary.BigEndian, &(er.PlatformID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(er.EncodingID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(er.Offset))
		if err != nil {
			return
		}
		cm.EncodingRecords[i] = er
	}
	for _, er := range cm.EncodingRecords {
		f.Seek(int64(offset)+int64(er.Offset), 0)
		er.Subtable, err = parseEncodingRecordSubtable(f)
		if err != nil {
			return
		}
	}
	return
}

// Tag is table name.
func (cm *CMap) Tag() Tag {
	return String2Tag("cmap")
}

// store writes binary expression of this table.
func (cm *CMap) store(w *errWriter) {
	w.write(&(cm.Header))
	for _, er := range cm.EncodingRecords {
		er.store(w)
	}
	padSpace(w, cm.Length())
}

// CheckSum for this table.
func (cm *CMap) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(cm)
}

// Length returns the size(byte) of this table.
func (cm *CMap) Length() uint32 {
	l := uint32(4) // CMapHeader
	for _, er := range cm.EncodingRecords {
		l += er.Length()
	}
	return l
}

// Exists returns true if this is not nil.
func (cm *CMap) Exists() bool {
	return cm != nil
}

// CMapHeader is a header block of a "cmap" table.
type CMapHeader struct {
	Version   uint16
	NumTables uint16
}

// EncodingRecord has a platform id, a encoding id, and a subtable.
type EncodingRecord struct {
	PlatformID PlatformID
	EncodingID EncodingID
	Offset     uint32
	Subtable   EncodingRecordSubtable
}

// CMap is the resolved cmap of the encoding record subtable.
func (er *EncodingRecord) CMap() map[int32]uint16 {
	return er.Subtable.GetCMap()
}

// Length returns the size(byte) of this EncodingRecord.
func (er *EncodingRecord) Length() uint32 {
	return uint32(8) + er.Subtable.GetLength()
}

// store writes binary expression of this EncodingRecord.
func (er *EncodingRecord) store(w *errWriter) {
	w.write(&(er.PlatformID))
	w.write(&(er.EncodingID))
	w.write(&(er.Offset))
	er.Subtable.store(w)
}

// EncodingRecordSubtable is a character-to-glyph-index mapping table.
type EncodingRecordSubtable interface {
	GetFormatNumber() EncodingRecordSubtableFormatNumber
	GetCMap() map[int32]uint16
	GetLength() uint32
	store(w *errWriter)
}

// EncodingRecordSubtableFormatNumber is a format specifier of encoding record subtables.
type EncodingRecordSubtableFormatNumber uint16

func (fn EncodingRecordSubtableFormatNumber) String() string {
	return fmt.Sprintf("Format %d", fn)
}

const (
	// EncodingRecordSubtableFormatNumber0 : the Apple standard character to glyph index mapping table.
	EncodingRecordSubtableFormatNumber0 = EncodingRecordSubtableFormatNumber(0)
	// EncodingRecordSubtableFormatNumber2 : This subtable is useful for the national character code standards used for Japanese, Chinese, and Korean characters.
	EncodingRecordSubtableFormatNumber2 = EncodingRecordSubtableFormatNumber(2)
	// EncodingRecordSubtableFormatNumber4 : the Microsoft standard character-to-glyph-index mapping table for fonts that support Unicode BMP characters.
	EncodingRecordSubtableFormatNumber4 = EncodingRecordSubtableFormatNumber(4)
	// EncodingRecordSubtableFormatNumber6 : Trimmed table mapping.
	EncodingRecordSubtableFormatNumber6 = EncodingRecordSubtableFormatNumber(6)
	// EncodingRecordSubtableFormatNumber12 : the Microsoft standard character-to-glyph-index mapping table for fonts supporting Unicode supplementary-plane characters (U+10000 to U+10FFFF).
	EncodingRecordSubtableFormatNumber12 = EncodingRecordSubtableFormatNumber(12)
)

func parseEncodingRecordSubtable(f *os.File) (st EncodingRecordSubtable, err error) {
	var format EncodingRecordSubtableFormatNumber
	err = binary.Read(f, binary.BigEndian, &format)
	if err != nil {
		return
	}
	switch format {
	case EncodingRecordSubtableFormatNumber0:
		st, err = parseEncodingRecordSubtableFormat0(f)
	case EncodingRecordSubtableFormatNumber2:
		st, err = parseEncodingRecordSubtableFormat2(f)
	case EncodingRecordSubtableFormatNumber4:
		st, err = parseEncodingRecordSubtableFormat4(f)
	case EncodingRecordSubtableFormatNumber6:
		st, err = parseEncodingRecordSubtableFormat6(f)
	case EncodingRecordSubtableFormatNumber12:
		st, err = parseEncodingRecordSubtableFormat12(f)
	default:
		err = fmt.Errorf("encoding record subtable %s is not suppored", format)
	}
	return
}

// EncodingRecordSubtableFormat0 is the Apple standard character to glyph index mapping table.
type EncodingRecordSubtableFormat0 struct {
	Length       uint16
	Language     uint16
	GlyphIDArray [256]uint8
	cmap         map[int32]uint16
}

func parseEncodingRecordSubtableFormat0(f *os.File) (st *EncodingRecordSubtableFormat0, err error) {
	st = &EncodingRecordSubtableFormat0{}
	err = binary.Read(f, binary.BigEndian, &(st.Length))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.Language))
	if err != nil {
		return
	}
	st.cmap = make(map[int32]uint16)
	for i := 0; i < 256; i++ {
		err = binary.Read(f, binary.BigEndian, &(st.GlyphIDArray[i]))
		if err != nil {
			return
		}
		st.cmap[int32(i)] = uint16(st.GlyphIDArray[i])
	}
	return
}

func (st *EncodingRecordSubtableFormat0) store(w *errWriter) {
	writeEncodingRecordSubtableFormatNumber(w, st.GetFormatNumber())
	w.write(&(st.Length))
	w.write(&(st.Language))
	w.write(st.GlyphIDArray)
	return
}

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat0) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber0
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat0) GetCMap() map[int32]uint16 {
	return st.cmap
}

// GetLength returns the length of this subtable.
func (st *EncodingRecordSubtableFormat0) GetLength() uint32 {
	return uint32(st.Length)
}

// EncodingRecordSubtableFormat2 is useful for the national character code standards used for Japanese, Chinese, and Korean characters.
type EncodingRecordSubtableFormat2 struct {
	Length               uint16
	Language             uint16
	SubHeaderKeys        [256]uint16
	FirstCode            []uint16
	EntryCount           []uint16
	IDDelta              []int16
	IDRangeOffset        []uint16
	idRangeOffsetAddress []int64
	cmap                 map[int32]uint16
}

func parseEncodingRecordSubtableFormat2(f *os.File) (st *EncodingRecordSubtableFormat2, err error) {
	st = &EncodingRecordSubtableFormat2{}
	err = binary.Read(f, binary.BigEndian, &(st.Length))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.Language))
	if err != nil {
		return
	}
	subHeaderNum := uint8(1)
	for i := 0; i < 256; i++ {
		err = binary.Read(f, binary.BigEndian, &(st.SubHeaderKeys[i]))
		if err != nil {
			return
		}
		if st.SubHeaderKeys[i] > 0 {
			subHeaderNum++
		}
	}
	st.FirstCode = make([]uint16, subHeaderNum)
	st.EntryCount = make([]uint16, subHeaderNum)
	st.IDDelta = make([]int16, subHeaderNum)
	st.IDRangeOffset = make([]uint16, subHeaderNum)
	st.idRangeOffsetAddress = make([]int64, subHeaderNum)
	for j := uint8(0); j < subHeaderNum; j++ {
		err = binary.Read(f, binary.BigEndian, &(st.FirstCode[j]))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.EntryCount[j]))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.IDDelta[j]))
		if err != nil {
			return
		}
		st.idRangeOffsetAddress[j], err = f.Seek(0, 1)
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.IDRangeOffset[j]))
		if err != nil {
			return
		}
	}
	st.cmap, err = st.createCMap(f)
	return
}

func (st *EncodingRecordSubtableFormat2) createCMap(f *os.File) (cmap map[int32]uint16, err error) {
	cmap = make(map[int32]uint16)
	for i := uint16(0); i < 256; i++ {
		gid, e := st.getGID(f, st.idRangeOffsetAddress[0], st.IDRangeOffset[0]+2*i, st.IDDelta[0])
		if e != nil {
			return nil, e
		}
		if gid > 0 {
			cmap[int32(i)] = gid
		}
		if st.SubHeaderKeys[i] > 0 {
			key := st.SubHeaderKeys[i] / 8
			for j := uint16(0); j < st.EntryCount[key]; j++ {
				c := st.FirstCode[key] + j + i*256
				gid, e := st.getGID(f, st.idRangeOffsetAddress[key], st.IDRangeOffset[key]+2*j, st.IDDelta[key])
				if e != nil {
					return nil, e
				}
				if gid > 0 {
					cmap[int32(c)] = gid
				}
			}
		}
	}
	return
}

func (st *EncodingRecordSubtableFormat2) getGID(f *os.File, addr int64, offset uint16, delta int16) (gid uint16, err error) {
	var dummy uint16
	_, err = f.Seek(addr+int64(offset), 0)
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &dummy)
	if dummy == 0 {
		gid = 0
	} else {
		gid = uint16((int32(dummy) + int32(delta)) % 65536)
	}
	return
}

func (st *EncodingRecordSubtableFormat2) store(w *errWriter) {
	// w.writeEncodingRecordSubtableFormatNumber(st.GetFormatNumber())
	// w.write(&(st.Length))
	// w.write(&(st.Language))
	// w.write(st.SubHeaderKeys)
	// for i, fc := range
	// return
}

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat2) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber2
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat2) GetCMap() map[int32]uint16 {
	return st.cmap
}

// GetLength returns the length of this subtable.
func (st *EncodingRecordSubtableFormat2) GetLength() uint32 {
	return uint32(st.Length)
}

// EncodingRecordSubtableFormat4 is the Microsoft standard character-to-glyph-index mapping table for fonts that support Unicode BMP characters.
type EncodingRecordSubtableFormat4 struct {
	Length               uint16
	Language             uint16
	SegCount             uint16
	SearchRange          uint16
	EntrySelector        uint16
	RangeShift           uint16
	EndCount             []uint16
	ReservedPad          uint16
	StartCount           []uint16
	IDDelta              []int16
	IDRangeOffset        []uint16
	IDRangeOffsetAddress []int64
	cmap                 map[int32]uint16
}

func parseEncodingRecordSubtableFormat4(f *os.File) (st *EncodingRecordSubtableFormat4, err error) {
	st = &EncodingRecordSubtableFormat4{}
	err = binary.Read(f, binary.BigEndian, &(st.Length))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.Language))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.SegCount))
	if err != nil {
		return
	}
	st.SegCount /= 2
	err = binary.Read(f, binary.BigEndian, &(st.SearchRange))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.EntrySelector))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.RangeShift))
	if err != nil {
		return
	}
	st.EndCount = make([]uint16, st.SegCount)
	for i := 0; i < int(st.SegCount); i++ {
		err = binary.Read(f, binary.BigEndian, &(st.EndCount[i]))
		if err != nil {
			return
		}
	}
	err = binary.Read(f, binary.BigEndian, &(st.ReservedPad))
	if err != nil {
		return
	}
	st.StartCount = make([]uint16, st.SegCount)
	for i := 0; i < int(st.SegCount); i++ {
		err = binary.Read(f, binary.BigEndian, &(st.StartCount[i]))
		if err != nil {
			return
		}
	}
	st.IDDelta = make([]int16, st.SegCount)
	for i := 0; i < int(st.SegCount); i++ {
		err = binary.Read(f, binary.BigEndian, &(st.IDDelta[i]))
		if err != nil {
			return
		}
	}
	st.IDRangeOffset = make([]uint16, st.SegCount)
	st.IDRangeOffsetAddress = make([]int64, st.SegCount)
	for i := 0; i < int(st.SegCount); i++ {
		st.IDRangeOffsetAddress[i], err = f.Seek(0, 1)
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.IDRangeOffset[i]))
		if err != nil {
			return
		}
	}
	st.cmap, err = st.createCMap(f)
	return
}

func (st *EncodingRecordSubtableFormat4) createCMap(f *os.File) (cmap map[int32]uint16, err error) {
	cmap = make(map[int32]uint16)
	for i := uint16(0); i < st.SegCount; i++ {
		if st.EndCount[i] == math.MaxUint16 {
			break
		}
		for c := st.StartCount[i]; c <= st.EndCount[i]; c++ {
			cmap[int32(c)], err = st.getGID(f, i, c)
			if err != nil {
				return
			}
		}
	}
	return
}

func (st *EncodingRecordSubtableFormat4) getGID(f *os.File, i uint16, c uint16) (gid uint16, err error) {
	if st.IDRangeOffset[i] == 0 {
		gid = uint16(int32(c) + int32(st.IDDelta[i]))
		return
	}
	pos := int64(st.IDRangeOffset[i]) + 2*(int64(c)-int64(st.StartCount[i])) + st.IDRangeOffsetAddress[i]
	_, err = f.Seek(pos, 0)
	if err != nil {
		return
	}
	var C uint16
	err = binary.Read(f, binary.BigEndian, &C)
	if err != nil {
		return
	}
	if C == 0 {
		gid = 0
	} else {
		gid = uint16((int32(C) + int32(st.IDDelta[i])) % 65536)
	}
	return
}

func (st *EncodingRecordSubtableFormat4) store(w *errWriter) {
}

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat4) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber4
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat4) GetCMap() map[int32]uint16 {
	return st.cmap
}

// GetLength returns the length of this subtable.
func (st *EncodingRecordSubtableFormat4) GetLength() uint32 {
	return uint32(st.Length)
}

// EncodingRecordSubtableFormat6 is Trimmed table mapping
type EncodingRecordSubtableFormat6 struct {
	Length       uint16
	Language     uint16
	firstCode    uint16
	entryCount   uint16
	glyphIDArray []uint16
	cmap         map[int32]uint16
}

func parseEncodingRecordSubtableFormat6(f *os.File) (st *EncodingRecordSubtableFormat6, err error) {
	st = &EncodingRecordSubtableFormat6{}
	err = binary.Read(f, binary.BigEndian, &(st.Length))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.Language))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.firstCode))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.entryCount))
	if err != nil {
		return
	}
	st.glyphIDArray = make([]uint16, st.entryCount)
	st.cmap = make(map[int32]uint16)
	for i := uint16(0); i < st.entryCount; i++ {
		err = binary.Read(f, binary.BigEndian, &(st.glyphIDArray[i]))
		if err != nil {
			return
		}
		if st.glyphIDArray[i] > 0 {
			c := st.firstCode + i
			st.cmap[int32(c)] = uint16(st.glyphIDArray[i])
		}
	}
	return
}

func (st *EncodingRecordSubtableFormat6) store(w *errWriter) {
}

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat6) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber6
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat6) GetCMap() map[int32]uint16 {
	return st.cmap
}

// GetLength returns the length of this subtable.
func (st *EncodingRecordSubtableFormat6) GetLength() uint32 {
	return uint32(st.Length)
}

// EncodingRecordSubtableFormat12 is the Microsoft standard character-to-glyph-index mapping table for fonts supporting Unicode supplementary-plane characters (U+10000 to U+10FFFF).
type EncodingRecordSubtableFormat12 struct {
	Length        uint32
	Language      uint32
	NumGroups     uint32
	startCharCode []uint32
	endCharCode   []uint32
	startGlyphID  []uint32
	cmap          map[int32]uint16
}

func parseEncodingRecordSubtableFormat12(f *os.File) (st *EncodingRecordSubtableFormat12, err error) {
	st = &EncodingRecordSubtableFormat12{}
	f.Seek(2, 1)
	err = binary.Read(f, binary.BigEndian, &(st.Length))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.Language))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(st.NumGroups))
	if err != nil {
		return
	}
	st.startCharCode = make([]uint32, int(st.NumGroups))
	st.endCharCode = make([]uint32, int(st.NumGroups))
	st.startGlyphID = make([]uint32, int(st.NumGroups))
	for i := uint32(0); i < st.NumGroups; i++ {
		err = binary.Read(f, binary.BigEndian, &(st.startCharCode[i]))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.endCharCode[i]))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(st.startGlyphID[i]))
		if err != nil {
			return
		}
	}
	st.cmap = st.createCMap()
	return
}

func (st *EncodingRecordSubtableFormat12) createCMap() (cmap map[int32]uint16) {
	cmap = make(map[int32]uint16)
	for i := uint32(0); i < st.NumGroups; i++ {
		for j := st.startCharCode[i]; j < st.endCharCode[i]; j++ {
			d := j - st.startCharCode[i]
			gid := st.startGlyphID[i] + d
			cmap[int32(j)] = uint16(gid)
		}
	}
	return
}

func (st *EncodingRecordSubtableFormat12) store(w *errWriter) {
}

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat12) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber12
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat12) GetCMap() map[int32]uint16 {
	return st.cmap
}

// GetLength returns the length of this subtable.
func (st *EncodingRecordSubtableFormat12) GetLength() uint32 {
	return uint32(st.Length)
}

func writeEncodingRecordSubtableFormatNumber(e *errWriter, n EncodingRecordSubtableFormatNumber) {
	b := []byte{byte(n / 256), byte(n % 256)}
	e.writeBin(b)
}
