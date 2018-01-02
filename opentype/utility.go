package opentype

import (
	"encoding/binary"
	"io"
)

type errWriter struct {
	w   io.Writer
	err error
}

func newErrWriter(w io.Writer) *errWriter {
	return &errWriter{
		w:   w,
		err: nil,
	}
}

func (e *errWriter) write(data interface{}) {
	if e.hasErr() {
		return
	}
	e.err = binary.Write(e.w, binary.BigEndian, data)
}

func (e *errWriter) writeBin(b []byte) {
	if e.hasErr() {
		return
	}
	_, e.err = e.w.Write(b)
}

func (e *errWriter) hasErr() bool {
	return e.err != nil
}
