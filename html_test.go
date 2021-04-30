package htmltotext

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	htmlParser "golang.org/x/net/html"
)

func TestConvert(t *testing.T) {
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

func TestConvert_Handler(t *testing.T) {
	conf := New(
		WithHandler("meta", func(ctx context.Context, token htmlParser.Token, w io.Writer, cerr chan error) {
			t.Log(token)
			cerr <- nil
		}),
		WithHandler("title", func(ctx context.Context, token htmlParser.Token, w io.Writer, cerr chan error) {
			r := ctx.Value("r").(io.Reader)
			z := ctx.Value("z").(*htmlParser.Tokenizer)
			r = io.MultiReader(bytes.NewReader(z.Buffered()), r)
			zz := htmlParser.NewTokenizerFragment(r, token.Data)
			zz.Next()
			t.Log(zz.Token(), zz.Err())
			t.Log(token)
			cerr <- nil
		}),
	)
	r := strings.NewReader(`<html>
		<head>
			<title>Title</title>
			<meta name="robots" content="noindex">
		</head>
		<body>
		Hello
		</body>
	</html>`)
	var buf bytes.Buffer
	conf.Convert(context.Background(), r, &buf)
}
