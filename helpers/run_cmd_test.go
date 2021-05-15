package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Parallel()

	echoOutput := RunCmd("/bin/echo something")
	require.Equal(t, "something", echoOutput)

	nosuch := RunCmd("/bin/no_such_command")
	require.Equal(t, "", nosuch)
}
