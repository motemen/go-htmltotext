package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	_ "github.com/motemen/go-loghttp/global"
	"github.com/motemen/go-nuts/httputil"

	htmltotext "github.com/motemen/go-htmltotext"
)

func main() {
	u, _ := url.Parse(os.Args[1])

	client := &http.Client{
		Transport: &httputil.ChardetTransport{},
	}

	conf := htmltotext.New(
		htmltotext.WithFramesSupport(nil),
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
