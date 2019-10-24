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
	stateNone state = iota
	stateSpace
	stateNewline
	stateParagraph
	stateStart
)

func newSqueezingWriter(w io.Writer) *squeezingWriter {
	return &squeezingWriter{
		writer: w,
		state:  stateStart,
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
		//w.state = stateStart
		//return w.writer.Write([]byte{'\n'})
	case stateParagraph:
	case stateStart:
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
	case stateStart:
	case stateParagraph:
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
	case stateStart:
		w.state = stateParagraph
	case stateParagraph:
		w.state = stateParagraph
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
