package htmltotext

import (
	"bytes"
	"fmt"
	"html"
	"io"

	htmlParser "golang.org/x/net/html"
)

var tagSkip = map[string]bool{
	"head":   true,
	"script": true,
	"style":  true,
	"title":  true,
}

var tagSingleBlock = map[string]bool{
	"br": true,
	"hr": true,
}

var tagBlock = map[string]bool{
	"div":    true,
	"tr":     true,
	"dt":     true,
	"dd":     true,
	"option": true,
}

var tagParagraph = map[string]bool{
	"p":          true,
	"pre":        true,
	"blockquote": true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"ul":         true,
	"ol":         true,
}

var tagInlineBlock = map[string]bool{
	// "button": true,
	"input": true,
	"td":    true,
	"li":    true,
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
			skip = tagSkip[string(tn)] // TODO: aria-hidden
			if tagSingleBlock[string(tn)] {
				w.InsertNewline()
			} else if tagParagraph[string(tn)] {
				w.InsertParagraph()
			} else if tagBlock[string(tn)] {
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
			skip = tagSkip[string(tn)]
			if tagBlock[string(tn)] {
				w.InsertNewline()
			} else if tagParagraph[string(tn)] {
				w.InsertParagraph()
			} else if tagInlineBlock[string(tn)] {
				w.InsertNewline()
			}
		}
	}

	return buf.Bytes(), nil
}
