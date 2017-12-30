package opentype

import (
	"encoding/binary"
	"io"
)

// bWrite is a shorthand of binary.Write
func bWrite(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}
