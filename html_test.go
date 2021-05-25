package htmltotext

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	htmlParser "golang.org/x/net/html"
)

func TestConfig_Convert(t *testing.T) {
	client := &http.Client{
		Transport: http.NewFileTransport(http.Dir(".")),
	}

	conf := New(
		WithFramesSupport(nil),
		WithHTTPClient(client),
		WithMaxDepth(1),
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

func TestConfig_Convert_Handler(t *testing.T) {
	var title string
	meta := map[string]string{}
	conf := New(
		WithHandler("meta", func(ctx context.Context, token htmlParser.Token, w io.Writer, cerr chan error) {
			var k, v string
			for _, a := range token.Attr {
				if a.Key == "name" {
					k = a.Val
				} else if a.Key == "content" {
					v = a.Val
				}
			}
			meta[k] = v
			cerr <- nil
		}),
		WithHandler("title", func(ctx context.Context, token htmlParser.Token, w io.Writer, cerr chan error) {
			r := ctx.Value(ContextKeyExperimentalReader).(io.Reader)
			z := ctx.Value(ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)
			r = io.MultiReader(bytes.NewReader(z.Buffered()), r)
			zz := htmlParser.NewTokenizerFragment(r, token.Data)
			zz.Next()
			title = zz.Token().String()
			cerr <- zz.Err()
		}),
	)
	r := strings.NewReader(`<html>
		<head>
			<title>My Site</title>
			<meta name="robots" content="noindex">
			<meta name="author" content="motemen">
		</head>
		<body>
		Hello
		</body>
	</html>`)

	var buf bytes.Buffer
	err := conf.Convert(context.Background(), r, &buf)
	if err != nil {
		t.Error(err)
	}
	if got, expected := title, "My Site"; got != expected {
		t.Errorf("%q != %q", got, expected)
	}
	if diff := cmp.Diff(map[string]string{"robots": "noindex", "author": "motemen"}, meta); diff != "" {
		t.Errorf("got diff:\n%s", diff)
	}
	if got, expected := buf.String(), "Hello"; got != expected {
		t.Errorf("%q != %q", got, expected)
	}
}
