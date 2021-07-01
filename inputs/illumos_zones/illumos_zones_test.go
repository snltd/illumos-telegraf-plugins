package illumos_zones

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
	zoneBootTime = func(zoneName string, _ int) (interface{}, error) {
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
			"age": float64(time.Now().Unix() - 1623333996),
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
			"age": float64(time.Now().Unix() - 1623333990),
		},
		time.Now(),
	),
}
