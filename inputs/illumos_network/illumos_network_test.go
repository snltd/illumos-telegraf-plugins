package illumos_network

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "ipackets64"},
		Zones:  []string{"cube-dns"},
	}

	testData := helpers.FromFixture("link--0--dns_net0.kstat")
	fields := parseNamedStats(s, testData)

	require.Equal(
		t,
		map[string]interface{}{
			"obytes64":   float64(69053870),
			"rbytes64":   float64(1518773044),
			"ipackets64": float64(1637072),
		},
		fields,
	)
}

func TestParseNamedStatsNoSelectedNics(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "ipackets64"},
		Zones:  []string{"cube-dns"},
		Vnics:  []string{"net0"},
	}

	testData := helpers.FromFixture("link--0--dns_net0.kstat")
	fields := parseNamedStats(s, testData)
	require.Equal(t, map[string]interface{}{}, fields)
}

func TestZoneTags(t *testing.T) {
	t.Parallel()

	zoneName = "global"

	require.Equal(
		t,
		map[string]string{
			"zone":  "cube-dns",
			"link":  "rge0",
			"speed": "1000mbit",
			"name":  "dns_net0",
		},
		zoneTags("cube-dns", "dns_net0", helpers.ParseZoneVnics(sampleDladmOutput)["dns_net0"]),
	)
}

func TestZoneTagsGlobal(t *testing.T) {
	t.Parallel()

	zoneName = "global"

	require.Equal(
		t,
		map[string]string{
			"zone":  "global",
			"link":  "none",
			"speed": "unknown",
			"name":  "rge0",
		},
		zoneTags("global", "rge0", helpers.ParseZoneVnics(sampleDladmOutput)["rge0"]),
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "collisions", "ierrors"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	metric := acc.GetTelegrafMetrics()[0]
	require.Equal(t, "net", metric.Name())
	require.True(t, metric.HasTag("zone"))
	require.True(t, metric.HasTag("link"))
	require.True(t, metric.HasTag("speed"))
	require.True(t, metric.HasTag("name"))

	for _, field := range s.Fields {
		_, present := metric.GetField(field)
		require.True(t, present)
	}
}

var sampleDladmOutput = `media_net0:cube-media:rge0:1000
dns_net0:cube-dns:rge0:1000
pkgsrc_net0:cube-pkgsrc:rge0:1000
backup_net0:cube-backup:rge0:1000`
