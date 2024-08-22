package helpers

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmdWorks(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := RunCmd("/bin/echo something")
	require.Equal(t, "something", stdout)
	require.Equal(t, "", stderr)
	require.Nil(t, err)
}

func TestRunCmdNoSuchCmd(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := RunCmd("/bin/no such thing")
	require.Equal(t, "", stdout)
	require.Equal(t, "", stderr)
	require.Error(t, err)
	require.IsType(t, &os.PathError{}, err)
}

// You definitely ought not to be able to do this. If you can, check yourself!
func TestRunCmdDisallowed(t *testing.T) {
	t.Parallel()

	if os.Geteuid() > 0 {
		stdout, stderr, err := RunCmd("/bin/mkdir /directory")
		require.Equal(t, "", stdout)
		require.Contains(t, stderr, "Permission denied")
		require.Error(t, err)
		require.IsType(t, &exec.ExitError{}, err)
	}
}

// The effort involved in properly testing this is too high. Just make sure it fails,
// and if you're running it in a global zone, as root, with a zone called 'no-such-zone', then
// you're just out of luck.
func TestRunCmdZlogin(t *testing.T) {
	t.Parallel()

	// This can raise different errors depending on the zone it's run in, and possibly on
	// privileges, so let's just assert an error.
	stdout, stderr, err := RunCmdInZone("/bin/pfexec", "/bin/date", "no-such-zone")
	require.Equal(t, "", stdout)
	require.NotEqual(t, "", stderr)
	require.Error(t, err)
	require.IsType(t, &exec.ExitError{}, err)
}
