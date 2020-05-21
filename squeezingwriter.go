package htmltotext

import (
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
)

type squeezingWriter struct {
	writer io.Writer
	state  state
}

type state int

const (
	stateStartOfDocument state = iota
	stateNone
	stateSpace
	stateNewline
	stateParagraph
	stateStart
)

func newSqueezingWriter(w io.Writer) *squeezingWriter {
	return &squeezingWriter{
		writer: w,
		state:  stateStartOfDocument,
	}
}

func (w *squeezingWriter) writeNonspace(p []byte) (int, error) {
	var nextState state
	var lead string

	switch w.state {
	case stateNone:
		lead, nextState = "", stateNone
	case stateSpace:
		lead, nextState = " ", stateNone
	case stateNewline:
		lead, nextState = "\n", stateNone
	case stateParagraph:
		lead, nextState = "\n\n", stateNone
	case stateStart:
		lead, nextState = "", stateNone
	case stateStartOfDocument:
		lead, nextState = "", stateNone
	}

	w.state = nextState

	return fmt.Fprintf(w.writer, "%s%s", lead, p)
}

func (w *squeezingWriter) writeSpace() (int, error) {
	switch w.state {
	case stateNone:
		w.state = stateSpace
	case stateSpace:
	case stateNewline:
	case stateParagraph:
	case stateStart:
	case stateStartOfDocument:
	}

	return 0, nil
}

func (w *squeezingWriter) Write(p []byte) (int, error) {
	s, leading, trailing := trimSpaces(string(p))
	if leading {
		w.writeSpace()
	}
	if s == "" {
		return 0, nil
	}
	n, err := w.writeNonspace([]byte(s))
	if trailing {
		w.writeSpace()
	}
	return n, err
}

func (w *squeezingWriter) InsertSpace() {
	switch w.state {
	case stateNone:
		w.state = stateSpace
	case stateSpace:
	case stateNewline:
		w.writeNonspace([]byte{'\n'})
		w.state = stateStart
	case stateParagraph:
	case stateStart:
	case stateStartOfDocument:
	}
}

func (w *squeezingWriter) InsertNewline() {
	switch w.state {
	case stateNone:
		w.state = stateNewline
	case stateSpace:
		w.state = stateNewline
	case stateNewline:
	case stateStart:
	case stateParagraph:
	}
}

func (w *squeezingWriter) InsertParagraph() {
	switch w.state {
	case stateNone:
		w.state = stateParagraph
	case stateSpace:
		w.state = stateParagraph
	case stateNewline:
		w.state = stateParagraph
	case stateParagraph:
		w.state = stateParagraph
	case stateStart:
		w.state = stateParagraph
	case stateStartOfDocument:
	}
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

func debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

type squeezingWriterQueue struct {
	sync  *squeezingWriter
	queue chan func() error
}

func (q *squeezingWriterQueue) Write(p []byte) (int, error) {
	s := make([]byte, len(p))
	copy(s, p)

	q.queue <- func() error {
		_, err := q.sync.Write(s)
		return err
	}
	return 0, nil
}

func (q *squeezingWriterQueue) InsertSpace() {
	q.queue <- func() error {
		q.sync.InsertSpace()
		return nil
	}
}

func (q *squeezingWriterQueue) InsertNewline() {
	q.queue <- func() error {
		q.sync.InsertNewline()
		return nil
	}
}

func (q *squeezingWriterQueue) InsertParagraph() {
	q.queue <- func() error {
		q.sync.InsertParagraph()
		return nil
	}
}
