package main

// Grabs the psinfo and usage for the given proc and serialises them to disk, for use as
// fixtures in testing the illumos_process Telegraf collector

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
)

func getPsinfo(pid string) psinfoT {
	var psinfo psinfoT

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

func getUsage(pid string) prusageT {
	file := fmt.Sprintf("/proc/%s/usage", pid)

	var prusage prusageT

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
		psinfoFile := fmt.Sprintf("%s.psinfo", pid)

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

		usageFile := fmt.Sprintf("%s.usage", pid)
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
