package htmltotext

import (
	"testing"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
)

func TestParseHTMLToText(t *testing.T) {
	files, err := filepath.Glob("testdata/*.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		want, err := ioutil.ReadFile(file + ".txt")
		if err != nil {
			t.Fatal(err)
		}

		text, err := HTMLToText(f)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(string(want), string(text)); diff != "" {
			t.Errorf("%s (-want +got):\n%s", filepath.Base(file), diff)
		}
	}
}
