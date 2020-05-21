package htmltotext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"

	htmlParser "golang.org/x/net/html"
)

type tagKind int

const (
	tagKindNormal tagKind = iota
	tagKindSkip
	tagKindSingleBlock
	tagKindBlock
	tagKindInlineBlock
	tagKindParagraph
)

var tagConfig = map[string]tagKind{
	"head":   tagKindSkip,
	"script": tagKindSkip,
	"style":  tagKindSkip,
	"title":  tagKindSkip,

	"br": tagKindSingleBlock,
	"hr": tagKindSingleBlock,

	"div":     tagKindBlock,
	"tr":      tagKindBlock,
	"dt":      tagKindBlock,
	"dd":      tagKindBlock,
	"option":  tagKindBlock,
	"section": tagKindBlock,
	"main":    tagKindBlock,
	"header":  tagKindBlock,
	"nav":     tagKindBlock,
	"article": tagKindBlock,
	"li":      tagKindBlock,

	"p":          tagKindParagraph,
	"pre":        tagKindParagraph,
	"blockquote": tagKindParagraph,
	"h1":         tagKindParagraph,
	"h2":         tagKindParagraph,
	"h3":         tagKindParagraph,
	"h4":         tagKindParagraph,
	"h5":         tagKindParagraph,
	"h6":         tagKindParagraph,
	"ul":         tagKindParagraph,
	"ol":         tagKindParagraph,
	"table":      tagKindParagraph,

	"input": tagKindInlineBlock,
	"th":    tagKindInlineBlock,
	"td":    tagKindInlineBlock,
}

var ErrSkipTag = errors.New("htmltotext: skip this tag")

func New(opts ...Option) *Config {
	config := &Config{
		handlers: DefaultTagHandlers,
		maxDepth: DefaultMaxDepth,
	}
	for _, o := range opts {
		o(config)
	}
	return config
}

type Handler func(context.Context, htmlParser.Token, io.Writer, chan error)

type Config struct {
	handlers   map[string]Handler
	maxDepth   int
	httpClient *http.Client
}

var DefaultMaxDepth = 2

var DefaultTagHandlers = map[string]Handler{
	"img": func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		for _, attr := range token.Attr {
			if attr.Key == "alt" {
				w.Write([]byte(attr.Val))
				break
			}
		}

		errc <- nil
	},
}

var FrameRecurseHandler = func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
	conf := FromContext(ctx)
	u := ctx.Value(ContextKeyURL).(*url.URL)
	for _, attr := range token.Attr {
		if attr.Key == "src" {
			go func() {
				rel, err := url.Parse(attr.Val)
				if err != nil {
					errc <- err
					return
				}

				rel = u.ResolveReference(rel)

				httpClient := conf.httpClient
				if httpClient == nil {
					httpClient = http.DefaultClient
				}
				resp, err := httpClient.Get(rel.String())
				if err != nil {
					errc <- err
					return
				}

				defer resp.Body.Close()
				ctx = context.WithValue(ctx, ContextKeyURL, rel)
				errc <- conf.Convert(ctx, resp.Body, w)
			}()
			return
		}
	}

	errc <- nil
}

var NoframesSkipHandler = func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
	errc <- ErrSkipTag
}

type Option func(*Config)

func WithHandler(tag string, h Handler) Option {
	return func(conf *Config) {
		conf.handlers[tag] = h
	}
}

func WithMaxDepth(depth int) Option {
	return func(conf *Config) {
		conf.maxDepth = depth
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(conf *Config) {
		conf.httpClient = httpClient
	}
}

func WithFramesSupport() Option {
	return func(conf *Config) {
		WithHandler("frame", FrameRecurseHandler)(conf)
		WithHandler("noframes", NoframesSkipHandler)(conf)
	}
}

type ContextKey struct{ string }

var (
	ContextKeyConfig = ContextKey{"config"}
	ContextKeyDepth  = ContextKey{"depth"}
	ContextKeyURL    = ContextKey{"url"}
)

func FromContext(ctx context.Context) *Config {
	return ctx.Value(ContextKeyConfig).(*Config)
}

func depthFromContext(ctx context.Context) int {
	depth, _ := ctx.Value(ContextKeyDepth).(int)
	return depth
}

func (conf *Config) Convert(ctx context.Context, r io.Reader, w io.Writer) error {
	if depthFromContext(ctx) > conf.maxDepth && conf.maxDepth != -1 {
		return nil
	}

	ctx = context.WithValue(ctx, ContextKeyConfig, conf)

	var buf bytes.Buffer

	sw := &squeezingWriterQueue{
		sync:  newSqueezingWriter(&buf),
		queue: make(chan func() error, 5),
	}

	done := make(chan struct{})
	go func() {
		for f := range sw.queue {
			_ = f()
		}
		done <- struct{}{}
	}()

	z := htmlParser.NewTokenizer(r)
	var skip bool
parseHTML:
	for {
		tt := z.Next()
		switch tt {
		case htmlParser.ErrorToken:
			if z.Err() == io.EOF {
				break parseHTML
			}
			return fmt.Errorf("parsing html: %w", z.Err())

		case htmlParser.TextToken:
			if !skip {
				io.WriteString(sw, html.UnescapeString(string(z.Text())))
			}

		case htmlParser.StartTagToken, htmlParser.SelfClosingTagToken:
			token := z.Token()
			kind := tagConfig[token.Data]

			skip = kind == tagKindSkip // TODO: aria-hidden
			switch kind {
			case tagKindSingleBlock:
				sw.InsertNewline()
			case tagKindParagraph:
				sw.InsertParagraph()
			case tagKindBlock:
				sw.InsertNewline()
			}

			if handler, ok := conf.handlers[token.Data]; ok {
				errc := make(chan error, 1)

				var buf bytes.Buffer
				ctx := context.WithValue(ctx, ContextKeyDepth, depthFromContext(ctx)+1)
				handler(ctx, token, &buf, errc)

				select {
				case err := <-errc:
					if err == ErrSkipTag {
						skip = true
					} else if err != nil {
						return err
					} else {
						sw.Write(buf.Bytes())
					}

				default:
					sw.queue <- func() error {
						err := <-errc
						if err != nil {
							return err
						}

						// FIXME make selectable by handler's return value
						sw.sync.InsertParagraph()
						sw.sync.Write(buf.Bytes())
						return nil
					}
				}
			} else {
				if token.Data == "noscript" || token.Data == "noframes" {
					z.NextIsNotRawText()
				}
			}

		case htmlParser.EndTagToken:
			tn, _ := z.TagName()
			kind := tagConfig[string(tn)]

			skip = false
			switch kind {
			case tagKindBlock:
				sw.InsertNewline()
			case tagKindParagraph:
				sw.InsertParagraph()
			case tagKindInlineBlock:
				sw.InsertSpace()
			}
		}
	}

	close(sw.queue)
	<-done

	_, err := io.Copy(w, &buf)
	return err
}
