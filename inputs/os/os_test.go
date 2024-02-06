package os

import (
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestParseOsRelease(t *testing.T) {
	t.Parallel()
	sampleData, _ := os.ReadFile("testdata/os-release")

	require.Equal(
		t,
		map[string]string{
			"name":     "OmniOS",
			"version":  "r151048f",
			"build_id": "151048.6.2023.12.10",
		},
		parseOsRelease(string(sampleData)),
	)

	require.Empty(t, parseOsRelease("some junk"))
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	osRelease = "testdata/os-release"
	s := &IllumosOS{}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
	)
}

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"os",
		map[string]string{
			"name":     "OmniOS",
			"build_id": "151048.6.2023.12.10",
			"version":  "r151048f",
		},
		map[string]interface{}{
			"release": 1,
		},
		time.Now(),
	),
}
