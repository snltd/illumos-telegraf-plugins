package smbserver

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosSmbServer{
		Fields: []string{"open_files", "connections"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))
	metric := acc.GetTelegrafMetrics()

	fmt.Printf("%v\n", metric)
}

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
			"connections": float64(2),
			"open_files":  float64(3),
		},
	)
}
