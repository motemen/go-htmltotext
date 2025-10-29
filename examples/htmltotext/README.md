# HTML to Text Converter Example

This example demonstrates how to use the go-htmltotext library to convert HTML to plain text.

## Usage

### From stdin:
```bash
echo '<h1>Hello</h1><p>World</p>' | go run main.go
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
go build -o htmltotext
./htmltotext sample.html
```
