package nfsserver

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

// The meat of the plugin is tested by TestParseNamedStats. This exercises the full code path,
// hittng real kstats.
func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosNfsServer{
		Fields:      []string{"read", "write", "remove", "create"},
		NfsVersions: []string{"v4"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))
	metric := acc.GetTelegrafMetrics()[0]

	require.Equal(t, "nfs.server", metric.Name())
	require.True(t, metric.HasTag("nfsVersion"))

	for _, field := range s.Fields {
		_, present := metric.GetField(field)
		require.True(t, present)
	}
}

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosNfsServer{
		Fields:      []string{"read", "write", "remove", "create"},
		NfsVersions: []string{"v4"},
	}

	testData := helpers.FromFixture("nfs--0--rfsproccnt_v4.kstat")
	fields := parseNamedStats(s, testData)

	require.Equal(
		t,
		fields,
		map[string]interface{}{
			"read":   float64(902),
			"write":  float64(1310),
			"remove": float64(94),
			"create": float64(6),
		},
	)
}
