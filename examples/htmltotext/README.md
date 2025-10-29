# HTML to Markdown Converter Example

This example demonstrates how to use the go-htmltotext library with custom handlers to convert HTML to Markdown format.

## Features

This example implements custom handlers for various HTML tags to output Markdown:

- **Headings** (`h1`-`h6`) → `#`, `##`, `###`, etc.
- **List items** (`li`) → `- `
- **Bold** (`strong`, `b`) → `**text**`
- **Italic** (`em`, `i`) → `*text*`
- **Links** (`a`) → `[text](url)`
- **Code** (`code`) → `` `text` ``
- **Blockquotes** (`blockquote`) → `> `

## Usage

### From stdin:
```bash
echo '<h1>Hello</h1><p>This is <strong>bold</strong> text.</p>' | go run main.go
```

Output:
```
# Hello

This is **bold** text.
```

### From file:
```bash
go run main.go sample.html
```

### Using pipe:
```bash
curl https://example.com | go run main.go
```

## Build and Install

```bash
go build -o htmltomarkdown
./htmltomarkdown sample.html
```

## How It Works

The example uses `htmltotext.WithHandler()` to register custom handlers for each HTML tag. Each handler:

1. Reads the tag's content using the tokenizer
2. Formats it according to Markdown syntax
3. Writes the formatted output

This demonstrates the flexibility of the go-htmltotext library's handler mechanism.
