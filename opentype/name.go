package opentype

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"unicode/utf16"
)

// Name is a "name" table.
// The naming table allows multilingual strings to be associated with the OpenType™ font file.
type Name struct {
	Format uint16
	// Number of name records.
	Count uint16
	// Offset to start of string storage (from start of table).
	StringOffset uint16
	// 	The name records.
	NameRecords []*NameRecord
	// Number of language-tag records.
	LangTagCount uint16
	// The language-tag records.
	LangTagRecords []*LangTagRecord
}

func parseName(f *os.File, offset uint32) (n *Name, err error) {
	n = &Name{}
	_, err = f.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(n.Format))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(n.Count))
	if err != nil {
		return
	}
	err = binary.Read(f, binary.BigEndian, &(n.StringOffset))
	if err != nil {
		return
	}
	storageOffset := int64(offset) + int64(n.StringOffset)
	n.NameRecords, err = parseNameRecords(f, n.Count, storageOffset)
	// Format 1
	if 1 == n.Format {
		err = binary.Read(f, binary.BigEndian, &(n.LangTagCount))
		if err != nil {
			return
		}
		n.LangTagRecords = make([]*LangTagRecord, n.LangTagCount)
		for i := 0; i < int(n.LangTagCount); i++ {
			r := &LangTagRecord{}
			err = binary.Read(f, binary.BigEndian, r)
			if err != nil {
				return
			}
			n.LangTagRecords[i] = r
		}
	}
	return
}

// Tag is table name.
func (n *Name) Tag() Tag {
	return String2Tag("name")
}

// store writes binary expression of this table.
func (n *Name) store(w *errWriter) {
	w.write(&(n.Format))
	w.write(&(n.Count))
	w.write(&(n.StringOffset))
	for _, nr := range n.NameRecords {
		w.write(&(nr.PlatformID))
		w.write(&(nr.EncodingID))
		w.write(&(nr.LanguageID))
		w.write(&(nr.NameID))
		w.write(&(nr.Length))
		w.write(&(nr.Offset))
	}
	if 1 == n.Format {
		w.write(&(n.LangTagCount))
		for _, ltr := range n.LangTagRecords {
			w.write(&(ltr.Length))
			w.write(&(ltr.Offset))
		}
	}
	for _, nr := range n.NameRecords {
		if PlatformIDMacintosh == nr.PlatformID {
			w.writeBin([]byte(nr.Value))
		} else {
			for _, u := range utf16.Encode([]rune(nr.Value)) {
				w.write(&(u))
			}
		}
	}
	padSpace(w, n.Length())
}

// CheckSum for this table.
func (n *Name) CheckSum() (checkSum uint32, err error) {
	return simpleCheckSum(n)
}

// Length returns the size(byte) of this table.
func (n *Name) Length() uint32 {
	l := 6 + 12*n.Count
	s := uint32(0)
	for _, nr := range n.NameRecords {
		s += (uint32)(nr.Length)
	}
	if 1 == n.Format {
		l += 2 + 4*n.LangTagCount
	}
	return uint32(l) + uint32(s)
}

// NameRecord contains platform specific metadata of the OpenType Font.
type NameRecord struct {
	// Platform ID.
	PlatformID PlatformID
	// Platform-specific encoding ID.
	EncodingID EncodingID
	// Language ID.
	LanguageID LanguageID
	// Name ID.
	NameID NameID
	// String length (in bytes).
	Length uint16
	// String offset from start of storage area (in bytes).
	Offset uint16
	// string value of the NameRecord.
	Value string
}

func parseNameRecords(f *os.File, count uint16, storageOffset int64) (nrs []*NameRecord, err error) {
	nrs = make([]*NameRecord, count)
	for i := 0; i < int(count); i++ {
		nr := &NameRecord{}
		err = binary.Read(f, binary.BigEndian, &(nr.PlatformID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(nr.EncodingID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(nr.LanguageID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(nr.NameID))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(nr.Length))
		if err != nil {
			return
		}
		err = binary.Read(f, binary.BigEndian, &(nr.Offset))
		if err != nil {
			return
		}
		nrs[i] = nr
	}
	for _, nr := range nrs {
		_, err = f.Seek(storageOffset+int64(nr.Offset), 0)
		if err != nil {
			return
		}
		if PlatformIDMacintosh == nr.PlatformID {
			b := make([]byte, nr.Length)
			err = binary.Read(f, binary.BigEndian, b)
			nr.Value = string(b)
		} else {
			s := make([]uint16, nr.Length/2)
			err = binary.Read(f, binary.BigEndian, s)
			nr.Value = string(utf16.Decode(s))
		}
		if err != nil {
			return
		}
	}
	return
}

// LangTagRecord contains language name of langauge id within the range 0x8000 to 0x8000 + langTagCount - 1.
type LangTagRecord struct {
	// Language-tag string length (in bytes)
	Length uint16
	// Language-tag string offset from start of storage area (in bytes).
	Offset uint16
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

// LanguageID is used to specify a particular language.
type LanguageID uint16

// Name returns the name of Language ID.
func (l LanguageID) Name(p PlatformID) string {
	switch p {
	case PlatformIDUnicode:
		return languageIDUnicode
	case PlatformIDWindows:
		switch l {
		case LanguageIDWindowsAfrikaansSouthAfrica:
			return "Windows Afrikaans-South Africa"
		case LanguageIDWindowsAlbanianAlbania:
			return "Windows Albanian-Albania"
		case LanguageIDWindowsAlsatianFrance:
			return "Windows Alsatian-France"
		case LanguageIDWindowsAmharicEthiopia:
			return "Windows Amharic-Ethiopia"
		case LanguageIDWindowsArabicAlgeria:
			return "Windows Arabic-Algeria"
		case LanguageIDWindowsArabicBahrain:
			return "Windows Arabic-Bahrain"
		case LanguageIDWindowsArabicEgypt:
			return "Windows Arabic-Egypt"
		case LanguageIDWindowsArabicIraq:
			return "Windows Arabic-Iraq"
		case LanguageIDWindowsArabicJordan:
			return "Windows Arabic-Jordan"
		case LanguageIDWindowsArabicKuwait:
			return "Windows Arabic-Kuwait"
		case LanguageIDWindowsArabicLebanon:
			return "Windows Arabic-Lebanon"
		case LanguageIDWindowsArabicLibya:
			return "Windows Arabic-Libya"
		case LanguageIDWindowsArabicMorocco:
			return "Windows Arabic-Morocco"
		case LanguageIDWindowsArabicOman:
			return "Windows Arabic-Oman"
		case LanguageIDWindowsArabicQatar:
			return "Windows Arabic-Qatar"
		case LanguageIDWindowsArabicSaudiArabia:
			return "Windows Arabic-Saudi Arabia"
		case LanguageIDWindowsArabicSyria:
			return "Windows Arabic-Syria"
		case LanguageIDWindowsArabicTunisia:
			return "Windows Arabic-Tunisia"
		case LanguageIDWindowsArabicUAE:
			return "Windows Arabic-U.A.E."
		case LanguageIDWindowsArabicYemen:
			return "Windows Arabic-Yemen"
		case LanguageIDWindowsArmenianArmenia:
			return "Windows Armenian-Armenia"
		case LanguageIDWindowsAssameseIndia:
			return "Windows Assamese-India"
		case LanguageIDWindowsAzeriCyrillicAzerbaijan:
			return "Windows Azeri (Cyrillic)-Azerbaijan"
		case LanguageIDWindowsAzeriLatinAzerbaijan:
			return "Windows Azeri (Latin)-Azerbaijan"
		case LanguageIDWindowsBashkirRussia:
			return "Windows Bashkir-Russia"
		case LanguageIDWindowsBasqueBasque:
			return "Windows Basque-Basque"
		case LanguageIDWindowsBelarusianBelarus:
			return "Windows Belarusian-Belarus"
		case LanguageIDWindowsBengaliBangladesh:
			return "Windows Bengali-Bangladesh"
		case LanguageIDWindowsBengaliIndia:
			return "Windows Bengali-India"
		case LanguageIDWindowsBosnianCyrillicBosniaandHerzegovina:
			return "Windows Bosnian (Cyrillic)-Bosnia and Herzegovina"
		case LanguageIDWindowsBosnianLatinBosniaandHerzegovina:
			return "Windows Bosnian (Latin)-Bosnia and Herzegovina"
		case LanguageIDWindowsBretonFrance:
			return "Windows Breton-France"
		case LanguageIDWindowsBulgarianBulgaria:
			return "Windows Bulgarian-Bulgaria"
		case LanguageIDWindowsCatalanCatalan:
			return "Windows Catalan-Catalan"
		case LanguageIDWindowsChineseHongKongSAR:
			return "Windows Chinese-Hong Kong S.A.R."
		case LanguageIDWindowsChineseMacaoSAR:
			return "Windows Chinese-Macao S.A.R."
		case LanguageIDWindowsChinesePeoplesRepublicofChina:
			return "Windows Chinese-People's Republic of China"
		case LanguageIDWindowsChineseSingapore:
			return "Windows Chinese-Singapore"
		case LanguageIDWindowsChineseTaiwan:
			return "Windows Chinese-Taiwan"
		case LanguageIDWindowsCorsicanFrance:
			return "Windows Corsican-France"
		case LanguageIDWindowsCroatianCroatia:
			return "Windows Croatian-Croatia"
		case LanguageIDWindowsCroatianLatinBosniaandHerzegovina:
			return "Windows Croatian (Latin)-Bosnia and Herzegovina"
		case LanguageIDWindowsCzechCzechRepublic:
			return "Windows Czech-Czech Republic"
		case LanguageIDWindowsDanishDenmark:
			return "Windows Danish-Denmark"
		case LanguageIDWindowsDariAfghanistan:
			return "Windows Dari-Afghanistan"
		case LanguageIDWindowsDivehiMaldives:
			return "Windows Divehi-Maldives"
		case LanguageIDWindowsDutchBelgium:
			return "Windows Dutch-Belgium"
		case LanguageIDWindowsDutchNetherlands:
			return "Windows Dutch-Netherlands"
		case LanguageIDWindowsEnglishAustralia:
			return "Windows English-Australia"
		case LanguageIDWindowsEnglishBelize:
			return "Windows English-Belize"
		case LanguageIDWindowsEnglishCanada:
			return "Windows English-Canada"
		case LanguageIDWindowsEnglishCaribbean:
			return "Windows English-Caribbean"
		case LanguageIDWindowsEnglishIndia:
			return "Windows English-India"
		case LanguageIDWindowsEnglishIreland:
			return "Windows English-Ireland"
		case LanguageIDWindowsEnglishJamaica:
			return "Windows English-Jamaica"
		case LanguageIDWindowsEnglishMalaysia:
			return "Windows English-Malaysia"
		case LanguageIDWindowsEnglishNewZealand:
			return "Windows English-New Zealand"
		case LanguageIDWindowsEnglishRepublicofthePhilippines:
			return "Windows English-Republic of the Philippines"
		case LanguageIDWindowsEnglishSingapore:
			return "Windows English-Singapore"
		case LanguageIDWindowsEnglishSouthAfrica:
			return "Windows English-South Africa"
		case LanguageIDWindowsEnglishTrinidadandTobago:
			return "Windows English-Trinidad and Tobago"
		case LanguageIDWindowsEnglishUnitedKingdom:
			return "Windows English-United Kingdom"
		case LanguageIDWindowsEnglishUnitedStates:
			return "Windows English-United States"
		case LanguageIDWindowsEnglishZimbabwe:
			return "Windows English-Zimbabwe"
		case LanguageIDWindowsEstonianEstonia:
			return "Windows Estonian-Estonia"
		case LanguageIDWindowsFaroeseFaroeIslands:
			return "Windows Faroese-Faroe Islands"
		case LanguageIDWindowsFilipinoPhilippines:
			return "Windows Filipino-Philippines"
		case LanguageIDWindowsFinnishFinland:
			return "Windows Finnish-Finland"
		case LanguageIDWindowsFrenchBelgium:
			return "Windows French-Belgium"
		case LanguageIDWindowsFrenchCanada:
			return "Windows French-Canada"
		case LanguageIDWindowsFrenchFrance:
			return "Windows French-France"
		case LanguageIDWindowsFrenchLuxembourg:
			return "Windows French-Luxembourg"
		case LanguageIDWindowsFrenchPrincipalityofMonaco:
			return "Windows French-Principality of Monaco"
		case LanguageIDWindowsFrenchSwitzerland:
			return "Windows French-Switzerland"
		case LanguageIDWindowsFrisianNetherlands:
			return "Windows Frisian-Netherlands"
		case LanguageIDWindowsGalicianGalician:
			return "Windows Galician-Galician"
		case LanguageIDWindowsGeorgianGeorgia:
			return "Windows Georgian-Georgia"
		case LanguageIDWindowsGermanAustria:
			return "Windows German-Austria"
		case LanguageIDWindowsGermanGermany:
			return "Windows German-Germany"
		case LanguageIDWindowsGermanLiechtenstein:
			return "Windows German-Liechtenstein"
		case LanguageIDWindowsGermanLuxembourg:
			return "Windows German-Luxembourg"
		case LanguageIDWindowsGermanSwitzerland:
			return "Windows German-Switzerland"
		case LanguageIDWindowsGreekGreece:
			return "Windows Greek-Greece"
		case LanguageIDWindowsGreenlandicGreenland:
			return "Windows Greenlandic-Greenland"
		case LanguageIDWindowsGujaratiIndia:
			return "Windows Gujarati-India"
		case LanguageIDWindowsHausaLatinNigeria:
			return "Windows Hausa (Latin)-Nigeria"
		case LanguageIDWindowsHebrewIsrael:
			return "Windows Hebrew-Israel"
		case LanguageIDWindowsHindiIndia:
			return "Windows Hindi-India"
		case LanguageIDWindowsHungarianHungary:
			return "Windows Hungarian-Hungary"
		case LanguageIDWindowsIcelandicIceland:
			return "Windows Icelandic-Iceland"
		case LanguageIDWindowsIgboNigeria:
			return "Windows Igbo-Nigeria"
		case LanguageIDWindowsIndonesianIndonesia:
			return "Windows Indonesian-Indonesia"
		case LanguageIDWindowsInuktitutCanada:
			return "Windows Inuktitut-Canada"
		case LanguageIDWindowsInuktitutLatinCanada:
			return "Windows Inuktitut (Latin)-Canada"
		case LanguageIDWindowsIrishIreland:
			return "Windows Irish-Ireland"
		case LanguageIDWindowsisiXhosaSouthAfrica:
			return "Windows isiXhosa-South Africa"
		case LanguageIDWindowsisiZuluSouthAfrica:
			return "Windows isiZulu-South Africa"
		case LanguageIDWindowsItalianItaly:
			return "Windows Italian-Italy"
		case LanguageIDWindowsItalianSwitzerland:
			return "Windows Italian-Switzerland"
		case LanguageIDWindowsJapaneseJapan:
			return "Windows Japanese-Japan"
		case LanguageIDWindowsKannadaIndia:
			return "Windows Kannada-India"
		case LanguageIDWindowsKazakhKazakhstan:
			return "Windows Kazakh-Kazakhstan"
		case LanguageIDWindowsKhmerCambodia:
			return "Windows Khmer-Cambodia"
		case LanguageIDWindowsKicheGuatemala:
			return "Windows K'iche-Guatemala"
		case LanguageIDWindowsKinyarwandaRwanda:
			return "Windows Kinyarwanda-Rwanda"
		case LanguageIDWindowsKiswahiliKenya:
			return "Windows Kiswahili-Kenya"
		case LanguageIDWindowsKonkaniIndia:
			return "Windows Konkani-India"
		case LanguageIDWindowsKoreanKorea:
			return "Windows Korean-Korea"
		case LanguageIDWindowsKyrgyzKyrgyzstan:
			return "Windows Kyrgyz-Kyrgyzstan"
		case LanguageIDWindowsLaoLaoPDR:
			return "Windows Lao-Lao P.D.R."
		case LanguageIDWindowsLatvianLatvia:
			return "Windows Latvian-Latvia"
		case LanguageIDWindowsLithuanianLithuania:
			return "Windows Lithuanian-Lithuania"
		case LanguageIDWindowsLowerSorbianGermany:
			return "Windows Lower Sorbian-Germany"
		case LanguageIDWindowsLuxembourgishLuxembourg:
			return "Windows Luxembourgish-Luxembourg"
		case LanguageIDWindowsMacedonianFYROMFormerYugoslavRepublicofMacedonia:
			return "Windows Macedonian (FYROM)-Former Yugoslav Republic of Macedonia"
		case LanguageIDWindowsMalayBruneiDarussalam:
			return "Windows Malay-Brunei Darussalam"
		case LanguageIDWindowsMalayMalaysia:
			return "Windows Malay-Malaysia"
		case LanguageIDWindowsMalayalamIndia:
			return "Windows Malayalam-India"
		case LanguageIDWindowsMalteseMalta:
			return "Windows Maltese-Malta"
		case LanguageIDWindowsMaoriNewZealand:
			return "Windows Maori-New Zealand"
		case LanguageIDWindowsMapudungunChile:
			return "Windows Mapudungun-Chile"
		case LanguageIDWindowsMarathiIndia:
			return "Windows Marathi-India"
		case LanguageIDWindowsMohawkMohawk:
			return "Windows Mohawk-Mohawk"
		case LanguageIDWindowsMongolianCyrillicMongolia:
			return "Windows Mongolian (Cyrillic)-Mongolia"
		case LanguageIDWindowsMongolianTraditionalPeoplesRepublicofChina:
			return "Windows Mongolian (Traditional)-People's Republic of China"
		case LanguageIDWindowsNepaliNepal:
			return "Windows Nepali-Nepal"
		case LanguageIDWindowsNorwegianBokmalNorway:
			return "Windows Norwegian (Bokmal)-Norway"
		case LanguageIDWindowsNorwegianNynorskNorway:
			return "Windows Norwegian (Nynorsk)-Norway"
		case LanguageIDWindowsOccitanFrance:
			return "Windows Occitan-France"
		case LanguageIDWindowsOdiaformerlyOriyaIndia:
			return "Windows Odia (formerly Oriya)-India"
		case LanguageIDWindowsPashtoAfghanistan:
			return "Windows Pashto-Afghanistan"
		case LanguageIDWindowsPolishPoland:
			return "Windows Polish-Poland"
		case LanguageIDWindowsPortugueseBrazil:
			return "Windows Portuguese-Brazil"
		case LanguageIDWindowsPortuguesePortugal:
			return "Windows Portuguese-Portugal"
		case LanguageIDWindowsPunjabiIndia:
			return "Windows Punjabi-India"
		case LanguageIDWindowsQuechuaBolivia:
			return "Windows Quechua-Bolivia"
		case LanguageIDWindowsQuechuaEcuador:
			return "Windows Quechua-Ecuador"
		case LanguageIDWindowsQuechuaPeru:
			return "Windows Quechua-Peru"
		case LanguageIDWindowsRomanianRomania:
			return "Windows Romanian-Romania"
		case LanguageIDWindowsRomanshSwitzerland:
			return "Windows Romansh-Switzerland"
		case LanguageIDWindowsRussianRussia:
			return "Windows Russian-Russia"
		case LanguageIDWindowsSamiInariFinland:
			return "Windows Sami (Inari)-Finland"
		case LanguageIDWindowsSamiLuleNorway:
			return "Windows Sami (Lule)-Norway"
		case LanguageIDWindowsSamiLuleSweden:
			return "Windows Sami (Lule)-Sweden"
		case LanguageIDWindowsSamiNorthernFinland:
			return "Windows Sami (Northern)-Finland"
		case LanguageIDWindowsSamiNorthernNorway:
			return "Windows Sami (Northern)-Norway"
		case LanguageIDWindowsSamiNorthernSweden:
			return "Windows Sami (Northern)-Sweden"
		case LanguageIDWindowsSamiSkoltFinland:
			return "Windows Sami (Skolt)-Finland"
		case LanguageIDWindowsSamiSouthernNorway:
			return "Windows Sami (Southern)-Norway"
		case LanguageIDWindowsSamiSouthernSweden:
			return "Windows Sami (Southern)-Sweden"
		case LanguageIDWindowsSanskritIndia:
			return "Windows Sanskrit-India"
		case LanguageIDWindowsSerbianCyrillicBosniaandHerzegovina:
			return "Windows Serbian (Cyrillic)-Bosnia and Herzegovina"
		case LanguageIDWindowsSerbianCyrillicSerbia:
			return "Windows Serbian (Cyrillic)-Serbia"
		case LanguageIDWindowsSerbianLatinBosniaandHerzegovina:
			return "Windows Serbian (Latin)-Bosnia and Herzegovina"
		case LanguageIDWindowsSerbianLatinSerbia:
			return "Windows Serbian (Latin)-Serbia"
		case LanguageIDWindowsSesothosaLeboaSouthAfrica:
			return "Windows Sesotho sa Leboa-South Africa"
		case LanguageIDWindowsSetswanaSouthAfrica:
			return "Windows Setswana-South Africa"
		case LanguageIDWindowsSinhalaSriLanka:
			return "Windows Sinhala-Sri Lanka"
		case LanguageIDWindowsSlovakSlovakia:
			return "Windows Slovak-Slovakia"
		case LanguageIDWindowsSlovenianSlovenia:
			return "Windows Slovenian-Slovenia"
		case LanguageIDWindowsSpanishArgentina:
			return "Windows Spanish-Argentina"
		case LanguageIDWindowsSpanishBolivia:
			return "Windows Spanish-Bolivia"
		case LanguageIDWindowsSpanishChile:
			return "Windows Spanish-Chile"
		case LanguageIDWindowsSpanishColombia:
			return "Windows Spanish-Colombia"
		case LanguageIDWindowsSpanishCostaRica:
			return "Windows Spanish-Costa Rica"
		case LanguageIDWindowsSpanishDominicanRepublic:
			return "Windows Spanish-Dominican Republic"
		case LanguageIDWindowsSpanishEcuador:
			return "Windows Spanish-Ecuador"
		case LanguageIDWindowsSpanishElSalvador:
			return "Windows Spanish-El Salvador"
		case LanguageIDWindowsSpanishGuatemala:
			return "Windows Spanish-Guatemala"
		case LanguageIDWindowsSpanishHonduras:
			return "Windows Spanish-Honduras"
		case LanguageIDWindowsSpanishMexico:
			return "Windows Spanish-Mexico"
		case LanguageIDWindowsSpanishNicaragua:
			return "Windows Spanish-Nicaragua"
		case LanguageIDWindowsSpanishPanama:
			return "Windows Spanish-Panama"
		case LanguageIDWindowsSpanishParaguay:
			return "Windows Spanish-Paraguay"
		case LanguageIDWindowsSpanishPeru:
			return "Windows Spanish-Peru"
		case LanguageIDWindowsSpanishPuertoRico:
			return "Windows Spanish-Puerto Rico"
		case LanguageIDWindowsSpanishModernSortSpain:
			return "Windows Spanish (Modern Sort)-Spain"
		case LanguageIDWindowsSpanishTraditionalSortSpain:
			return "Windows Spanish (Traditional Sort)-Spain"
		case LanguageIDWindowsSpanishUnitedStates:
			return "Windows Spanish-United States"
		case LanguageIDWindowsSpanishUruguay:
			return "Windows Spanish-Uruguay"
		case LanguageIDWindowsSpanishVenezuela:
			return "Windows Spanish-Venezuela"
		case LanguageIDWindowsSwedenFinland:
			return "Windows Sweden-Finland"
		case LanguageIDWindowsSwedishSweden:
			return "Windows Swedish-Sweden"
		case LanguageIDWindowsSyriacSyria:
			return "Windows Syriac-Syria"
		case LanguageIDWindowsTajikCyrillicTajikistan:
			return "Windows Tajik (Cyrillic)-Tajikistan"
		case LanguageIDWindowsTamazightLatinAlgeria:
			return "Windows Tamazight (Latin)-Algeria"
		case LanguageIDWindowsTamilIndia:
			return "Windows Tamil-India"
		case LanguageIDWindowsTatarRussia:
			return "Windows Tatar-Russia"
		case LanguageIDWindowsTeluguIndia:
			return "Windows Telugu-India"
		case LanguageIDWindowsThaiThailand:
			return "Windows Thai-Thailand"
		case LanguageIDWindowsTibetanPRC:
			return "Windows Tibetan-PRC"
		case LanguageIDWindowsTurkishTurkey:
			return "Windows Turkish-Turkey"
		case LanguageIDWindowsTurkmenTurkmenistan:
			return "Windows Turkmen-Turkmenistan"
		case LanguageIDWindowsUighurPRC:
			return "Windows Uighur-PRC"
		case LanguageIDWindowsUkrainianUkraine:
			return "Windows Ukrainian-Ukraine"
		case LanguageIDWindowsUpperSorbianGermany:
			return "Windows Upper Sorbian-Germany"
		case LanguageIDWindowsUrduIslamicRepublicofPakistan:
			return "Windows Urdu-Islamic Republic of Pakistan"
		case LanguageIDWindowsUzbekCyrillicUzbekistan:
			return "Windows Uzbek (Cyrillic)-Uzbekistan"
		case LanguageIDWindowsUzbekLatinUzbekistan:
			return "Windows Uzbek (Latin)-Uzbekistan"
		case LanguageIDWindowsVietnameseVietnam:
			return "Windows Vietnamese-Vietnam"
		case LanguageIDWindowsWelshUnitedKingdom:
			return "Windows Welsh-United Kingdom"
		case LanguageIDWindowsWolofSenegal:
			return "Windows Wolof-Senegal"
		case LanguageIDWindowsYakutRussia:
			return "Windows Yakut-Russia"
		case LanguageIDWindowsYiPRC:
			return "Windows Yi-PRC"
		case LanguageIDWindowsYorubaNigeria:
			return "Windows Yoruba-Nigeria"
		}
	case PlatformIDMacintosh:
		switch l {
		case LanguageIDMacintoshEnglish:
			return "Macintosh English"
		case LanguageIDMacintoshFrench:
			return "Macintosh French"
		case LanguageIDMacintoshGerman:
			return "Macintosh German"
		case LanguageIDMacintoshItalian:
			return "Macintosh Italian"
		case LanguageIDMacintoshDutch:
			return "Macintosh Dutch"
		case LanguageIDMacintoshSwedish:
			return "Macintosh Swedish"
		case LanguageIDMacintoshSpanish:
			return "Macintosh Spanish"
		case LanguageIDMacintoshDanish:
			return "Macintosh Danish"
		case LanguageIDMacintoshPortuguese:
			return "Macintosh Portuguese"
		case LanguageIDMacintoshNorwegian:
			return "Macintosh Norwegian"
		case LanguageIDMacintoshHebrew:
			return "Macintosh Hebrew"
		case LanguageIDMacintoshJapanese:
			return "Macintosh Japanese"
		case LanguageIDMacintoshArabic:
			return "Macintosh Arabic"
		case LanguageIDMacintoshFinnish:
			return "Macintosh Finnish"
		case LanguageIDMacintoshGreek:
			return "Macintosh Greek"
		case LanguageIDMacintoshIcelandic:
			return "Macintosh Icelandic"
		case LanguageIDMacintoshMaltese:
			return "Macintosh Maltese"
		case LanguageIDMacintoshTurkish:
			return "Macintosh Turkish"
		case LanguageIDMacintoshCroatian:
			return "Macintosh Croatian"
		case LanguageIDMacintoshChineseTraditional:
			return "Macintosh Chinese (Traditional)"
		case LanguageIDMacintoshUrdu:
			return "Macintosh Urdu"
		case LanguageIDMacintoshHindi:
			return "Macintosh Hindi"
		case LanguageIDMacintoshThai:
			return "Macintosh Thai"
		case LanguageIDMacintoshKorean:
			return "Macintosh Korean"
		case LanguageIDMacintoshLithuanian:
			return "Macintosh Lithuanian"
		case LanguageIDMacintoshPolish:
			return "Macintosh Polish"
		case LanguageIDMacintoshHungarian:
			return "Macintosh Hungarian"
		case LanguageIDMacintoshEstonian:
			return "Macintosh Estonian"
		case LanguageIDMacintoshLatvian:
			return "Macintosh Latvian"
		case LanguageIDMacintoshSami:
			return "Macintosh Sami"
		case LanguageIDMacintoshFaroese:
			return "Macintosh Faroese"
		case LanguageIDMacintoshFarsiPersian:
			return "Macintosh Farsi/Persian"
		case LanguageIDMacintoshRussian:
			return "Macintosh Russian"
		case LanguageIDMacintoshChineseSimplified:
			return "Macintosh Chinese (Simplified)"
		case LanguageIDMacintoshFlemish:
			return "Macintosh Flemish"
		case LanguageIDMacintoshIrishGaelic:
			return "Macintosh Irish Gaelic"
		case LanguageIDMacintoshAlbanian:
			return "Macintosh Albanian"
		case LanguageIDMacintoshRomanian:
			return "Macintosh Romanian"
		case LanguageIDMacintoshCzech:
			return "Macintosh Czech"
		case LanguageIDMacintoshSlovak:
			return "Macintosh Slovak"
		case LanguageIDMacintoshSlovenian:
			return "Macintosh Slovenian"
		case LanguageIDMacintoshYiddish:
			return "Macintosh Yiddish"
		case LanguageIDMacintoshSerbian:
			return "Macintosh Serbian"
		case LanguageIDMacintoshMacedonian:
			return "Macintosh Macedonian"
		case LanguageIDMacintoshBulgarian:
			return "Macintosh Bulgarian"
		case LanguageIDMacintoshUkrainian:
			return "Macintosh Ukrainian"
		case LanguageIDMacintoshByelorussian:
			return "Macintosh Byelorussian"
		case LanguageIDMacintoshUzbek:
			return "Macintosh Uzbek"
		case LanguageIDMacintoshKazakh:
			return "Macintosh Kazakh"
		case LanguageIDMacintoshAzerbaijaniCyrillicscript:
			return "Macintosh Azerbaijani (Cyrillic script)"
		case LanguageIDMacintoshAzerbaijaniArabicscript:
			return "Macintosh Azerbaijani (Arabic script)"
		case LanguageIDMacintoshArmenian:
			return "Macintosh Armenian"
		case LanguageIDMacintoshGeorgian:
			return "Macintosh Georgian"
		case LanguageIDMacintoshMoldavian:
			return "Macintosh Moldavian"
		case LanguageIDMacintoshKirghiz:
			return "Macintosh Kirghiz"
		case LanguageIDMacintoshTajiki:
			return "Macintosh Tajiki"
		case LanguageIDMacintoshTurkmen:
			return "Macintosh Turkmen"
		case LanguageIDMacintoshMongolianMongolianscript:
			return "Macintosh Mongolian (Mongolian script)"
		case LanguageIDMacintoshMongolianCyrillicscript:
			return "Macintosh Mongolian (Cyrillic script)"
		case LanguageIDMacintoshPashto:
			return "Macintosh Pashto"
		case LanguageIDMacintoshKurdish:
			return "Macintosh Kurdish"
		case LanguageIDMacintoshKashmiri:
			return "Macintosh Kashmiri"
		case LanguageIDMacintoshSindhi:
			return "Macintosh Sindhi"
		case LanguageIDMacintoshTibetan:
			return "Macintosh Tibetan"
		case LanguageIDMacintoshNepali:
			return "Macintosh Nepali"
		case LanguageIDMacintoshSanskrit:
			return "Macintosh Sanskrit"
		case LanguageIDMacintoshMarathi:
			return "Macintosh Marathi"
		case LanguageIDMacintoshBengali:
			return "Macintosh Bengali"
		case LanguageIDMacintoshAssamese:
			return "Macintosh Assamese"
		case LanguageIDMacintoshGujarati:
			return "Macintosh Gujarati"
		case LanguageIDMacintoshPunjabi:
			return "Macintosh Punjabi"
		case LanguageIDMacintoshOriya:
			return "Macintosh Oriya"
		case LanguageIDMacintoshMalayalam:
			return "Macintosh Malayalam"
		case LanguageIDMacintoshKannada:
			return "Macintosh Kannada"
		case LanguageIDMacintoshTamil:
			return "Macintosh Tamil"
		case LanguageIDMacintoshTelugu:
			return "Macintosh Telugu"
		case LanguageIDMacintoshSinhalese:
			return "Macintosh Sinhalese"
		case LanguageIDMacintoshBurmese:
			return "Macintosh Burmese"
		case LanguageIDMacintoshKhmer:
			return "Macintosh Khmer"
		case LanguageIDMacintoshLao:
			return "Macintosh Lao"
		case LanguageIDMacintoshVietnamese:
			return "Macintosh Vietnamese"
		case LanguageIDMacintoshIndonesian:
			return "Macintosh Indonesian"
		case LanguageIDMacintoshTagalong:
			return "Macintosh Tagalong"
		case LanguageIDMacintoshMalayRomanscript:
			return "Macintosh Malay (Roman script)"
		case LanguageIDMacintoshMalayArabicscript:
			return "Macintosh Malay (Arabic script)"
		case LanguageIDMacintoshAmharic:
			return "Macintosh Amharic"
		case LanguageIDMacintoshTigrinya:
			return "Macintosh Tigrinya"
		case LanguageIDMacintoshGalla:
			return "Macintosh Galla"
		case LanguageIDMacintoshSomali:
			return "Macintosh Somali"
		case LanguageIDMacintoshSwahili:
			return "Macintosh Swahili"
		case LanguageIDMacintoshKinyarwandaRuanda:
			return "Macintosh Kinyarwanda/Ruanda"
		case LanguageIDMacintoshRundi:
			return "Macintosh Rundi"
		case LanguageIDMacintoshNyanjaChewa:
			return "Macintosh Nyanja/Chewa"
		case LanguageIDMacintoshMalagasy:
			return "Macintosh Malagasy"
		case LanguageIDMacintoshEsperanto:
			return "Macintosh Esperanto"
		case LanguageIDMacintoshWelsh:
			return "Macintosh Welsh"
		case LanguageIDMacintoshBasque:
			return "Macintosh Basque"
		case LanguageIDMacintoshCatalan:
			return "Macintosh Catalan"
		case LanguageIDMacintoshLatin:
			return "Macintosh Latin"
		case LanguageIDMacintoshQuenchua:
			return "Macintosh Quenchua"
		case LanguageIDMacintoshGuarani:
			return "Macintosh Guarani"
		case LanguageIDMacintoshAymara:
			return "Macintosh Aymara"
		case LanguageIDMacintoshTatar:
			return "Macintosh Tatar"
		case LanguageIDMacintoshUighur:
			return "Macintosh Uighur"
		case LanguageIDMacintoshDzongkha:
			return "Macintosh Dzongkha"
		case LanguageIDMacintoshJavaneseRomanscript:
			return "Macintosh Javanese (Roman script)"
		case LanguageIDMacintoshSundaneseRomanscript:
			return "Macintosh Sundanese (Roman script)"
		case LanguageIDMacintoshGalician:
			return "Macintosh Galician"
		case LanguageIDMacintoshAfrikaans:
			return "Macintosh Afrikaans"
		case LanguageIDMacintoshBreton:
			return "Macintosh Breton"
		case LanguageIDMacintoshInuktitut:
			return "Macintosh Inuktitut"
		case LanguageIDMacintoshScottishGaelic:
			return "Macintosh Scottish Gaelic"
		case LanguageIDMacintoshManxGaelic:
			return "Macintosh Manx Gaelic"
		case LanguageIDMacintoshIrishGaelicwithdotabove:
			return "Macintosh Irish Gaelic (with dot above)"
		case LanguageIDMacintoshTongan:
			return "Macintosh Tongan"
		case LanguageIDMacintoshGreekpolytonic:
			return "Macintosh Greek (polytonic)"
		case LanguageIDMacintoshGreenlandic:
			return "Macintosh Greenlandic"
		case LanguageIDMacintoshAzerbaijaniRomanscript:
			return "Macintosh Azerbaijani (Roman script)"
		}
	case PlatformIDISO:
		return languageIDISO
	}
	return unknownLanguageID
}

func (l LanguageID) String(p PlatformID) string {
	return fmt.Sprintf("(%d):%s", l, l.Name(p))
}

const (
	// LanguageIDWindowsAfrikaansSouthAfrica : Windows Afrikaans-South Africa
	LanguageIDWindowsAfrikaansSouthAfrica = LanguageID(0x0436)
	// LanguageIDWindowsAlbanianAlbania : Windows Albanian-Albania
	LanguageIDWindowsAlbanianAlbania = LanguageID(0x041C)
	// LanguageIDWindowsAlsatianFrance : Windows Alsatian-France
	LanguageIDWindowsAlsatianFrance = LanguageID(0x0484)
	// LanguageIDWindowsAmharicEthiopia : Windows Amharic-Ethiopia
	LanguageIDWindowsAmharicEthiopia = LanguageID(0x045E)
	// LanguageIDWindowsArabicAlgeria : Windows Arabic-Algeria
	LanguageIDWindowsArabicAlgeria = LanguageID(0x1401)
	// LanguageIDWindowsArabicBahrain : Windows Arabic-Bahrain
	LanguageIDWindowsArabicBahrain = LanguageID(0x3C01)
	// LanguageIDWindowsArabicEgypt : Windows Arabic-Egypt
	LanguageIDWindowsArabicEgypt = LanguageID(0x0C01)
	// LanguageIDWindowsArabicIraq : Windows Arabic-Iraq
	LanguageIDWindowsArabicIraq = LanguageID(0x0801)
	// LanguageIDWindowsArabicJordan : Windows Arabic-Jordan
	LanguageIDWindowsArabicJordan = LanguageID(0x2C01)
	// LanguageIDWindowsArabicKuwait : Windows Arabic-Kuwait
	LanguageIDWindowsArabicKuwait = LanguageID(0x3401)
	// LanguageIDWindowsArabicLebanon : Windows Arabic-Lebanon
	LanguageIDWindowsArabicLebanon = LanguageID(0x3001)
	// LanguageIDWindowsArabicLibya : Windows Arabic-Libya
	LanguageIDWindowsArabicLibya = LanguageID(0x1001)
	// LanguageIDWindowsArabicMorocco : Windows Arabic-Morocco
	LanguageIDWindowsArabicMorocco = LanguageID(0x1801)
	// LanguageIDWindowsArabicOman : Windows Arabic-Oman
	LanguageIDWindowsArabicOman = LanguageID(0x2001)
	// LanguageIDWindowsArabicQatar : Windows Arabic-Qatar
	LanguageIDWindowsArabicQatar = LanguageID(0x4001)
	// LanguageIDWindowsArabicSaudiArabia : Windows Arabic-Saudi Arabia
	LanguageIDWindowsArabicSaudiArabia = LanguageID(0x0401)
	// LanguageIDWindowsArabicSyria : Windows Arabic-Syria
	LanguageIDWindowsArabicSyria = LanguageID(0x2801)
	// LanguageIDWindowsArabicTunisia : Windows Arabic-Tunisia
	LanguageIDWindowsArabicTunisia = LanguageID(0x1C01)
	// LanguageIDWindowsArabicUAE : Windows Arabic-U.A.E.
	LanguageIDWindowsArabicUAE = LanguageID(0x3801)
	// LanguageIDWindowsArabicYemen : Windows Arabic-Yemen
	LanguageIDWindowsArabicYemen = LanguageID(0x2401)
	// LanguageIDWindowsArmenianArmenia : Windows Armenian-Armenia
	LanguageIDWindowsArmenianArmenia = LanguageID(0x042B)
	// LanguageIDWindowsAssameseIndia : Windows Assamese-India
	LanguageIDWindowsAssameseIndia = LanguageID(0x044D)
	// LanguageIDWindowsAzeriCyrillicAzerbaijan : Windows Azeri (Cyrillic)-Azerbaijan
	LanguageIDWindowsAzeriCyrillicAzerbaijan = LanguageID(0x082C)
	// LanguageIDWindowsAzeriLatinAzerbaijan : Windows Azeri (Latin)-Azerbaijan
	LanguageIDWindowsAzeriLatinAzerbaijan = LanguageID(0x042C)
	// LanguageIDWindowsBashkirRussia : Windows Bashkir-Russia
	LanguageIDWindowsBashkirRussia = LanguageID(0x046D)
	// LanguageIDWindowsBasqueBasque : Windows Basque-Basque
	LanguageIDWindowsBasqueBasque = LanguageID(0x042D)
	// LanguageIDWindowsBelarusianBelarus : Windows Belarusian-Belarus
	LanguageIDWindowsBelarusianBelarus = LanguageID(0x0423)
	// LanguageIDWindowsBengaliBangladesh : Windows Bengali-Bangladesh
	LanguageIDWindowsBengaliBangladesh = LanguageID(0x0845)
	// LanguageIDWindowsBengaliIndia : Windows Bengali-India
	LanguageIDWindowsBengaliIndia = LanguageID(0x0445)
	// LanguageIDWindowsBosnianCyrillicBosniaandHerzegovina : Windows Bosnian (Cyrillic)-Bosnia and Herzegovina
	LanguageIDWindowsBosnianCyrillicBosniaandHerzegovina = LanguageID(0x201A)
	// LanguageIDWindowsBosnianLatinBosniaandHerzegovina : Windows Bosnian (Latin)-Bosnia and Herzegovina
	LanguageIDWindowsBosnianLatinBosniaandHerzegovina = LanguageID(0x141A)
	// LanguageIDWindowsBretonFrance : Windows Breton-France
	LanguageIDWindowsBretonFrance = LanguageID(0x047E)
	// LanguageIDWindowsBulgarianBulgaria : Windows Bulgarian-Bulgaria
	LanguageIDWindowsBulgarianBulgaria = LanguageID(0x0402)
	// LanguageIDWindowsCatalanCatalan : Windows Catalan-Catalan
	LanguageIDWindowsCatalanCatalan = LanguageID(0x0403)
	// LanguageIDWindowsChineseHongKongSAR : Windows Chinese-Hong Kong S.A.R.
	LanguageIDWindowsChineseHongKongSAR = LanguageID(0x0C04)
	// LanguageIDWindowsChineseMacaoSAR : Windows Chinese-Macao S.A.R.
	LanguageIDWindowsChineseMacaoSAR = LanguageID(0x1404)
	// LanguageIDWindowsChinesePeoplesRepublicofChina : Windows Chinese-People's Republic of China
	LanguageIDWindowsChinesePeoplesRepublicofChina = LanguageID(0x0804)
	// LanguageIDWindowsChineseSingapore : Windows Chinese-Singapore
	LanguageIDWindowsChineseSingapore = LanguageID(0x1004)
	// LanguageIDWindowsChineseTaiwan : Windows Chinese-Taiwan
	LanguageIDWindowsChineseTaiwan = LanguageID(0x0404)
	// LanguageIDWindowsCorsicanFrance : Windows Corsican-France
	LanguageIDWindowsCorsicanFrance = LanguageID(0x0483)
	// LanguageIDWindowsCroatianCroatia : Windows Croatian-Croatia
	LanguageIDWindowsCroatianCroatia = LanguageID(0x041A)
	// LanguageIDWindowsCroatianLatinBosniaandHerzegovina : Windows Croatian (Latin)-Bosnia and Herzegovina
	LanguageIDWindowsCroatianLatinBosniaandHerzegovina = LanguageID(0x101A)
	// LanguageIDWindowsCzechCzechRepublic : Windows Czech-Czech Republic
	LanguageIDWindowsCzechCzechRepublic = LanguageID(0x0405)
	// LanguageIDWindowsDanishDenmark : Windows Danish-Denmark
	LanguageIDWindowsDanishDenmark = LanguageID(0x0406)
	// LanguageIDWindowsDariAfghanistan : Windows Dari-Afghanistan
	LanguageIDWindowsDariAfghanistan = LanguageID(0x048C)
	// LanguageIDWindowsDivehiMaldives : Windows Divehi-Maldives
	LanguageIDWindowsDivehiMaldives = LanguageID(0x0465)
	// LanguageIDWindowsDutchBelgium : Windows Dutch-Belgium
	LanguageIDWindowsDutchBelgium = LanguageID(0x0813)
	// LanguageIDWindowsDutchNetherlands : Windows Dutch-Netherlands
	LanguageIDWindowsDutchNetherlands = LanguageID(0x0413)
	// LanguageIDWindowsEnglishAustralia : Windows English-Australia
	LanguageIDWindowsEnglishAustralia = LanguageID(0x0C09)
	// LanguageIDWindowsEnglishBelize : Windows English-Belize
	LanguageIDWindowsEnglishBelize = LanguageID(0x2809)
	// LanguageIDWindowsEnglishCanada : Windows English-Canada
	LanguageIDWindowsEnglishCanada = LanguageID(0x1009)
	// LanguageIDWindowsEnglishCaribbean : Windows English-Caribbean
	LanguageIDWindowsEnglishCaribbean = LanguageID(0x2409)
	// LanguageIDWindowsEnglishIndia : Windows English-India
	LanguageIDWindowsEnglishIndia = LanguageID(0x4009)
	// LanguageIDWindowsEnglishIreland : Windows English-Ireland
	LanguageIDWindowsEnglishIreland = LanguageID(0x1809)
	// LanguageIDWindowsEnglishJamaica : Windows English-Jamaica
	LanguageIDWindowsEnglishJamaica = LanguageID(0x2009)
	// LanguageIDWindowsEnglishMalaysia : Windows English-Malaysia
	LanguageIDWindowsEnglishMalaysia = LanguageID(0x4409)
	// LanguageIDWindowsEnglishNewZealand : Windows English-New Zealand
	LanguageIDWindowsEnglishNewZealand = LanguageID(0x1409)
	// LanguageIDWindowsEnglishRepublicofthePhilippines : Windows English-Republic of the Philippines
	LanguageIDWindowsEnglishRepublicofthePhilippines = LanguageID(0x3409)
	// LanguageIDWindowsEnglishSingapore : Windows English-Singapore
	LanguageIDWindowsEnglishSingapore = LanguageID(0x4809)
	// LanguageIDWindowsEnglishSouthAfrica : Windows English-South Africa
	LanguageIDWindowsEnglishSouthAfrica = LanguageID(0x1C09)
	// LanguageIDWindowsEnglishTrinidadandTobago : Windows English-Trinidad and Tobago
	LanguageIDWindowsEnglishTrinidadandTobago = LanguageID(0x2C09)
	// LanguageIDWindowsEnglishUnitedKingdom : Windows English-United Kingdom
	LanguageIDWindowsEnglishUnitedKingdom = LanguageID(0x0809)
	// LanguageIDWindowsEnglishUnitedStates : Windows English-United States
	LanguageIDWindowsEnglishUnitedStates = LanguageID(0x0409)
	// LanguageIDWindowsEnglishZimbabwe : Windows English-Zimbabwe
	LanguageIDWindowsEnglishZimbabwe = LanguageID(0x3009)
	// LanguageIDWindowsEstonianEstonia : Windows Estonian-Estonia
	LanguageIDWindowsEstonianEstonia = LanguageID(0x0425)
	// LanguageIDWindowsFaroeseFaroeIslands : Windows Faroese-Faroe Islands
	LanguageIDWindowsFaroeseFaroeIslands = LanguageID(0x0438)
	// LanguageIDWindowsFilipinoPhilippines : Windows Filipino-Philippines
	LanguageIDWindowsFilipinoPhilippines = LanguageID(0x0464)
	// LanguageIDWindowsFinnishFinland : Windows Finnish-Finland
	LanguageIDWindowsFinnishFinland = LanguageID(0x040B)
	// LanguageIDWindowsFrenchBelgium : Windows French-Belgium
	LanguageIDWindowsFrenchBelgium = LanguageID(0x080C)
	// LanguageIDWindowsFrenchCanada : Windows French-Canada
	LanguageIDWindowsFrenchCanada = LanguageID(0x0C0C)
	// LanguageIDWindowsFrenchFrance : Windows French-France
	LanguageIDWindowsFrenchFrance = LanguageID(0x040C)
	// LanguageIDWindowsFrenchLuxembourg : Windows French-Luxembourg
	LanguageIDWindowsFrenchLuxembourg = LanguageID(0x140c)
	// LanguageIDWindowsFrenchPrincipalityofMonaco : Windows French-Principality of Monaco
	LanguageIDWindowsFrenchPrincipalityofMonaco = LanguageID(0x180C)
	// LanguageIDWindowsFrenchSwitzerland : Windows French-Switzerland
	LanguageIDWindowsFrenchSwitzerland = LanguageID(0x100C)
	// LanguageIDWindowsFrisianNetherlands : Windows Frisian-Netherlands
	LanguageIDWindowsFrisianNetherlands = LanguageID(0x0462)
	// LanguageIDWindowsGalicianGalician : Windows Galician-Galician
	LanguageIDWindowsGalicianGalician = LanguageID(0x0456)
	// LanguageIDWindowsGeorgianGeorgia : Windows Georgian-Georgia
	LanguageIDWindowsGeorgianGeorgia = LanguageID(0x0437)
	// LanguageIDWindowsGermanAustria : Windows German-Austria
	LanguageIDWindowsGermanAustria = LanguageID(0x0C07)
	// LanguageIDWindowsGermanGermany : Windows German-Germany
	LanguageIDWindowsGermanGermany = LanguageID(0x0407)
	// LanguageIDWindowsGermanLiechtenstein : Windows German-Liechtenstein
	LanguageIDWindowsGermanLiechtenstein = LanguageID(0x1407)
	// LanguageIDWindowsGermanLuxembourg : Windows German-Luxembourg
	LanguageIDWindowsGermanLuxembourg = LanguageID(0x1007)
	// LanguageIDWindowsGermanSwitzerland : Windows German-Switzerland
	LanguageIDWindowsGermanSwitzerland = LanguageID(0x0807)
	// LanguageIDWindowsGreekGreece : Windows Greek-Greece
	LanguageIDWindowsGreekGreece = LanguageID(0x0408)
	// LanguageIDWindowsGreenlandicGreenland : Windows Greenlandic-Greenland
	LanguageIDWindowsGreenlandicGreenland = LanguageID(0x046F)
	// LanguageIDWindowsGujaratiIndia : Windows Gujarati-India
	LanguageIDWindowsGujaratiIndia = LanguageID(0x0447)
	// LanguageIDWindowsHausaLatinNigeria : Windows Hausa (Latin)-Nigeria
	LanguageIDWindowsHausaLatinNigeria = LanguageID(0x0468)
	// LanguageIDWindowsHebrewIsrael : Windows Hebrew-Israel
	LanguageIDWindowsHebrewIsrael = LanguageID(0x040D)
	// LanguageIDWindowsHindiIndia : Windows Hindi-India
	LanguageIDWindowsHindiIndia = LanguageID(0x0439)
	// LanguageIDWindowsHungarianHungary : Windows Hungarian-Hungary
	LanguageIDWindowsHungarianHungary = LanguageID(0x040E)
	// LanguageIDWindowsIcelandicIceland : Windows Icelandic-Iceland
	LanguageIDWindowsIcelandicIceland = LanguageID(0x040F)
	// LanguageIDWindowsIgboNigeria : Windows Igbo-Nigeria
	LanguageIDWindowsIgboNigeria = LanguageID(0x0470)
	// LanguageIDWindowsIndonesianIndonesia : Windows Indonesian-Indonesia
	LanguageIDWindowsIndonesianIndonesia = LanguageID(0x0421)
	// LanguageIDWindowsInuktitutCanada : Windows Inuktitut-Canada
	LanguageIDWindowsInuktitutCanada = LanguageID(0x045D)
	// LanguageIDWindowsInuktitutLatinCanada : Windows Inuktitut (Latin)-Canada
	LanguageIDWindowsInuktitutLatinCanada = LanguageID(0x085D)
	// LanguageIDWindowsIrishIreland : Windows Irish-Ireland
	LanguageIDWindowsIrishIreland = LanguageID(0x083C)
	// LanguageIDWindowsisiXhosaSouthAfrica : Windows isiXhosa-South Africa
	LanguageIDWindowsisiXhosaSouthAfrica = LanguageID(0x0434)
	// LanguageIDWindowsisiZuluSouthAfrica : Windows isiZulu-South Africa
	LanguageIDWindowsisiZuluSouthAfrica = LanguageID(0x0435)
	// LanguageIDWindowsItalianItaly : Windows Italian-Italy
	LanguageIDWindowsItalianItaly = LanguageID(0x0410)
	// LanguageIDWindowsItalianSwitzerland : Windows Italian-Switzerland
	LanguageIDWindowsItalianSwitzerland = LanguageID(0x0810)
	// LanguageIDWindowsJapaneseJapan : Windows Japanese-Japan
	LanguageIDWindowsJapaneseJapan = LanguageID(0x0411)
	// LanguageIDWindowsKannadaIndia : Windows Kannada-India
	LanguageIDWindowsKannadaIndia = LanguageID(0x044B)
	// LanguageIDWindowsKazakhKazakhstan : Windows Kazakh-Kazakhstan
	LanguageIDWindowsKazakhKazakhstan = LanguageID(0x043F)
	// LanguageIDWindowsKhmerCambodia : Windows Khmer-Cambodia
	LanguageIDWindowsKhmerCambodia = LanguageID(0x0453)
	// LanguageIDWindowsKicheGuatemala : Windows K'iche-Guatemala
	LanguageIDWindowsKicheGuatemala = LanguageID(0x0486)
	// LanguageIDWindowsKinyarwandaRwanda : Windows Kinyarwanda-Rwanda
	LanguageIDWindowsKinyarwandaRwanda = LanguageID(0x0487)
	// LanguageIDWindowsKiswahiliKenya : Windows Kiswahili-Kenya
	LanguageIDWindowsKiswahiliKenya = LanguageID(0x0441)
	// LanguageIDWindowsKonkaniIndia : Windows Konkani-India
	LanguageIDWindowsKonkaniIndia = LanguageID(0x0457)
	// LanguageIDWindowsKoreanKorea : Windows Korean-Korea
	LanguageIDWindowsKoreanKorea = LanguageID(0x0412)
	// LanguageIDWindowsKyrgyzKyrgyzstan : Windows Kyrgyz-Kyrgyzstan
	LanguageIDWindowsKyrgyzKyrgyzstan = LanguageID(0x0440)
	// LanguageIDWindowsLaoLaoPDR : Windows Lao-Lao P.D.R.
	LanguageIDWindowsLaoLaoPDR = LanguageID(0x0454)
	// LanguageIDWindowsLatvianLatvia : Windows Latvian-Latvia
	LanguageIDWindowsLatvianLatvia = LanguageID(0x0426)
	// LanguageIDWindowsLithuanianLithuania : Windows Lithuanian-Lithuania
	LanguageIDWindowsLithuanianLithuania = LanguageID(0x0427)
	// LanguageIDWindowsLowerSorbianGermany : Windows Lower Sorbian-Germany
	LanguageIDWindowsLowerSorbianGermany = LanguageID(0x082E)
	// LanguageIDWindowsLuxembourgishLuxembourg : Windows Luxembourgish-Luxembourg
	LanguageIDWindowsLuxembourgishLuxembourg = LanguageID(0x046E)
	// LanguageIDWindowsMacedonianFYROMFormerYugoslavRepublicofMacedonia : Windows Macedonian (FYROM)-Former Yugoslav Republic of Macedonia
	LanguageIDWindowsMacedonianFYROMFormerYugoslavRepublicofMacedonia = LanguageID(0x042F)
	// LanguageIDWindowsMalayBruneiDarussalam : Windows Malay-Brunei Darussalam
	LanguageIDWindowsMalayBruneiDarussalam = LanguageID(0x083E)
	// LanguageIDWindowsMalayMalaysia : Windows Malay-Malaysia
	LanguageIDWindowsMalayMalaysia = LanguageID(0x043E)
	// LanguageIDWindowsMalayalamIndia : Windows Malayalam-India
	LanguageIDWindowsMalayalamIndia = LanguageID(0x044C)
	// LanguageIDWindowsMalteseMalta : Windows Maltese-Malta
	LanguageIDWindowsMalteseMalta = LanguageID(0x043A)
	// LanguageIDWindowsMaoriNewZealand : Windows Maori-New Zealand
	LanguageIDWindowsMaoriNewZealand = LanguageID(0x0481)
	// LanguageIDWindowsMapudungunChile : Windows Mapudungun-Chile
	LanguageIDWindowsMapudungunChile = LanguageID(0x047A)
	// LanguageIDWindowsMarathiIndia : Windows Marathi-India
	LanguageIDWindowsMarathiIndia = LanguageID(0x044E)
	// LanguageIDWindowsMohawkMohawk : Windows Mohawk-Mohawk
	LanguageIDWindowsMohawkMohawk = LanguageID(0x047C)
	// LanguageIDWindowsMongolianCyrillicMongolia : Windows Mongolian (Cyrillic)-Mongolia
	LanguageIDWindowsMongolianCyrillicMongolia = LanguageID(0x0450)
	// LanguageIDWindowsMongolianTraditionalPeoplesRepublicofChina : Windows Mongolian (Traditional)-People's Republic of China
	LanguageIDWindowsMongolianTraditionalPeoplesRepublicofChina = LanguageID(0x0850)
	// LanguageIDWindowsNepaliNepal : Windows Nepali-Nepal
	LanguageIDWindowsNepaliNepal = LanguageID(0x0461)
	// LanguageIDWindowsNorwegianBokmalNorway : Windows Norwegian (Bokmal)-Norway
	LanguageIDWindowsNorwegianBokmalNorway = LanguageID(0x0414)
	// LanguageIDWindowsNorwegianNynorskNorway : Windows Norwegian (Nynorsk)-Norway
	LanguageIDWindowsNorwegianNynorskNorway = LanguageID(0x0814)
	// LanguageIDWindowsOccitanFrance : Windows Occitan-France
	LanguageIDWindowsOccitanFrance = LanguageID(0x0482)
	// LanguageIDWindowsOdiaformerlyOriyaIndia : Windows Odia (formerly Oriya)-India
	LanguageIDWindowsOdiaformerlyOriyaIndia = LanguageID(0x0448)
	// LanguageIDWindowsPashtoAfghanistan : Windows Pashto-Afghanistan
	LanguageIDWindowsPashtoAfghanistan = LanguageID(0x0463)
	// LanguageIDWindowsPolishPoland : Windows Polish-Poland
	LanguageIDWindowsPolishPoland = LanguageID(0x0415)
	// LanguageIDWindowsPortugueseBrazil : Windows Portuguese-Brazil
	LanguageIDWindowsPortugueseBrazil = LanguageID(0x0416)
	// LanguageIDWindowsPortuguesePortugal : Windows Portuguese-Portugal
	LanguageIDWindowsPortuguesePortugal = LanguageID(0x0816)
	// LanguageIDWindowsPunjabiIndia : Windows Punjabi-India
	LanguageIDWindowsPunjabiIndia = LanguageID(0x0446)
	// LanguageIDWindowsQuechuaBolivia : Windows Quechua-Bolivia
	LanguageIDWindowsQuechuaBolivia = LanguageID(0x046B)
	// LanguageIDWindowsQuechuaEcuador : Windows Quechua-Ecuador
	LanguageIDWindowsQuechuaEcuador = LanguageID(0x086B)
	// LanguageIDWindowsQuechuaPeru : Windows Quechua-Peru
	LanguageIDWindowsQuechuaPeru = LanguageID(0x0C6B)
	// LanguageIDWindowsRomanianRomania : Windows Romanian-Romania
	LanguageIDWindowsRomanianRomania = LanguageID(0x0418)
	// LanguageIDWindowsRomanshSwitzerland : Windows Romansh-Switzerland
	LanguageIDWindowsRomanshSwitzerland = LanguageID(0x0417)
	// LanguageIDWindowsRussianRussia : Windows Russian-Russia
	LanguageIDWindowsRussianRussia = LanguageID(0x0419)
	// LanguageIDWindowsSamiInariFinland : Windows Sami (Inari)-Finland
	LanguageIDWindowsSamiInariFinland = LanguageID(0x243B)
	// LanguageIDWindowsSamiLuleNorway : Windows Sami (Lule)-Norway
	LanguageIDWindowsSamiLuleNorway = LanguageID(0x103B)
	// LanguageIDWindowsSamiLuleSweden : Windows Sami (Lule)-Sweden
	LanguageIDWindowsSamiLuleSweden = LanguageID(0x143B)
	// LanguageIDWindowsSamiNorthernFinland : Windows Sami (Northern)-Finland
	LanguageIDWindowsSamiNorthernFinland = LanguageID(0x0C3B)
	// LanguageIDWindowsSamiNorthernNorway : Windows Sami (Northern)-Norway
	LanguageIDWindowsSamiNorthernNorway = LanguageID(0x043B)
	// LanguageIDWindowsSamiNorthernSweden : Windows Sami (Northern)-Sweden
	LanguageIDWindowsSamiNorthernSweden = LanguageID(0x083B)
	// LanguageIDWindowsSamiSkoltFinland : Windows Sami (Skolt)-Finland
	LanguageIDWindowsSamiSkoltFinland = LanguageID(0x203B)
	// LanguageIDWindowsSamiSouthernNorway : Windows Sami (Southern)-Norway
	LanguageIDWindowsSamiSouthernNorway = LanguageID(0x183B)
	// LanguageIDWindowsSamiSouthernSweden : Windows Sami (Southern)-Sweden
	LanguageIDWindowsSamiSouthernSweden = LanguageID(0x1C3B)
	// LanguageIDWindowsSanskritIndia : Windows Sanskrit-India
	LanguageIDWindowsSanskritIndia = LanguageID(0x044F)
	// LanguageIDWindowsSerbianCyrillicBosniaandHerzegovina : Windows Serbian (Cyrillic)-Bosnia and Herzegovina
	LanguageIDWindowsSerbianCyrillicBosniaandHerzegovina = LanguageID(0x1C1A)
	// LanguageIDWindowsSerbianCyrillicSerbia : Windows Serbian (Cyrillic)-Serbia
	LanguageIDWindowsSerbianCyrillicSerbia = LanguageID(0x0C1A)
	// LanguageIDWindowsSerbianLatinBosniaandHerzegovina : Windows Serbian (Latin)-Bosnia and Herzegovina
	LanguageIDWindowsSerbianLatinBosniaandHerzegovina = LanguageID(0x181A)
	// LanguageIDWindowsSerbianLatinSerbia : Windows Serbian (Latin)-Serbia
	LanguageIDWindowsSerbianLatinSerbia = LanguageID(0x081A)
	// LanguageIDWindowsSesothosaLeboaSouthAfrica : Windows Sesotho sa Leboa-South Africa
	LanguageIDWindowsSesothosaLeboaSouthAfrica = LanguageID(0x046C)
	// LanguageIDWindowsSetswanaSouthAfrica : Windows Setswana-South Africa
	LanguageIDWindowsSetswanaSouthAfrica = LanguageID(0x0432)
	// LanguageIDWindowsSinhalaSriLanka : Windows Sinhala-Sri Lanka
	LanguageIDWindowsSinhalaSriLanka = LanguageID(0x045B)
	// LanguageIDWindowsSlovakSlovakia : Windows Slovak-Slovakia
	LanguageIDWindowsSlovakSlovakia = LanguageID(0x041B)
	// LanguageIDWindowsSlovenianSlovenia : Windows Slovenian-Slovenia
	LanguageIDWindowsSlovenianSlovenia = LanguageID(0x0424)
	// LanguageIDWindowsSpanishArgentina : Windows Spanish-Argentina
	LanguageIDWindowsSpanishArgentina = LanguageID(0x2C0A)
	// LanguageIDWindowsSpanishBolivia : Windows Spanish-Bolivia
	LanguageIDWindowsSpanishBolivia = LanguageID(0x400A)
	// LanguageIDWindowsSpanishChile : Windows Spanish-Chile
	LanguageIDWindowsSpanishChile = LanguageID(0x340A)
	// LanguageIDWindowsSpanishColombia : Windows Spanish-Colombia
	LanguageIDWindowsSpanishColombia = LanguageID(0x240A)
	// LanguageIDWindowsSpanishCostaRica : Windows Spanish-Costa Rica
	LanguageIDWindowsSpanishCostaRica = LanguageID(0x140A)
	// LanguageIDWindowsSpanishDominicanRepublic : Windows Spanish-Dominican Republic
	LanguageIDWindowsSpanishDominicanRepublic = LanguageID(0x1C0A)
	// LanguageIDWindowsSpanishEcuador : Windows Spanish-Ecuador
	LanguageIDWindowsSpanishEcuador = LanguageID(0x300A)
	// LanguageIDWindowsSpanishElSalvador : Windows Spanish-El Salvador
	LanguageIDWindowsSpanishElSalvador = LanguageID(0x440A)
	// LanguageIDWindowsSpanishGuatemala : Windows Spanish-Guatemala
	LanguageIDWindowsSpanishGuatemala = LanguageID(0x100A)
	// LanguageIDWindowsSpanishHonduras : Windows Spanish-Honduras
	LanguageIDWindowsSpanishHonduras = LanguageID(0x480A)
	// LanguageIDWindowsSpanishMexico : Windows Spanish-Mexico
	LanguageIDWindowsSpanishMexico = LanguageID(0x080A)
	// LanguageIDWindowsSpanishNicaragua : Windows Spanish-Nicaragua
	LanguageIDWindowsSpanishNicaragua = LanguageID(0x4C0A)
	// LanguageIDWindowsSpanishPanama : Windows Spanish-Panama
	LanguageIDWindowsSpanishPanama = LanguageID(0x180A)
	// LanguageIDWindowsSpanishParaguay : Windows Spanish-Paraguay
	LanguageIDWindowsSpanishParaguay = LanguageID(0x3C0A)
	// LanguageIDWindowsSpanishPeru : Windows Spanish-Peru
	LanguageIDWindowsSpanishPeru = LanguageID(0x280A)
	// LanguageIDWindowsSpanishPuertoRico : Windows Spanish-Puerto Rico
	LanguageIDWindowsSpanishPuertoRico = LanguageID(0x500A)
	// LanguageIDWindowsSpanishModernSortSpain : Windows Spanish (Modern Sort)-Spain
	LanguageIDWindowsSpanishModernSortSpain = LanguageID(0x0C0A)
	// LanguageIDWindowsSpanishTraditionalSortSpain : Windows Spanish (Traditional Sort)-Spain
	LanguageIDWindowsSpanishTraditionalSortSpain = LanguageID(0x040A)
	// LanguageIDWindowsSpanishUnitedStates : Windows Spanish-United States
	LanguageIDWindowsSpanishUnitedStates = LanguageID(0x540A)
	// LanguageIDWindowsSpanishUruguay : Windows Spanish-Uruguay
	LanguageIDWindowsSpanishUruguay = LanguageID(0x380A)
	// LanguageIDWindowsSpanishVenezuela : Windows Spanish-Venezuela
	LanguageIDWindowsSpanishVenezuela = LanguageID(0x200A)
	// LanguageIDWindowsSwedenFinland : Windows Sweden-Finland
	LanguageIDWindowsSwedenFinland = LanguageID(0x081D)
	// LanguageIDWindowsSwedishSweden : Windows Swedish-Sweden
	LanguageIDWindowsSwedishSweden = LanguageID(0x041D)
	// LanguageIDWindowsSyriacSyria : Windows Syriac-Syria
	LanguageIDWindowsSyriacSyria = LanguageID(0x045A)
	// LanguageIDWindowsTajikCyrillicTajikistan : Windows Tajik (Cyrillic)-Tajikistan
	LanguageIDWindowsTajikCyrillicTajikistan = LanguageID(0x0428)
	// LanguageIDWindowsTamazightLatinAlgeria : Windows Tamazight (Latin)-Algeria
	LanguageIDWindowsTamazightLatinAlgeria = LanguageID(0x085F)
	// LanguageIDWindowsTamilIndia : Windows Tamil-India
	LanguageIDWindowsTamilIndia = LanguageID(0x0449)
	// LanguageIDWindowsTatarRussia : Windows Tatar-Russia
	LanguageIDWindowsTatarRussia = LanguageID(0x0444)
	// LanguageIDWindowsTeluguIndia : Windows Telugu-India
	LanguageIDWindowsTeluguIndia = LanguageID(0x044A)
	// LanguageIDWindowsThaiThailand : Windows Thai-Thailand
	LanguageIDWindowsThaiThailand = LanguageID(0x041E)
	// LanguageIDWindowsTibetanPRC : Windows Tibetan-PRC
	LanguageIDWindowsTibetanPRC = LanguageID(0x0451)
	// LanguageIDWindowsTurkishTurkey : Windows Turkish-Turkey
	LanguageIDWindowsTurkishTurkey = LanguageID(0x041F)
	// LanguageIDWindowsTurkmenTurkmenistan : Windows Turkmen-Turkmenistan
	LanguageIDWindowsTurkmenTurkmenistan = LanguageID(0x0442)
	// LanguageIDWindowsUighurPRC : Windows Uighur-PRC
	LanguageIDWindowsUighurPRC = LanguageID(0x0480)
	// LanguageIDWindowsUkrainianUkraine : Windows Ukrainian-Ukraine
	LanguageIDWindowsUkrainianUkraine = LanguageID(0x0422)
	// LanguageIDWindowsUpperSorbianGermany : Windows Upper Sorbian-Germany
	LanguageIDWindowsUpperSorbianGermany = LanguageID(0x042E)
	// LanguageIDWindowsUrduIslamicRepublicofPakistan : Windows Urdu-Islamic Republic of Pakistan
	LanguageIDWindowsUrduIslamicRepublicofPakistan = LanguageID(0x0420)
	// LanguageIDWindowsUzbekCyrillicUzbekistan : Windows Uzbek (Cyrillic)-Uzbekistan
	LanguageIDWindowsUzbekCyrillicUzbekistan = LanguageID(0x0843)
	// LanguageIDWindowsUzbekLatinUzbekistan : Windows Uzbek (Latin)-Uzbekistan
	LanguageIDWindowsUzbekLatinUzbekistan = LanguageID(0x0443)
	// LanguageIDWindowsVietnameseVietnam : Windows Vietnamese-Vietnam
	LanguageIDWindowsVietnameseVietnam = LanguageID(0x042A)
	// LanguageIDWindowsWelshUnitedKingdom : Windows Welsh-United Kingdom
	LanguageIDWindowsWelshUnitedKingdom = LanguageID(0x0452)
	// LanguageIDWindowsWolofSenegal : Windows Wolof-Senegal
	LanguageIDWindowsWolofSenegal = LanguageID(0x0488)
	// LanguageIDWindowsYakutRussia : Windows Yakut-Russia
	LanguageIDWindowsYakutRussia = LanguageID(0x0485)
	// LanguageIDWindowsYiPRC : Windows Yi-PRC
	LanguageIDWindowsYiPRC = LanguageID(0x0478)
	// LanguageIDWindowsYorubaNigeria : Windows Yoruba-Nigeria
	LanguageIDWindowsYorubaNigeria = LanguageID(0x046A)
	// LanguageIDMacintoshEnglish : Macintosh English
	LanguageIDMacintoshEnglish = LanguageID(0)
	// LanguageIDMacintoshFrench : Macintosh French
	LanguageIDMacintoshFrench = LanguageID(1)
	// LanguageIDMacintoshGerman : Macintosh German
	LanguageIDMacintoshGerman = LanguageID(2)
	// LanguageIDMacintoshItalian : Macintosh Italian
	LanguageIDMacintoshItalian = LanguageID(3)
	// LanguageIDMacintoshDutch : Macintosh Dutch
	LanguageIDMacintoshDutch = LanguageID(4)
	// LanguageIDMacintoshSwedish : Macintosh Swedish
	LanguageIDMacintoshSwedish = LanguageID(5)
	// LanguageIDMacintoshSpanish : Macintosh Spanish
	LanguageIDMacintoshSpanish = LanguageID(6)
	// LanguageIDMacintoshDanish : Macintosh Danish
	LanguageIDMacintoshDanish = LanguageID(7)
	// LanguageIDMacintoshPortuguese : Macintosh Portuguese
	LanguageIDMacintoshPortuguese = LanguageID(8)
	// LanguageIDMacintoshNorwegian : Macintosh Norwegian
	LanguageIDMacintoshNorwegian = LanguageID(9)
	// LanguageIDMacintoshHebrew : Macintosh Hebrew
	LanguageIDMacintoshHebrew = LanguageID(10)
	// LanguageIDMacintoshJapanese : Macintosh Japanese
	LanguageIDMacintoshJapanese = LanguageID(11)
	// LanguageIDMacintoshArabic : Macintosh Arabic
	LanguageIDMacintoshArabic = LanguageID(12)
	// LanguageIDMacintoshFinnish : Macintosh Finnish
	LanguageIDMacintoshFinnish = LanguageID(13)
	// LanguageIDMacintoshGreek : Macintosh Greek
	LanguageIDMacintoshGreek = LanguageID(14)
	// LanguageIDMacintoshIcelandic : Macintosh Icelandic
	LanguageIDMacintoshIcelandic = LanguageID(15)
	// LanguageIDMacintoshMaltese : Macintosh Maltese
	LanguageIDMacintoshMaltese = LanguageID(16)
	// LanguageIDMacintoshTurkish : Macintosh Turkish
	LanguageIDMacintoshTurkish = LanguageID(17)
	// LanguageIDMacintoshCroatian : Macintosh Croatian
	LanguageIDMacintoshCroatian = LanguageID(18)
	// LanguageIDMacintoshChineseTraditional : Macintosh Chinese (Traditional)
	LanguageIDMacintoshChineseTraditional = LanguageID(19)
	// LanguageIDMacintoshUrdu : Macintosh Urdu
	LanguageIDMacintoshUrdu = LanguageID(20)
	// LanguageIDMacintoshHindi : Macintosh Hindi
	LanguageIDMacintoshHindi = LanguageID(21)
	// LanguageIDMacintoshThai : Macintosh Thai
	LanguageIDMacintoshThai = LanguageID(22)
	// LanguageIDMacintoshKorean : Macintosh Korean
	LanguageIDMacintoshKorean = LanguageID(23)
	// LanguageIDMacintoshLithuanian : Macintosh Lithuanian
	LanguageIDMacintoshLithuanian = LanguageID(24)
	// LanguageIDMacintoshPolish : Macintosh Polish
	LanguageIDMacintoshPolish = LanguageID(25)
	// LanguageIDMacintoshHungarian : Macintosh Hungarian
	LanguageIDMacintoshHungarian = LanguageID(26)
	// LanguageIDMacintoshEstonian : Macintosh Estonian
	LanguageIDMacintoshEstonian = LanguageID(27)
	// LanguageIDMacintoshLatvian : Macintosh Latvian
	LanguageIDMacintoshLatvian = LanguageID(28)
	// LanguageIDMacintoshSami : Macintosh Sami
	LanguageIDMacintoshSami = LanguageID(29)
	// LanguageIDMacintoshFaroese : Macintosh Faroese
	LanguageIDMacintoshFaroese = LanguageID(30)
	// LanguageIDMacintoshFarsiPersian : Macintosh Farsi/Persian
	LanguageIDMacintoshFarsiPersian = LanguageID(31)
	// LanguageIDMacintoshRussian : Macintosh Russian
	LanguageIDMacintoshRussian = LanguageID(32)
	// LanguageIDMacintoshChineseSimplified : Macintosh Chinese (Simplified)
	LanguageIDMacintoshChineseSimplified = LanguageID(33)
	// LanguageIDMacintoshFlemish : Macintosh Flemish
	LanguageIDMacintoshFlemish = LanguageID(34)
	// LanguageIDMacintoshIrishGaelic : Macintosh Irish Gaelic
	LanguageIDMacintoshIrishGaelic = LanguageID(35)
	// LanguageIDMacintoshAlbanian : Macintosh Albanian
	LanguageIDMacintoshAlbanian = LanguageID(36)
	// LanguageIDMacintoshRomanian : Macintosh Romanian
	LanguageIDMacintoshRomanian = LanguageID(37)
	// LanguageIDMacintoshCzech : Macintosh Czech
	LanguageIDMacintoshCzech = LanguageID(38)
	// LanguageIDMacintoshSlovak : Macintosh Slovak
	LanguageIDMacintoshSlovak = LanguageID(39)
	// LanguageIDMacintoshSlovenian : Macintosh Slovenian
	LanguageIDMacintoshSlovenian = LanguageID(40)
	// LanguageIDMacintoshYiddish : Macintosh Yiddish
	LanguageIDMacintoshYiddish = LanguageID(41)
	// LanguageIDMacintoshSerbian : Macintosh Serbian
	LanguageIDMacintoshSerbian = LanguageID(42)
	// LanguageIDMacintoshMacedonian : Macintosh Macedonian
	LanguageIDMacintoshMacedonian = LanguageID(43)
	// LanguageIDMacintoshBulgarian : Macintosh Bulgarian
	LanguageIDMacintoshBulgarian = LanguageID(44)
	// LanguageIDMacintoshUkrainian : Macintosh Ukrainian
	LanguageIDMacintoshUkrainian = LanguageID(45)
	// LanguageIDMacintoshByelorussian : Macintosh Byelorussian
	LanguageIDMacintoshByelorussian = LanguageID(46)
	// LanguageIDMacintoshUzbek : Macintosh Uzbek
	LanguageIDMacintoshUzbek = LanguageID(47)
	// LanguageIDMacintoshKazakh : Macintosh Kazakh
	LanguageIDMacintoshKazakh = LanguageID(48)
	// LanguageIDMacintoshAzerbaijaniCyrillicscript : Macintosh Azerbaijani (Cyrillic script)
	LanguageIDMacintoshAzerbaijaniCyrillicscript = LanguageID(49)
	// LanguageIDMacintoshAzerbaijaniArabicscript : Macintosh Azerbaijani (Arabic script)
	LanguageIDMacintoshAzerbaijaniArabicscript = LanguageID(50)
	// LanguageIDMacintoshArmenian : Macintosh Armenian
	LanguageIDMacintoshArmenian = LanguageID(51)
	// LanguageIDMacintoshGeorgian : Macintosh Georgian
	LanguageIDMacintoshGeorgian = LanguageID(52)
	// LanguageIDMacintoshMoldavian : Macintosh Moldavian
	LanguageIDMacintoshMoldavian = LanguageID(53)
	// LanguageIDMacintoshKirghiz : Macintosh Kirghiz
	LanguageIDMacintoshKirghiz = LanguageID(54)
	// LanguageIDMacintoshTajiki : Macintosh Tajiki
	LanguageIDMacintoshTajiki = LanguageID(55)
	// LanguageIDMacintoshTurkmen : Macintosh Turkmen
	LanguageIDMacintoshTurkmen = LanguageID(56)
	// LanguageIDMacintoshMongolianMongolianscript : Macintosh Mongolian (Mongolian script)
	LanguageIDMacintoshMongolianMongolianscript = LanguageID(57)
	// LanguageIDMacintoshMongolianCyrillicscript : Macintosh Mongolian (Cyrillic script)
	LanguageIDMacintoshMongolianCyrillicscript = LanguageID(58)
	// LanguageIDMacintoshPashto : Macintosh Pashto
	LanguageIDMacintoshPashto = LanguageID(59)
	// LanguageIDMacintoshKurdish : Macintosh Kurdish
	LanguageIDMacintoshKurdish = LanguageID(60)
	// LanguageIDMacintoshKashmiri : Macintosh Kashmiri
	LanguageIDMacintoshKashmiri = LanguageID(61)
	// LanguageIDMacintoshSindhi : Macintosh Sindhi
	LanguageIDMacintoshSindhi = LanguageID(62)
	// LanguageIDMacintoshTibetan : Macintosh Tibetan
	LanguageIDMacintoshTibetan = LanguageID(63)
	// LanguageIDMacintoshNepali : Macintosh Nepali
	LanguageIDMacintoshNepali = LanguageID(64)
	// LanguageIDMacintoshSanskrit : Macintosh Sanskrit
	LanguageIDMacintoshSanskrit = LanguageID(65)
	// LanguageIDMacintoshMarathi : Macintosh Marathi
	LanguageIDMacintoshMarathi = LanguageID(66)
	// LanguageIDMacintoshBengali : Macintosh Bengali
	LanguageIDMacintoshBengali = LanguageID(67)
	// LanguageIDMacintoshAssamese : Macintosh Assamese
	LanguageIDMacintoshAssamese = LanguageID(68)
	// LanguageIDMacintoshGujarati : Macintosh Gujarati
	LanguageIDMacintoshGujarati = LanguageID(69)
	// LanguageIDMacintoshPunjabi : Macintosh Punjabi
	LanguageIDMacintoshPunjabi = LanguageID(70)
	// LanguageIDMacintoshOriya : Macintosh Oriya
	LanguageIDMacintoshOriya = LanguageID(71)
	// LanguageIDMacintoshMalayalam : Macintosh Malayalam
	LanguageIDMacintoshMalayalam = LanguageID(72)
	// LanguageIDMacintoshKannada : Macintosh Kannada
	LanguageIDMacintoshKannada = LanguageID(73)
	// LanguageIDMacintoshTamil : Macintosh Tamil
	LanguageIDMacintoshTamil = LanguageID(74)
	// LanguageIDMacintoshTelugu : Macintosh Telugu
	LanguageIDMacintoshTelugu = LanguageID(75)
	// LanguageIDMacintoshSinhalese : Macintosh Sinhalese
	LanguageIDMacintoshSinhalese = LanguageID(76)
	// LanguageIDMacintoshBurmese : Macintosh Burmese
	LanguageIDMacintoshBurmese = LanguageID(77)
	// LanguageIDMacintoshKhmer : Macintosh Khmer
	LanguageIDMacintoshKhmer = LanguageID(78)
	// LanguageIDMacintoshLao : Macintosh Lao
	LanguageIDMacintoshLao = LanguageID(79)
	// LanguageIDMacintoshVietnamese : Macintosh Vietnamese
	LanguageIDMacintoshVietnamese = LanguageID(80)
	// LanguageIDMacintoshIndonesian : Macintosh Indonesian
	LanguageIDMacintoshIndonesian = LanguageID(81)
	// LanguageIDMacintoshTagalong : Macintosh Tagalong
	LanguageIDMacintoshTagalong = LanguageID(82)
	// LanguageIDMacintoshMalayRomanscript : Macintosh Malay (Roman script)
	LanguageIDMacintoshMalayRomanscript = LanguageID(83)
	// LanguageIDMacintoshMalayArabicscript : Macintosh Malay (Arabic script)
	LanguageIDMacintoshMalayArabicscript = LanguageID(84)
	// LanguageIDMacintoshAmharic : Macintosh Amharic
	LanguageIDMacintoshAmharic = LanguageID(85)
	// LanguageIDMacintoshTigrinya : Macintosh Tigrinya
	LanguageIDMacintoshTigrinya = LanguageID(86)
	// LanguageIDMacintoshGalla : Macintosh Galla
	LanguageIDMacintoshGalla = LanguageID(87)
	// LanguageIDMacintoshSomali : Macintosh Somali
	LanguageIDMacintoshSomali = LanguageID(88)
	// LanguageIDMacintoshSwahili : Macintosh Swahili
	LanguageIDMacintoshSwahili = LanguageID(89)
	// LanguageIDMacintoshKinyarwandaRuanda : Macintosh Kinyarwanda/Ruanda
	LanguageIDMacintoshKinyarwandaRuanda = LanguageID(90)
	// LanguageIDMacintoshRundi : Macintosh Rundi
	LanguageIDMacintoshRundi = LanguageID(91)
	// LanguageIDMacintoshNyanjaChewa : Macintosh Nyanja/Chewa
	LanguageIDMacintoshNyanjaChewa = LanguageID(92)
	// LanguageIDMacintoshMalagasy : Macintosh Malagasy
	LanguageIDMacintoshMalagasy = LanguageID(93)
	// LanguageIDMacintoshEsperanto : Macintosh Esperanto
	LanguageIDMacintoshEsperanto = LanguageID(94)
	// LanguageIDMacintoshWelsh : Macintosh Welsh
	LanguageIDMacintoshWelsh = LanguageID(128)
	// LanguageIDMacintoshBasque : Macintosh Basque
	LanguageIDMacintoshBasque = LanguageID(129)
	// LanguageIDMacintoshCatalan : Macintosh Catalan
	LanguageIDMacintoshCatalan = LanguageID(130)
	// LanguageIDMacintoshLatin : Macintosh Latin
	LanguageIDMacintoshLatin = LanguageID(131)
	// LanguageIDMacintoshQuenchua : Macintosh Quenchua
	LanguageIDMacintoshQuenchua = LanguageID(132)
	// LanguageIDMacintoshGuarani : Macintosh Guarani
	LanguageIDMacintoshGuarani = LanguageID(133)
	// LanguageIDMacintoshAymara : Macintosh Aymara
	LanguageIDMacintoshAymara = LanguageID(134)
	// LanguageIDMacintoshTatar : Macintosh Tatar
	LanguageIDMacintoshTatar = LanguageID(135)
	// LanguageIDMacintoshUighur : Macintosh Uighur
	LanguageIDMacintoshUighur = LanguageID(136)
	// LanguageIDMacintoshDzongkha : Macintosh Dzongkha
	LanguageIDMacintoshDzongkha = LanguageID(137)
	// LanguageIDMacintoshJavaneseRomanscript : Macintosh Javanese (Roman script)
	LanguageIDMacintoshJavaneseRomanscript = LanguageID(138)
	// LanguageIDMacintoshSundaneseRomanscript : Macintosh Sundanese (Roman script)
	LanguageIDMacintoshSundaneseRomanscript = LanguageID(139)
	// LanguageIDMacintoshGalician : Macintosh Galician
	LanguageIDMacintoshGalician = LanguageID(140)
	// LanguageIDMacintoshAfrikaans : Macintosh Afrikaans
	LanguageIDMacintoshAfrikaans = LanguageID(141)
	// LanguageIDMacintoshBreton : Macintosh Breton
	LanguageIDMacintoshBreton = LanguageID(142)
	// LanguageIDMacintoshInuktitut : Macintosh Inuktitut
	LanguageIDMacintoshInuktitut = LanguageID(143)
	// LanguageIDMacintoshScottishGaelic : Macintosh Scottish Gaelic
	LanguageIDMacintoshScottishGaelic = LanguageID(144)
	// LanguageIDMacintoshManxGaelic : Macintosh Manx Gaelic
	LanguageIDMacintoshManxGaelic = LanguageID(145)
	// LanguageIDMacintoshIrishGaelicwithdotabove : Macintosh Irish Gaelic (with dot above)
	LanguageIDMacintoshIrishGaelicwithdotabove = LanguageID(146)
	// LanguageIDMacintoshTongan : Macintosh Tongan
	LanguageIDMacintoshTongan = LanguageID(147)
	// LanguageIDMacintoshGreekpolytonic : Macintosh Greek (polytonic)
	LanguageIDMacintoshGreekpolytonic = LanguageID(148)
	// LanguageIDMacintoshGreenlandic : Macintosh Greenlandic
	LanguageIDMacintoshGreenlandic = LanguageID(149)
	// LanguageIDMacintoshAzerbaijaniRomanscript : Macintosh Azerbaijani (Roman script)
	LanguageIDMacintoshAzerbaijaniRomanscript = LanguageID(150)
	// languageIDUnicode : There are no platform-specific language IDs defined for the Unicode platform.
	languageIDUnicode = "NotDefined"
	// There are no ISO-specific language IDs, and language-tag records are not supported on this platform.
	languageIDISO = "NotDefined"
	// unknown
	unknownLanguageID = "Unknown"
)

// NameID is a metadata name of a OpenType.
type NameID uint16

// Name returns the name of Name ID.
func (n NameID) Name() string {
	if n > 25 {
		if n > 255 {
			return nameIDreservedForFontSpecific + strconv.Itoa(int(n))
		}
		return nameIDreservedForFutureStandard + strconv.Itoa(int(n))
	}
	switch n {
	case NameIDCopyrightNotice:
		return "Copyright notice"
	case NameIDFontFamilyName:
		return "Font Family name"
	case NameIDFontSubfamilyName:
		return "Font Subfamily name"
	case NameIDUniqueFontIdentifier:
		return "Unique font identifier"
	case NameIDFontFullName:
		return "Full font name"
	case NameIDVersion:
		return "Version"
	case NameIDPostScriptName:
		return "PostScript name"
	case NameIDTrademark:
		return "Trademark"
	case NameIDManufacturerName:
		return "Manufacturer name"
	case NameIDDesignerName:
		return "Designer name"
	case NameIDDescription:
		return "description of the typeface"
	case NameIDURLVendor:
		return "URL of font vendor"
	case NameIDURLDesigner:
		return "URL of typeface designer"
	case NameIDLicenseDescription:
		return "description of how the font may be legally used"
	case NameIDLicenseInfoURL:
		return "URL where additional licensing information can be found"
	case NameIDReserved:
		return "Reserved"
	case NameIDTypographicFamilyName:
		return "Typographic Family name"
	case NameIDTypographicSubfamilyName:
		return "Typographic Subfamily name"
	case NameIDCompatibleFull:
		return "Compatible Full"
	case NameIDSampleText:
		return "Sample text"
	case NameIDPostScriptCIDFindfontName:
		return "PostScript CID findfont name"
	case NameIDWWSFamilyName:
		return "WWS Family Name"
	case NameIDWWSSubfamilyName:
		return "WWS Subfamily Name"
	case NameIDLightBackgroundPalette:
		return "Light Background Palette"
	case NameIDDarkBackgroundPalette:
		return "Dark Background Palette"
	case NameIDVariationsPostScriptNamePrefix:
		return "Variations PostScript Name"
	}
	return ""
}

func (n NameID) String() string {
	return fmt.Sprintf("(%d):%s", n, n.Name())
}

const (
	// NameIDCopyrightNotice : Copyright notice.
	NameIDCopyrightNotice = NameID(0)
	// NameIDFontFamilyName : Font Family name. This family name is assumed to be shared among fonts that differ only in weight or style (italic, oblique).
	NameIDFontFamilyName = NameID(1)
	// NameIDFontSubfamilyName : Font Subfamily name. The Font Subfamily name distinguishes the fonts in a group with the same Font Family name (name ID 1).
	NameIDFontSubfamilyName = NameID(2)
	// NameIDUniqueFontIdentifier : Unique font identifier
	NameIDUniqueFontIdentifier = NameID(3)
	// NameIDFontFullName : Full font name that reflects all family and relevant subfamily descriptors.
	NameIDFontFullName = NameID(4)
	// NameIDVersion : Version string. Should begin with the syntax 'Version <number>.<number>' (upper case, lower case, or mixed, with a space between “Version” and the number).
	NameIDVersion = NameID(5)
	// NameIDPostScriptName : PostScript name for the font; Name ID 6 specifies a string which is used to invoke a PostScript language font that corresponds to this OpenType font.
	NameIDPostScriptName = NameID(6)
	// NameIDTrademark : Trademark; this is used to save any trademark notice/information for this font.
	NameIDTrademark = NameID(7)
	// NameIDManufacturerName : Manufacturer Name.
	NameIDManufacturerName = NameID(8)
	// NameIDDesignerName : name of the designer of the typeface.
	NameIDDesignerName = NameID(9)
	// NameIDDescription : description of the typeface. Can contain revision information, usage recommendations, history, features, etc.
	NameIDDescription = NameID(10)
	// NameIDURLVendor : URL of font vendor (with protocol, e.g., http://, ftp://). If a unique serial number is embedded in the URL, it can be used to register the font.
	NameIDURLVendor = NameID(11)
	// NameIDURLDesigner : URL of typeface designer (with protocol, e.g., http://, ftp://).
	NameIDURLDesigner = NameID(12)
	// NameIDLicenseDescription : description of how the font may be legally used, or different example scenarios for licensed use. This field should be written in plain language, not legalese.
	NameIDLicenseDescription = NameID(13)
	// NameIDLicenseInfoURL : URL where additional licensing information can be found.
	NameIDLicenseInfoURL = NameID(14)
	// NameIDReserved : Reserved.
	NameIDReserved = NameID(15)
	// NameIDTypographicFamilyName : The typographic family grouping doesn't impose any constraints on the number of faces within it, in contrast with the 4-style family grouping (ID 1), which is present both for historical reasons and to express style linking groups. If name ID 16 is absent, then name ID 1 is considered to be the typographic family name.
	NameIDTypographicFamilyName = NameID(16)
	// NameIDTypographicSubfamilyName : This allows font designers to specify a subfamily name within the typographic family grouping. This string must be unique within a particular typographic family. If it is absent, then name ID 2 is considered to be the typographic subfamily name.
	NameIDTypographicSubfamilyName = NameID(17)
	// NameIDCompatibleFull : On the Macintosh, the menu name is constructed using the FOND resource. This usually matches the Full Name.
	NameIDCompatibleFull = NameID(18)
	// NameIDSampleText : This can be the font name, or any other text that the designer thinks is the best sample to display the font in.
	NameIDSampleText = NameID(19)
	// NameIDPostScriptCIDFindfontName : Its presence in a font means that the nameID 6 holds a PostScript font name that is meant to be used with the “composefont” invocation in order to invoke the font in a PostScript interpreter.
	NameIDPostScriptCIDFindfontName = NameID(20)
	// NameIDWWSFamilyName : Used to provide a WWS-conformant family name in case the entries for IDs 16 and 17 do not conform to the WWS model.
	NameIDWWSFamilyName = NameID(21)
	// NameIDWWSSubfamilyName : Used in conjunction with ID 21, this ID provides a WWS-conformant subfamily name (reflecting only weight, width and slope attributes) in case the entries for IDs 16 and 17 do not conform to the WWS model.
	NameIDWWSSubfamilyName = NameID(22)
	// NameIDLightBackgroundPalette : if used in the CPAL table’s Palette Labels Array, specifies that the corresponding color palette in the CPAL table is appropriate to use with the font when displaying it on a light background such as white.
	NameIDLightBackgroundPalette = NameID(23)
	// NameIDDarkBackgroundPalette : if used in the CPAL table’s Palette Labels Array, specifies that the corresponding color palette in the CPAL table is appropriate to use with the font when displaying it on a dark background such as black.
	NameIDDarkBackgroundPalette = NameID(24)
	// NameIDVariationsPostScriptNamePrefix : If present in a variable font, it may be used as the family prefix in the PostScript Name Generation for Variation Fonts algorithm.
	NameIDVariationsPostScriptNamePrefix = NameID(25)
	// Name IDs 26 to 255, inclusive, are reserved for future standard names.
	nameIDreservedForFutureStandard = "reserved for future standard names"
	// Name IDs 256 to 32767, inclusive, are reserved for font-specific names such as those referenced by a font's layout features.
	nameIDreservedForFontSpecific = "reserved for font-specific names"
)
