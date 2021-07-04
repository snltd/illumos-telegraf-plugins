package io

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// This test is sketchy - feed in some captured kstat data.
func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosIO{
		Devices: []string{"sd1"},
		Modules: []string{"sd", "zfs"},
		Fields:  []string{"reads", "nread", "writes", "nwritten"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	metric := acc.GetTelegrafMetrics()[0]
	require.Equal(t, "io", metric.Name())

	for _, field := range s.Fields {
		require.True(t, metric.HasField(field))
	}

	require.True(t, metric.HasTag("serialNo"))
	require.True(t, metric.HasTag("product"))
}
