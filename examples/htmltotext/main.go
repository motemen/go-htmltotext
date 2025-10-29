package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"strings"

	htmltotext "github.com/motemen/go-htmltotext"
	htmlParser "golang.org/x/net/html"
)

// readUntilEndTag reads tokens until the matching end tag is found and returns the text content
func readUntilEndTag(z *htmlParser.Tokenizer, tagName string) string {
	var content strings.Builder
	depth := 1

	for {
		tt := z.Next()
		switch tt {
		case htmlParser.ErrorToken:
			return content.String()
		case htmlParser.TextToken:
			content.WriteString(html.UnescapeString(string(z.Text())))
		case htmlParser.StartTagToken, htmlParser.SelfClosingTagToken:
			token := z.Token()
			if token.Data == tagName {
				depth++
			}
		case htmlParser.EndTagToken:
			token := z.Token()
			if token.Data == tagName {
				depth--
				if depth == 0 {
					return content.String()
				}
			}
		}
	}
}

// headingHandler creates a handler for heading tags (h1-h6)
func headingHandler(level int) htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		z := ctx.Value(htmltotext.ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)
		content := readUntilEndTag(z, token.Data)

		prefix := strings.Repeat("#", level)
		fmt.Fprintf(w, "%s %s", prefix, content)
		errc <- nil
	}
}

// listItemHandler creates a handler for list items
func listItemHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		w.Write([]byte("- "))
		errc <- nil
	}
}

// strongHandler creates a handler for bold text
func strongHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		z := ctx.Value(htmltotext.ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)
		content := readUntilEndTag(z, token.Data)

		fmt.Fprintf(w, "**%s**", content)
		errc <- nil
	}
}

// emHandler creates a handler for italic text
func emHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		z := ctx.Value(htmltotext.ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)
		content := readUntilEndTag(z, token.Data)

		fmt.Fprintf(w, "*%s*", content)
		errc <- nil
	}
}

// linkHandler creates a handler for links
func linkHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		z := ctx.Value(htmltotext.ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)

		// Get href attribute
		var href string
		for _, attr := range token.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}

		content := readUntilEndTag(z, "a")

		if href != "" {
			fmt.Fprintf(w, "[%s](%s)", content, href)
		} else {
			w.Write([]byte(content))
		}
		errc <- nil
	}
}

// codeHandler creates a handler for inline code
func codeHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		z := ctx.Value(htmltotext.ContextKeyExperimentalTokenizer).(*htmlParser.Tokenizer)
		content := readUntilEndTag(z, "code")

		fmt.Fprintf(w, "`%s`", content)
		errc <- nil
	}
}

// blockquoteHandler creates a handler for blockquotes
func blockquoteHandler() htmltotext.Handler {
	return func(ctx context.Context, token htmlParser.Token, w io.Writer, errc chan error) {
		w.Write([]byte("> "))
		errc <- nil
	}
}

func main() {
	var reader io.Reader

	// If argument is provided, read from file; otherwise read from stdin
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer file.Close()
		reader = file
	} else {
		reader = os.Stdin
	}

	// Create converter with Markdown handlers
	conf := htmltotext.New(
		// Heading handlers
		htmltotext.WithHandler("h1", headingHandler(1)),
		htmltotext.WithHandler("h2", headingHandler(2)),
		htmltotext.WithHandler("h3", headingHandler(3)),
		htmltotext.WithHandler("h4", headingHandler(4)),
		htmltotext.WithHandler("h5", headingHandler(5)),
		htmltotext.WithHandler("h6", headingHandler(6)),

		// List item handler
		htmltotext.WithHandler("li", listItemHandler()),

		// Text formatting handlers
		htmltotext.WithHandler("strong", strongHandler()),
		htmltotext.WithHandler("b", strongHandler()),
		htmltotext.WithHandler("em", emHandler()),
		htmltotext.WithHandler("i", emHandler()),

		// Link handler
		htmltotext.WithHandler("a", linkHandler()),

		// Code handler
		htmltotext.WithHandler("code", codeHandler()),

		// Blockquote handler
		htmltotext.WithHandler("blockquote", blockquoteHandler()),
	)

	// Convert HTML to Markdown
	ctx := context.Background()
	err := conf.Convert(ctx, reader, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout)
}
