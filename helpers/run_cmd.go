package helpers

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// RunCmd runs a command, returning output as a string, and IGNORING RETURN CODE AND STDERR!
// Commands are specified as a simple string.  You can't do pipes and stuff because of the way Go
// forks.
var RunCmd = func(cmd string) (string, string, error) {
	chunks := strings.SplitN(cmd, " ", 2)

	var executor *exec.Cmd
	var stdout, stderr bytes.Buffer

	if len(chunks) == 2 { //nolint:gomnd
		executor = exec.Command(chunks[0], strings.Split(chunks[1], " ")...) //nolint:gosec
	} else {
		executor = exec.Command(chunks[0]) //nolint:gosec
	}

	executor.Stdout = &stdout
	executor.Stderr = &stderr

	err := executor.Run()

	outString := strings.TrimSpace(stdout.String())
	errString := strings.TrimSpace(stderr.String())

	if err != nil {
		log.Printf("error running %s: %s\n", cmd, err)

		return outString, errString, err
	}

	return outString, errString, nil
}

// RunCmdPfexec runs a command via pfexec(1), returning its output as a string. Same caveats as
// RunCmd().
func RunCmdPfexec(cmd string) (string, string, error) {
	stdout, stderr, err := RunCmd(fmt.Sprintf("/bin/pfexec %s", cmd))

	return stdout, stderr, err
}

func RunCmdInZone(zone, cmd string) (string, string, error) {
	stdout, stderr, err := RunCmd(fmt.Sprintf("/bin/pfexec /usr/sbin/zlogin %s \"%s\"", zone, cmd))

	return stdout, stderr, err
}
