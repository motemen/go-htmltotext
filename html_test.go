package htmltotext

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"testing"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	client := &http.Client{
		Transport: http.NewFileTransport(http.Dir(".")),
	}

	conf := New(
		WithFramesSupport(),
		WithHTTPClient(client),
	)
	files, err := filepath.Glob("testdata/*.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Base(file)[:1] == "_" {
			continue
		}
		t.Run(file, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyURL, &url.URL{
				Scheme: "file",
				Path:   file,
			})

			var buf bytes.Buffer
			err = conf.Convert(ctx, f, &buf)
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
