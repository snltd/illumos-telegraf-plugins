package illumos_zfs_arc

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosZfsArc{
		Fields: []string{"hits", "l2_hits", "prefetch_data_hits", "prefetch_metadata_hits"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	metric := acc.GetTelegrafMetrics()[0]

	require.Equal(t, "zfs.arcstats", metric.Name())
	require.Zero(t, len(metric.TagList()))

	for _, field := range s.Fields {
		_, present := metric.GetField(field)
		require.True(t, present)
	}
}

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosZfsArc{
		Fields: []string{"c", "prefetch_data_hits", "prefetch_data_misses"},
	}

	testData := helpers.FromFixture("zfs:0:arcstats.kstat")
	fields := parseNamedStats(s, testData)

	require.Equal(
		t,
		map[string]interface{}{
			"c":                    float64(6753403688),
			"prefetch_data_hits":   float64(1209174),
			"prefetch_data_misses": float64(1807386),
		},
		fields,
	)
}
