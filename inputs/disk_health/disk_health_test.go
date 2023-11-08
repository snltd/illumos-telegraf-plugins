package diskhealth

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestCamelCase(t *testing.T) {
	t.Parallel()
  actual, err := camelCase("Soft Errors")
  require.NoError(t, err)
	require.Equal(t, "softErrors", actual)

  actual, err = camelCase("word")
  require.NoError(t, err)
	require.Equal(t, "word", actual)

  actual, err = camelCase("One tWO three")
  require.NoError(t, err)
	require.Equal(t, "oneTwoThree", actual)

  actual, err = camelCase("")
  require.Error(t, err)
}

func TestParseNamedStatsBlkDev(t *testing.T) {
	t.Parallel()

	s := &IllumosDiskHealth{
		Devices: []string{"sd6"},
		Fields:  []string{"Hard Errors", "Soft Errors", "Transport Errors", "Illegal Request"},
		Tags:    []string{"Vendor", "Serial No", "Product", "Revision"},
	}

	testData := helpers.FromFixture("blkdeverr--0--blkdev0,err.kstat")
	fields, tags := parseNamedStats(s, testData)

	require.Equal(
		t,
		fields,
		map[string]interface{}{
			"hardErrors":      float64(0),
			"illegalRequest":  float64(0),
			"softErrors":      float64(0),
			"transportErrors": float64(0),
		},
	)

	require.Equal(
		t,
		tags,
		map[string]string{
			"revision": "P9CR30A",
			"serialNo": "2301E699B2E7",
		},
	)
}

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosDiskHealth{
		Devices: []string{"sd6"},
		Fields:  []string{"Hard Errors", "Soft Errors", "Transport Errors", "Illegal Request"},
		Tags:    []string{"Vendor", "Serial No", "Product", "Revision"},
	}

	testData := helpers.FromFixture("sderr--6--sd6,err.kstat")
	fields, tags := parseNamedStats(s, testData)

	require.Equal(
		t,
		fields,
		map[string]interface{}{
			"hardErrors":      float64(0),
			"illegalRequest":  float64(1148),
			"softErrors":      float64(0),
			"transportErrors": float64(0),
		},
	)

	require.Equal(
		t,
		tags,
		map[string]string{
			"product":  "My Passport 2627",
			"revision": "4008",
			"serialNo": "WXP1E7916Z6K",
			"vendor":   "WD",
		},
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosDiskHealth{
		Devices: []string{"sd0"},
		Fields:  []string{"Hard Errors", "Transport Errors", "Illegal Request"},
		Tags:    []string{"Vendor", "Serial No", "Product"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testMetric := acc.GetTelegrafMetrics()[0]
	require.Equal(t, "diskHealth", testMetric.Name())

	requiredFields := []string{"hardErrors", "transportErrors", "illegalRequest"}

	for _, field := range requiredFields {
		_, present := testMetric.GetField(field)
		require.True(t, present)
	}

	requiredTags := []string{"vendor", "serialNo", "product"}

	for _, tag := range requiredTags {
		require.True(t, testMetric.HasTag(tag))
	}
}
