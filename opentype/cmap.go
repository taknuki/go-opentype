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

// PlatformID is used to specify a particular character encoding.
type PlatformID uint16

// Name returns the name of Platform ID.
func (p PlatformID) Name() string {
	switch p {
	case PlatformIDUnicode:
		return "Unicode"
	case PlatformIDMacintosh:
		return "Macintosh"
	case PlatformIDISO:
		return "ISO"
	case PlatformIDWindows:
		return "Windows"
	case PlatformIDCustom:
		return "Custom"
	default:
		return unknownPlatformID
	}
}

func (p PlatformID) String() string {
	return fmt.Sprintf("(%d):%s", p, p.Name())
}

const (
	// PlatformIDUnicode : Unicode
	PlatformIDUnicode = PlatformID(0)
	// PlatformIDMacintosh : Macintosh
	PlatformIDMacintosh = PlatformID(1)
	// PlatformIDISO : ISO(deperecated)
	PlatformIDISO = PlatformID(2)
	// PlatformIDWindows : Windows
	PlatformIDWindows = PlatformID(3)
	// PlatformIDCustom : Custom
	PlatformIDCustom = PlatformID(4)
	// unknown
	unknownPlatformID = "Unknown"
)

// EncodingID is used to specify a particular character encoding.
type EncodingID uint16

// Name returns the name of Encoding ID.
func (e EncodingID) Name(p PlatformID) string {
	switch p {
	case PlatformIDUnicode:
		switch e {
		case EncodingIDUnicode10:
			return "Unicode 1.0 semantics"
		case EncodingIDUnicode11:
			return "Unicode 1.1 semantics"
		case EncodingIDUnicodeUCS:
			return "ISO/IEC 10646 semantics"
		case EncodingIDUnicode2BMP:
			return "Unicode 2.0 and onwards semantics, Unicode BMP only (cmap subtable formats 0, 4, 6)"
		case EncodingIDUnicode2Full:
			return "Unicode 2.0 and onwards semantics, Unicode full repertoire (cmap subtable formats 0, 4, 6, 10, 12)"
		case EncodingIDUnicodeVariation:
			return "Unicode Variation Sequences (cmap subtable format 14"
		case EncodingIDUnicodeFull:
			return "Unicode full repertoire (cmap subtable formats 0, 4, 6, 10, 12, 13)"
		}
	case PlatformIDWindows:
		switch e {
		case EncodingIDWindowsSymbol:
			return "Windows Symbol"
		case EncodingIDWindowsUnicodeBMP:
			return "Windows Unicode BMP"
		case EncodingIDWindowsSJIS:
			return "Windows ShiftJIS"
		case EncodingIDWindowsPRC:
			return "Windows PRC"
		case EncodingIDWindowsBig5:
			return "Windows Big5"
		case EncodingIDWindowsWansung:
			return "Windows Wansung"
		case EncodingIDWindowsJohab:
			return "Windows Johab"
		case EncodingIDWindowsUnicodeUCS4:
			return "Windows Unicode UCS-4"
		}
	case PlatformIDMacintosh:
		switch e {
		case EncodingIDMacintoshRoman:
			return "Macintosh Roman"
		case EncodingIDMacintoshJapanese:
			return "Macintosh Japanese"
		case EncodingIDMacintoshChineseTraditional:
			return "Macintosh Chinese (Traditional)"
		case EncodingIDMacintoshKorean:
			return "Macintosh Korean"
		case EncodingIDMacintoshArabic:
			return "Macintosh Arabic"
		case EncodingIDMacintoshHebrew:
			return "Macintosh Hebrew"
		case EncodingIDMacintoshGreek:
			return "Macintosh Greek"
		case EncodingIDMacintoshRussian:
			return "Macintosh Russian"
		case EncodingIDMacintoshRSymbol:
			return "Macintosh RSymbol"
		case EncodingIDMacintoshDevanagari:
			return "Macintosh Devanagari"
		case EncodingIDMacintoshGurmukhi:
			return "Macintosh Gurmukhi"
		case EncodingIDMacintoshGujarati:
			return "Macintosh Gujarati"
		case EncodingIDMacintoshOriya:
			return "Macintosh Oriya"
		case EncodingIDMacintoshBengali:
			return "Macintosh Bengali"
		case EncodingIDMacintoshTamil:
			return "Macintosh Tamil"
		case EncodingIDMacintoshTelugu:
			return "Macintosh Telugu"
		case EncodingIDMacintoshKannada:
			return "Macintosh Kannada"
		case EncodingIDMacintoshMalayalam:
			return "Macintosh Malayalam"
		case EncodingIDMacintoshSinhalese:
			return "Macintosh Sinhalese"
		case EncodingIDMacintoshBurmese:
			return "Macintosh Burmese"
		case EncodingIDMacintoshKhmer:
			return "Macintosh Khmer"
		case EncodingIDMacintoshThai:
			return "Macintosh Thai"
		case EncodingIDMacintoshLaotian:
			return "Macintosh Laotian"
		case EncodingIDMacintoshGeorgian:
			return "Macintosh Georgian"
		case EncodingIDMacintoshArmenian:
			return "Macintosh Armenian"
		case EncodingIDMacintoshChineseSimplified:
			return "Macintosh Chinese (Simplified)"
		case EncodingIDMacintoshTibetan:
			return "Macintosh Tibetan"
		case EncodingIDMacintoshMongolian:
			return "Macintosh Mongolian"
		case EncodingIDMacintoshGeez:
			return "Macintosh Geez"
		case EncodingIDMacintoshSlavic:
			return "Macintosh Slavic"
		case EncodingIDMacintoshVietnamese:
			return "Macintosh Vietnamese"
		case EncodingIDMacintoshSindhi:
			return "Macintosh Sindhi"
		case EncodingIDMacintoshUninterpreted:
			return "Macintosh Uninterpreted"
		}
	}
	return unknownEncodingID
}

func (e EncodingID) String(p PlatformID) string {
	return fmt.Sprintf("(%d):%s", e, e.Name(p))
}

const (
	// EncodingIDUnicode10 : Unicode 1.0 semantics
	EncodingIDUnicode10 = EncodingID(0)
	// EncodingIDUnicode11 : Unicode 1.1 semantics
	EncodingIDUnicode11 = EncodingID(1)
	// EncodingIDUnicodeUCS : ISO/IEC 10646 semantics
	EncodingIDUnicodeUCS = EncodingID(2)
	// EncodingIDUnicode2BMP : Unicode 2.0 and onwards semantics, Unicode BMP only (cmap subtable formats 0, 4, 6).
	EncodingIDUnicode2BMP = EncodingID(3)
	// EncodingIDUnicode2Full : Unicode 2.0 and onwards semantics, Unicode full repertoire (cmap subtable formats 0, 4, 6, 10, 12).
	EncodingIDUnicode2Full = EncodingID(4)
	// EncodingIDUnicodeVariation : Unicode Variation Sequences (cmap subtable format 14).
	EncodingIDUnicodeVariation = EncodingID(5)
	// EncodingIDUnicodeFull : Unicode full repertoire (cmap subtable formats 0, 4, 6, 10, 12, 13).
	EncodingIDUnicodeFull = EncodingID(6)
	// EncodingIDWindowsSymbol : Windows Symbol
	EncodingIDWindowsSymbol = EncodingID(0)
	// EncodingIDWindowsUnicodeBMP : Windows Unicode BMP (UCS-2)
	EncodingIDWindowsUnicodeBMP = EncodingID(1)
	// EncodingIDWindowsSJIS : Windows ShiftJIS
	EncodingIDWindowsSJIS = EncodingID(2)
	// EncodingIDWindowsPRC : Windows PRC
	EncodingIDWindowsPRC = EncodingID(3)
	// EncodingIDWindowsBig5 : Windows Big5
	EncodingIDWindowsBig5 = EncodingID(4)
	// EncodingIDWindowsWansung : Windows Wansung
	EncodingIDWindowsWansung = EncodingID(5)
	// EncodingIDWindowsJohab : Windows Johab
	EncodingIDWindowsJohab = EncodingID(6)
	// EncodingIDWindowsUnicodeUCS4 : Windows Unicode UCS-4
	EncodingIDWindowsUnicodeUCS4 = EncodingID(10)
	// EncodingIDMacintoshRoman : Macintosh Roman
	EncodingIDMacintoshRoman = EncodingID(0)
	// EncodingIDMacintoshJapanese : Macintosh Japanese
	EncodingIDMacintoshJapanese = EncodingID(1)
	// EncodingIDMacintoshChineseTraditional : Macintosh Chinese (Traditional)
	EncodingIDMacintoshChineseTraditional = EncodingID(2)
	// EncodingIDMacintoshKorean : Macintosh Korean
	EncodingIDMacintoshKorean = EncodingID(3)
	// EncodingIDMacintoshArabic : Macintosh Arabic
	EncodingIDMacintoshArabic = EncodingID(4)
	// EncodingIDMacintoshHebrew : Macintosh Hebrew
	EncodingIDMacintoshHebrew = EncodingID(5)
	// EncodingIDMacintoshGreek : Macintosh Greek
	EncodingIDMacintoshGreek = EncodingID(6)
	// EncodingIDMacintoshRussian : Macintosh Russian
	EncodingIDMacintoshRussian = EncodingID(7)
	// EncodingIDMacintoshRSymbol : Macintosh RSymbol
	EncodingIDMacintoshRSymbol = EncodingID(8)
	// EncodingIDMacintoshDevanagari : Macintosh Devanagari
	EncodingIDMacintoshDevanagari = EncodingID(9)
	// EncodingIDMacintoshGurmukhi : Macintosh Gurmukhi
	EncodingIDMacintoshGurmukhi = EncodingID(10)
	// EncodingIDMacintoshGujarati : Macintosh Gujarati
	EncodingIDMacintoshGujarati = EncodingID(11)
	// EncodingIDMacintoshOriya : Macintosh Oriya
	EncodingIDMacintoshOriya = EncodingID(12)
	// EncodingIDMacintoshBengali : Macintosh Bengali
	EncodingIDMacintoshBengali = EncodingID(13)
	// EncodingIDMacintoshTamil : Macintosh Tamil
	EncodingIDMacintoshTamil = EncodingID(14)
	// EncodingIDMacintoshTelugu : Macintosh Telugu
	EncodingIDMacintoshTelugu = EncodingID(15)
	// EncodingIDMacintoshKannada : Macintosh Kannada
	EncodingIDMacintoshKannada = EncodingID(16)
	// EncodingIDMacintoshMalayalam : Macintosh Malayalam
	EncodingIDMacintoshMalayalam = EncodingID(17)
	// EncodingIDMacintoshSinhalese : Macintosh Sinhalese
	EncodingIDMacintoshSinhalese = EncodingID(18)
	// EncodingIDMacintoshBurmese : Macintosh Burmese
	EncodingIDMacintoshBurmese = EncodingID(19)
	// EncodingIDMacintoshKhmer : Macintosh Khmer
	EncodingIDMacintoshKhmer = EncodingID(20)
	// EncodingIDMacintoshThai : Macintosh Thai
	EncodingIDMacintoshThai = EncodingID(21)
	// EncodingIDMacintoshLaotian : Macintosh Laotian
	EncodingIDMacintoshLaotian = EncodingID(22)
	// EncodingIDMacintoshGeorgian : Macintosh Georgian
	EncodingIDMacintoshGeorgian = EncodingID(23)
	// EncodingIDMacintoshArmenian : Macintosh Armenian
	EncodingIDMacintoshArmenian = EncodingID(24)
	// EncodingIDMacintoshChineseSimplified : Macintosh Chinese (Simplified)
	EncodingIDMacintoshChineseSimplified = EncodingID(25)
	// EncodingIDMacintoshTibetan : Macintosh Tibetan
	EncodingIDMacintoshTibetan = EncodingID(26)
	// EncodingIDMacintoshMongolian : Macintosh Mongolian
	EncodingIDMacintoshMongolian = EncodingID(27)
	// EncodingIDMacintoshGeez : Macintosh Geez
	EncodingIDMacintoshGeez = EncodingID(28)
	// EncodingIDMacintoshSlavic : Macintosh Slavic
	EncodingIDMacintoshSlavic = EncodingID(29)
	// EncodingIDMacintoshVietnamese : Macintosh Vietnamese
	EncodingIDMacintoshVietnamese = EncodingID(30)
	// EncodingIDMacintoshSindhi : Macintosh Sindhi
	EncodingIDMacintoshSindhi = EncodingID(31)
	// EncodingIDMacintoshUninterpreted : Macintosh Uninterpreted
	EncodingIDMacintoshUninterpreted = EncodingID(32)
	// unknown
	unknownEncodingID = "Unknown"
)

// GetCMap returns the resolved cmap of the encoding record subtable.
func (er *EncodingRecord) GetCMap() map[int32]uint16 {
	return er.Subtable.GetCMap()
}

// EncodingRecordSubtable is a character-to-glyph-index mapping table.
type EncodingRecordSubtable interface {
	GetFormatNumber() EncodingRecordSubtableFormatNumber
	GetCMap() map[int32]uint16
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

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat0) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber0
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat0) GetCMap() map[int32]uint16 {
	return st.cmap
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

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat2) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber2
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat2) GetCMap() map[int32]uint16 {
	return st.cmap
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

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat4) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber4
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat4) GetCMap() map[int32]uint16 {
	return st.cmap
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

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat6) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber6
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat6) GetCMap() map[int32]uint16 {
	return st.cmap
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

// GetFormatNumber returns the the format number of the encoding record subtable.
func (st *EncodingRecordSubtableFormat12) GetFormatNumber() EncodingRecordSubtableFormatNumber {
	return EncodingRecordSubtableFormatNumber12
}

// GetCMap returns the resolved cmap of the encoding record subtable.
func (st *EncodingRecordSubtableFormat12) GetCMap() map[int32]uint16 {
	return st.cmap
}
