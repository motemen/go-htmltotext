package htmltotext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"

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

type contextKey struct {
	name string
}

var ContextKeyPageURL = &contextKey{"url"}

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

type TagHandler func(context.Context, htmlParser.Token, io.Writer) error

type Config struct {
	TagHandlers map[string]TagHandler
}

var DefaultTagHandlers = map[string]TagHandler{
	"img": func(ctx context.Context, token htmlParser.Token, w io.Writer) error {
		for _, attr := range token.Attr {
			if attr.Key == "alt" {
				w.Write([]byte(attr.Val))
				break
			}
		}

		return nil
	},
}

func (conf Config) Convert(ctx context.Context, r io.Reader, w io.Writer) error {
	if conf.TagHandlers == nil {
		conf.TagHandlers = DefaultTagHandlers
	}

	sw := newSqueezingWriter(w)

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

			if handler, ok := conf.TagHandlers[token.Data]; ok {
				var buf bytes.Buffer
				err := handler(ctx, token, &buf)
				if err == ErrSkipTag {
					skip = true
				} else if err != nil {
					return err
				}

				io.WriteString(sw, buf.String())
			}

		case htmlParser.EndTagToken:
			tn, _ := z.TagName()
			kind := tagConfig[string(tn)]

			skip = kind == tagKindSkip
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

	return nil
}
