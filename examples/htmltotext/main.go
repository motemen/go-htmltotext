package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	htmltotext "github.com/motemen/go-htmltotext"
)

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

	// Create converter with default settings
	conf := htmltotext.New()

	// Convert HTML to text
	ctx := context.Background()
	err := conf.Convert(ctx, reader, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout)
}
