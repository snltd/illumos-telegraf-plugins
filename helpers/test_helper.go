package helpers

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/illumos/go-kstat"
)

// FromFixture loads serialized kstat data off disk and returns the real data. The filename is
// relative to testdata/.
func FromFixture(filename string) []*kstat.Named {
	var kstatData []*kstat.Named

	filename = filepath.Join("testdata", filename)
	raw, err := os.Open(filename)

	if err != nil {
		log.Fatal(fmt.Sprintf("Could not load serialized data from disk: %v\n", err))
	}

	dec := gob.NewDecoder(raw)
	err = dec.Decode(&kstatData)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load decode kstat data: %v\n", err)
		os.Exit(1)
	}

	return kstatData
}
