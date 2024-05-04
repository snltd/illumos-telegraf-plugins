package main

// Grabs the psinfo and usage for the given proc and serialises them to disk, for use as
// fixtures in testing the illumos_process Telegraf collector

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"path"
)

func getPsinfo(pid string) psinfo_t {
	var psinfo psinfo_t

	file := fmt.Sprintf("/proc/%s/psinfo", pid)
	fh, err := os.Open(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open psinfo for %s", pid)
		os.Exit(1)
	}

	err = binary.Read(fh, binary.LittleEndian, &psinfo)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not capture psinfo")
		os.Exit(1)
	}

	return psinfo
}

func getUsage(pid string) prusage_t {
	file := fmt.Sprintf("/proc/%s/usage", pid)

	var prusage prusage_t

	fh, err := os.Open(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open usage")
		os.Exit(1)
	}

	err = binary.Read(fh, binary.LittleEndian, &prusage)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not capture usage")
		os.Exit(1)
	}

	return prusage
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: capture_proc <pid>...")
		os.Exit(1)
	}

	for _, pid := range os.Args[1:] {
		psinfoDir := fmt.Sprintf("%s", pid)

		if err := os.Mkdir(psinfoDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Could not create directory: %s", err)
			os.Exit(2)
		}

		psinfoFile := path.Join(psinfoDir, "psinfo")

		var psinfoBuf bytes.Buffer

		psinfo := getPsinfo(pid)
		enc := gob.NewEncoder(&psinfoBuf)
		err := enc.Encode(psinfo)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not encode psinfo data %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(psinfoFile, psinfoBuf.Bytes(), 0o644) //nolint

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write serialized psinfo data to disk: %v\n", err)
			os.Exit(1)
		}

		var usageBuf bytes.Buffer

		usageFile := path.Join(psinfoDir, "usage")
		usage := getUsage(pid)
		enc = gob.NewEncoder(&usageBuf)
		err = enc.Encode(usage)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not encode usage data %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(usageFile, usageBuf.Bytes(), 0o644) //nolint

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write serialized usage data to disk: %v\n", err)
			os.Exit(1)
		}
	}
}
