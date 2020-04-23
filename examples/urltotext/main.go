// +build demo

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	_ "github.com/motemen/go-loghttp/global"

	htmltotext "github.com/motemen/go-htmltotext"
)

func getContent(u *url.URL) (io.Reader, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	return charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
}

func main() {
	u, _ := url.Parse(os.Args[1])
	r, _ := getContent(u)

	var conf htmltotext.Config
	conf.TagHandlers = htmltotext.DefaultTagHandlers
	conf.TagHandlers["frame"] = func(ctx context.Context, token html.Token, w io.Writer) error {
		for _, attr := range token.Attr {
			if attr.Key == "src" {
				rel, _ := url.Parse(attr.Val)
				rel = u.ResolveReference(rel)
				r, _ := getContent(rel)
				return conf.Convert(ctx, r, w)
			}
		}

		return nil
	}
	conf.TagHandlers["noframes"] = func(ctx context.Context, token html.Token, w io.Writer) error {
		return htmltotext.ErrSkipTag
	}

	var buf bytes.Buffer
	err := conf.Convert(context.Background(), r, &buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
}
