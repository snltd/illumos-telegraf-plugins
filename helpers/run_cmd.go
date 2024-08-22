package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
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

	return outString, errString, err
}

// RunCmdInZone runs a command via zlogin, unless the given zone is the current zone, in which
// case it calls RunCmd. Quoting the command breaks it.
func RunCmdInZone(prefix string, cmd string, zone ZoneName) (string, string, error) {
	if zone != CurrentZone() {
		cmd = fmt.Sprintf("%s /usr/sbin/zlogin %s %s", prefix, zone, cmd)
	}

	return RunCmd(cmd)
}

func HaveFileInZone(file string, zone ZoneName) bool {
	if zone != CurrentZone() {
		file = path.Join("/zones", string(zone), "root", file)
	}

	_, err := os.Stat(file)

	return err == nil
}
