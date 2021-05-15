package helpers

import (
	"os/exec"
	"strings"
)

// RunCmd runs a command, returning output as a string, and IGNORING RETURN CODE AND STDERR!
// Commands are specified as a simple string.  You can't do pipes and stuff because of the way Go
// forks.
var RunCmd = func(cmd string) string {
	chunks := strings.SplitN(cmd, " ", 2)

	var c *exec.Cmd

	if len(chunks) == 2 { //nolint
		c = exec.Command(chunks[0], strings.Split(chunks[1], " ")...)
	} else {
		c = exec.Command(chunks[0])
	}

	p, _ := c.CombinedOutput()

	return strings.TrimSpace(string(p))
}

// RunCmdPfexec runs a command via pfexec(1), returning its output as a string. Same caveats as
// RunCmd().
func RunCmdPfexec(cmd string) string {
	return RunCmd("pfexec " + cmd)
}
