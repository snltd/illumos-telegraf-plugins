package main

// Serializes all kstats to disk. Useful for generating fixture data to mock out tests which
// require kstats.

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/illumos/go-kstat"
)

func main() {
	token, err := kstat.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot get kstat token.")
	}

	stats := token.All()
	fmt.Printf("%T\n", stats)

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(stats)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not encode data %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("all.kstat", buf.Bytes(), 0o644)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not write serialized data to disk: %v\n", err)
		os.Exit(1)
	}
}
