package opentype

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type errIO struct {
	err error
}

func (e errIO) hasErr() bool {
	return e.err != nil
}

func (e errIO) errorf(format string) error {
	if e.hasErr() {
		return fmt.Errorf(format, e.err)
	}
	return nil
}

type errWriter struct {
	errIO
	w io.Writer
}

func newErrWriter(w io.Writer) *errWriter {
	return &errWriter{
		errIO: errIO{
			err: nil,
		},
		w: w,
	}
}

func (e *errWriter) write(data interface{}) {
	if e.hasErr() {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			e.err = fmt.Errorf("panic: %s", err)
		}
	}()
	e.err = binary.Write(e.w, binary.BigEndian, data)
}

func (e *errWriter) writeBin(b []byte) {
	if e.hasErr() {
		return
	}
	_, e.err = e.w.Write(b)
}

type errReader struct {
	errIO
	r io.Reader
}

func newErrReader(r io.Reader) *errReader {
	return &errReader{
		errIO: errIO{
			err: nil,
		},
		r: r,
	}
}

func (e *errReader) read(data interface{}) {
	if e.hasErr() {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			e.err = fmt.Errorf("panic: %s", err)
		}
	}()
	e.err = binary.Read(e.r, binary.BigEndian, data)
}

type optionalFontParser struct {
	tableRecords map[string]*TableRecord
	errs         []string
}

func newOptionalFontParser(tableRecords map[string]*TableRecord) *optionalFontParser {
	return &optionalFontParser{
		tableRecords: tableRecords,
		errs:         make([]string, 0),
	}
}

func (p *optionalFontParser) parse(tag string, optional bool, parser func(tr *TableRecord) error) {
	tr, ok := p.tableRecords[tag]
	if !ok {
		if !optional {
			p.errs = append(p.errs, fmt.Sprintf("%s: table record is not found", tag))
		}
		return
	}
	err := parser(tr)
	if err != nil {
		p.errs = append(p.errs, fmt.Sprintf("%s: %s", tag, err))
	}
}

func (p *optionalFontParser) err() error {
	if len(p.errs) == 0 {
		return nil
	}
	return fmt.Errorf("parsing OpenType font failed: [%s]", strings.Join(p.errs, ", "))
}

func tableRequired(target ...Table) error {
	missed := make([]string, 0)
	for _, t := range target {
		if t == nil {
			missed = append(missed, t.Tag().String())
		}
	}
	if len(missed) > 0 {
		return fmt.Errorf("requires %s", strings.Join(missed, ","))
	}
	return nil
}
