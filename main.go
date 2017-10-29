package main

import (
	"fmt"
	"os"

	"github.com/taknuki/go-opentype/opentype"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go-opentype fontfile")
		os.Exit(1)
	}
	fileName := os.Args[1]
	err := cmdMain(fileName)
	if err != nil {
		os.Exit(1)
	}
}

func cmdMain(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()
	ttc, err := opentype.IsFontCollection(f)
	if err != nil {
		return
	}
	if ttc {
		dumpFontCollection(f)
	} else {
		dumpFont(f)
	}
	return nil
}

func dumpFontCollection(f *os.File) {
	fc, err := opentype.ParseFontCollections(f)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, font := range fc.Fonts {
			dumpCMap(font)
		}
	}
}

func dumpFont(f *os.File) {
	font, err := opentype.ParseFont(f)
	if err != nil {
		fmt.Println(err)
	} else {
		dumpCMap(font)
	}
}

func dumpCMap(font *opentype.Font) {
	cm := font.CMap
	for _, er := range cm.EncodingRecords {
		fmt.Printf("platform: %s encoding: %s format: %d\n", er.PlatformID, er.EncodingID.String(er.PlatformID), er.Subtable.GetFormatNumber())
		for i := int32(32); i <= 300; i++ {
			if val, ok := er.GetCMap()[i]; ok {
				fmt.Printf("char:%s gid:%d\n", string(rune(i)), val)
			}
		}
	}
}
