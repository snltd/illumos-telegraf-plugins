package smbserver

import (
	// "fmt"
	"testing"

	// "github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

// My dev zone doesn't have any smbsrv kstats, so this is as much testing as I can do.
func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosSmbServer{
		Fields: []string{"connections", "open_files"},
	}

	testData := helpers.FromFixture("smbsrv--0--smbsrv.kstat")
	fields := parseNamedStats(s, testData)

	require.Equal(
		t,
		fields,
		map[string]interface{}{
			"connections": float64(1),
			"open_files":  float64(0),
		},
	)
}
