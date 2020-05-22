package htmltotext

import (
	"testing"

	"bytes"
	"fmt"

	"github.com/google/go-cmp/cmp"
)

func TestSpaceSqueezingWriter(t *testing.T) {
	var buf bytes.Buffer
	w := newSqueezingWriter(&buf)

	assertBuf := func(expected string) {
		t.Helper()
		if diff := cmp.Diff(expected, buf.String()); diff != "" {
			t.Errorf("(-expected +got):\n%s", diff)
		}
	}

	fmt.Fprint(w, "  hoge  ")
	assertBuf("hoge")

	fmt.Fprint(w, "  fuga  ")
	assertBuf("hoge fuga")

	fmt.Fprint(w, "1")
	assertBuf("hoge fuga 1")

	fmt.Fprint(w, " 2 ")
	assertBuf("hoge fuga 1 2")

	w.InsertSpace()
	assertBuf("hoge fuga 1 2")

	fmt.Fprint(w, "")
	assertBuf("hoge fuga 1 2")

	fmt.Fprint(w, "3")
	assertBuf("hoge fuga 1 2 3")

	w.InsertSpace()
	assertBuf("hoge fuga 1 2 3")

	fmt.Fprint(w, "4")
	assertBuf("hoge fuga 1 2 3 4")

	w.InsertNewline()
	assertBuf("hoge fuga 1 2 3 4")

	w.writeNonspace(nil) // flush
	assertBuf("hoge fuga 1 2 3 4\n")

	w.InsertNewline()
	assertBuf("hoge fuga 1 2 3 4\n")

	w.InsertSpace()
	assertBuf("hoge fuga 1 2 3 4\n\n")

	w.Write([]byte{'a'})
	assertBuf("hoge fuga 1 2 3 4\n\na")

	// InsertParagraph on stateStart
	w.InsertNewline()
	w.InsertSpace()
	if w.state != stateStart {
		t.Fatalf("expected w.state == stateStart; got %v", w.state)
	}
	w.InsertParagraph()
	w.Write([]byte{'x'})
	assertBuf("hoge fuga 1 2 3 4\n\na\n\n\nx")
}

func TestSpaceSqueezingWriter_2(t *testing.T) {
	var buf bytes.Buffer
	w := newSqueezingWriter(&buf)

	fmt.Fprint(w, "1")
	w.InsertNewline()

	fmt.Fprint(w, "2")
	w.InsertNewline()

	fmt.Fprint(w, "3")
	w.InsertNewline()

	if diff := cmp.Diff("1\n2\n3", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
}
