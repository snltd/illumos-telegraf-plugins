package main

// Serializes all Named kstats under the given path, to disk. Useful for generating fixture data
// to mock out tests which require kstats.
// The data is of type []*kstat.Named, the filename is the kstat name with `.kstat` appended.

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/illumos/go-kstat"
)

func main() {
	if len(os.Args) != 2 { //nolint
		fmt.Fprintln(os.Stderr, "usage: capture-kstat <kstat>")
		os.Exit(1)
	}

	kstatName := os.Args[1]
	file := fmt.Sprintf("%s.kstat", strings.ReplaceAll(kstatName, ":", "--"))
	chunks := strings.Split(kstatName, ":")

	if len(chunks) != 3 { //nolint
		fmt.Fprintln(os.Stderr, "kstat must be of the form module:instance:name")
		os.Exit(1)
	}

	module := chunks[0]
	instance, _ := strconv.Atoi(chunks[1])
	name := chunks[2]

	token, err := kstat.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot get kstat token.")
	}

	rawKstat, err := token.Lookup(module, instance, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot get kstat.")
		os.Exit(1)
	}

	stats, err := rawKstat.AllNamed()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get named kstat data.")
		os.Exit(1)
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(stats)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not encode data %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(file, buf.Bytes(), 0o644) //nolint

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not write serialized data to disk: %v\n", err)
		os.Exit(1)
	}
}
