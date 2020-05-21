package htmltotext

import (
	"bytes"
	"context"
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
		t.Run(file, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			err = New().Convert(context.Background(), f, &buf)
			if err != nil {
				t.Fatal(err)
			}

			text := buf.String()

			want, err := ioutil.ReadFile(file + ".txt")
			if err != nil {
				t.Log(string(text))
				t.Fatal(err)
			}

			if diff := cmp.Diff(string(want), string(text)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", filepath.Base(file), diff)
			}
		})
	}
}
