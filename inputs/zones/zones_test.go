package zones

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosZones{}

	// Return a predictable boot time for each zone in the zoneadmOutput
	zoneBootTime = func(zoneName helpers.ZoneName, _ int) (interface{}, error) {
		var ts float64

		if zoneName == "cube-media" {
			ts = 321222
		} else {
			ts = 80607
		}

		return float64(time.Now().Unix()) - ts, nil
	}

	zoneDir = "testdata/zones"

	makeZoneMap = func() helpers.ZoneMap {
		return helpers.ParseZones(zoneadmOutput)
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

func TestZoneAge(t *testing.T) {
	t.Parallel()

	age, err := zoneAge("testdata/zones", "cube-ws")

	require.Nil(t, err)
	require.Equal(t, age, float64(time.Now().Unix()-zoneTimestamp("cube-ws")))
}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179`

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "installed",
			"ipType": "excl",
			"brand":  "pkgsrc",
			"name":   "cube-ws",
		},
		map[string]interface{}{
			"uptime": float64(80607),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "installed",
			"ipType": "excl",
			"brand":  "pkgsrc",
			"name":   "cube-ws",
		},
		map[string]interface{}{
			"age": float64(time.Now().Unix() - zoneTimestamp("cube-ws")),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "running",
			"ipType": "excl",
			"brand":  "lipkg",
			"name":   "cube-media",
		},
		map[string]interface{}{
			"uptime": float64(321222),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "running",
			"ipType": "excl",
			"brand":  "lipkg",
			"name":   "cube-media",
		},
		map[string]interface{}{
			"age": float64(time.Now().Unix() - zoneTimestamp("cube-media")),
		},
		time.Now(),
	),
}

func zoneTimestamp(zone string) int64 {
	if zone == "cube-ws" {
		return 1588780462 // epoch timestamp of cube-ws.xml test file. 2021-01-27T12:49
	}

	return 1611751778
}
