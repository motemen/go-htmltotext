package htmltotext

import (
	"io"
	"strings"
	"unicode"
)

// TODO: squeeze newlines

type squeezingWriter struct {
	writer   io.Writer
	nextChar byte
	atStart  bool
}

func newSqueezingWriter(w io.Writer) *squeezingWriter {
	return &squeezingWriter{
		writer:  w,
		atStart: true,
	}
}

func (w *squeezingWriter) Write(p []byte) (int, error) {
	s, leading, trailing := trimSpaces(string(p))
	if s == "" {
		return 0, nil
	}
	if w.nextChar != '\000' {
		w.writer.Write([]byte{w.nextChar})
	} else if leading && !w.atStart {
		w.writer.Write([]byte{' '})
	}
	if trailing {
		w.nextChar = ' '
	} else {
		w.nextChar = '\000'
	}
	w.atStart = false
	return w.writer.Write([]byte(s))
}

func (w *squeezingWriter) WriteSpace() {
	w.nextChar = ' '
	w.atStart = false
}

func (w *squeezingWriter) WriteNewLine() {
	w.nextChar = '\n'
	w.atStart = true
}

func trimSpace(trimFunc func(string, func(rune) bool) string, s string) (result string, trimmed bool) {
	result = trimFunc(s, unicode.IsSpace)
	trimmed = len(result) != len(s)
	return
}

func trimSpaces(s string) (string, bool, bool) {
	s, leading := trimSpace(strings.TrimLeftFunc, s)
	s, trailing := trimSpace(strings.TrimRightFunc, s)
	return s, leading, trailing
}
