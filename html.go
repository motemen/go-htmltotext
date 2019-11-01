package htmltotext

import (
	"bytes"
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

	"input": tagKindInlineBlock,
	"td":    tagKindInlineBlock,
	"li":    tagKindInlineBlock,
}

func HTMLToText(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	w := newSqueezingWriter(&buf)

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
			return nil, fmt.Errorf("parsing html: %w", z.Err())

		case htmlParser.TextToken:
			if !skip {
				io.WriteString(w, html.UnescapeString(string(z.Text())))
			}

		case htmlParser.StartTagToken, htmlParser.SelfClosingTagToken:
			tn, _ := z.TagName()
			kind := tagConfig[string(tn)]

			skip = kind == tagKindSkip // TODO: aria-hidden
			switch kind {
			case tagKindSingleBlock:
				w.InsertNewline()
			case tagKindParagraph:
				w.InsertParagraph()
			case tagKindBlock:
				w.InsertNewline()
			}
			if string(tn) == "img" {
				for {
					k, v, more := z.TagAttr()
					if string(k) == "alt" {
						w.Write(v)
					}
					if !more {
						break
					}
				}
			}

		case htmlParser.EndTagToken:
			tn, _ := z.TagName()
			kind := tagConfig[string(tn)]

			skip = kind == tagKindSkip
			switch kind {
			case tagKindBlock:
				w.InsertNewline()
			case tagKindParagraph:
				w.InsertParagraph()
			case tagKindInlineBlock:
				w.InsertNewline()
			}
		}
	}

	return buf.Bytes(), nil
}
