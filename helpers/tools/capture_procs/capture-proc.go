package main

// Grabs the psinfo and usage for the given proc and serialises them to disk, for use as
// fixtures in testing the illumos_proc Telegraf collector

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

func getPsinfo(pid string) psinfoT {
	var psinfo psinfoT

	file := fmt.Sprintf("/proc/%s/psinfo", pid)
	fh, err := os.Open(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open psinfo")
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
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: capture-proc <pid>")
		os.Exit(1)
	}

	pid := os.Args[1]
	psinfoFile := fmt.Sprintf("%s.psinfo", pid)

	var psinfoBuf bytes.Buffer

	psinfo := getPsinfo(pid)
	enc := gob.NewEncoder(&psinfoBuf)
	err := enc.Encode(psinfo)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not encode psinfo data %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(psinfoFile, psinfoBuf.Bytes(), 0o644) //nolint

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

	err = ioutil.WriteFile(usageFile, usageBuf.Bytes(), 0o644) //nolint

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not write serialized usage data to disk: %v\n", err)
		os.Exit(1)
	}
}

// The following types come from /usr/include/sys/procfs.h, with thanks to
// https://github.com/mitchellh/go-ps/blob/master/process_solaris.go for getting me started

type ushortT uint16
type idT int32
type pidT int32
type uidT int32
type gidT int32
type devT uint64
type sizeT uint64
type uintptrT uint64
type ulongT uint64
type timestrucT [2]int64

type prusageT struct {
	PrLwpid    idT           // lwp id.  0: process or defunct
	PrCount    int32         // number of contributing lwps
	PrTstamp   timestrucT    // current time stamp
	PrCreate   timestrucT    // process/lwp creation time stamp
	PrTerm     timestrucT    // process/lwp termination time stamp
	PrRtime    timestrucT    // total lwp real (elapsed) time
	PrUtime    timestrucT    // user level cpu time
	PrStime    timestrucT    // system call cpu time
	PrTtime    timestrucT    // other system trap cpu time
	PrTftime   timestrucT    // text page fault sleep time
	PrDftime   timestrucT    // data page fault sleep time
	PrKftime   timestrucT    // kernel page fault sleep time
	PrLtime    timestrucT    // user lock wait sleep time
	PrSlptime  timestrucT    // all other sleep time
	PrWtime    timestrucT    // wait-cpu (latency) time
	PrStoptime timestrucT    // stopped time
	Filltime   [6]timestrucT // filler for future expansion
	PrMinf     ulongT        // minor page faults
	PrMajf     ulongT        // major page faults
	PrNswap    ulongT        // swaps
	PrInblk    ulongT        // input blocks
	PrOublk    ulongT        // output blocks
	PrMsnd     ulongT        // messages sent
	PrMrcv     ulongT        // messages received
	PrSigs     ulongT        // signals received
	PrVctx     ulongT        // voluntary context switches
	PrIctx     ulongT        // involuntary context switches
	PrSysc     ulongT        // system calls
	PrIoch     ulongT        // chars read and written
	Filler     [10]ulongT    // filler for future expansion
}

type psinfoT struct {
	PrFlag     int32    // process flags (DEPRECATED; do not use)
	PrNlwp     int32    // number of active lwps in the process
	PrPid      pidT     // unique process id
	PrPpid     pidT     // process id of parent
	PrPgid     pidT     // pid of process group leader
	PrSid      pidT     // session id
	PrUID      uidT     // real user id
	PrEuid     uidT     // effective user id
	PrGID      gidT     // real group id
	PrEgid     gidT     // effective group id
	PrAddr     uintptrT // address of process
	PrSize     sizeT    // size of process image in Kbytes
	PrRssize   sizeT    // resident set size in Kbytes
	PrPad1     sizeT
	PrTtydev   devT    // controlling tty device (or PRNODEV)
	PrPctcpu   ushortT // % of recent cpu time used by all lwps
	PrPctmem   ushortT // % of system memory used by process
	PrPad64bit [4]byte
	PrStart    timestrucT // process start time, from the epoch
	PrTime     timestrucT // usr+sys cpu time for this process
	PrCtime    timestrucT // usr+sys cpu time for reaped children
	PrFname    [16]byte   // name of execed file
	PrPsargs   [80]byte   // initial characters of arg list
	PrWstat    int32      // if zombie, the wait() status
	PrArgc     int32      // initial argument count
	PrArgv     uintptrT   // address of initial argument vector
	PrEnvp     uintptrT   // address of initial environment vector
	PrDmodel   [1]byte    // data model of the process
	PrPad2     [3]byte
	PrTaskid   idT       // task id
	PrProjid   idT       // project id
	PrNzomb    int32     // number of zombie lwps in the process
	PrPoolid   idT       // pool id
	PrZoneid   idT       // zone id
	PrContract idT       // process contract
	PrFiller   int32     // reserved for future use
	PrLwp      [128]byte // information for representative lwp
}
