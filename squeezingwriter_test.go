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

	fmt.Fprint(w, "  hoge  ")
	if diff := cmp.Diff("hoge", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}

	fmt.Fprint(w, "  fuga  ")
	if diff := cmp.Diff("hoge fuga", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}

	fmt.Fprint(w, "1")
	if diff := cmp.Diff("hoge fuga 1", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	fmt.Fprint(w, " 2 ")
	if diff := cmp.Diff("hoge fuga 1 2", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	w.WriteSpace()
	if diff := cmp.Diff("hoge fuga 1 2", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	fmt.Fprint(w, "")
	if diff := cmp.Diff("hoge fuga 1 2", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	fmt.Fprint(w, "3")
	if diff := cmp.Diff("hoge fuga 1 2 3", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	w.WriteSpace()
	if diff := cmp.Diff("hoge fuga 1 2 3", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
	fmt.Fprint(w, "4")
	if diff := cmp.Diff("hoge fuga 1 2 3 4", buf.String()); diff != "" {
		t.Errorf("%s", diff)
	}
}
