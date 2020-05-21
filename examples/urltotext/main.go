package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/html/charset"

	_ "github.com/motemen/go-loghttp/global"

	htmltotext "github.com/motemen/go-htmltotext"
)

type CharsetTransport struct {
	Base http.RoundTripper
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (t *CharsetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return resp, err
	}

	resp.Body = &readCloser{
		Reader: r,
		Closer: resp.Body,
	}
	return resp, nil
}

func main() {
	u, _ := url.Parse(os.Args[1])

	client := &http.Client{
		Transport: &CharsetTransport{},
	}

	conf := htmltotext.New(
		htmltotext.WithFramesSupport(),
		htmltotext.WithHTTPClient(client),
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, htmltotext.ContextKeyURL, u)

	resp, _ := client.Get(u.String())

	err := conf.Convert(ctx, resp.Body, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
