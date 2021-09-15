package zpool

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestPluginAllMetrics(t *testing.T) {
	t.Parallel()

	s := &IllumosZpool{}

	zpoolOutput = func() string {
		return sampleOutput
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetricsFull,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetricsFull = []telegraf.Metric{
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "big",
		},
		map[string]interface{}{
			"size":   3.98023209254912e+12,
			"alloc":  2.95768627871744e+12,
			"free":   1.029718409216e+12,
			"frag":   2,
			"cap":    74,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "fast",
		},
		map[string]interface{}{
			"size":   2.81320357888e+11,
			"alloc":  1.11669149696e+11,
			"free":   1.69651208192e+11,
			"frag":   25,
			"cap":    39,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"size":   2.13674622976e+11,
			"alloc":  6.13106581504e+10,
			"free":   1.52471339008e+11,
			"frag":   63,
			"cap":    28,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
}

func TestPluginSelectedMetrics(t *testing.T) {
	t.Parallel()

	s := &IllumosZpool{
		Fields: []string{"cap", "health"},
	}

	zpoolOutput = func() string {
		return sampleOutput
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetricsSelected,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetricsSelected = []telegraf.Metric{
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "big",
		},
		map[string]interface{}{
			"cap":    74,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "fast",
		},
		map[string]interface{}{
			"cap":    39,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"cap":    28,
			"health": 0,
		},
		time.Now(),
	),
}

func TestPluginSinglePoolWithStatus(t *testing.T) {
	t.Parallel()

	s := &IllumosZpool{
		Fields: []string{"alloc"},
		Status: true,
	}

	zpoolOutput = func() string {
		return sampleSinglePoolOutput
	}

	zpoolStatusOutput = func(pool string) string {
		return sampleStatusNormalOutput
	}

	timeSince = func(timestamp time.Time) float64 {
		return 10000
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetricsSelectedStatus,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetricsSelectedStatus = []telegraf.Metric{
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"alloc": float64(6.13106581504e+10),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool.status",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"resilverTime":   float64(0),
			"scrubTime":      float64(0),
			"timeSinceScrub": float64(10000),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool.status.errors",
		map[string]string{
			"device": "rpool",
			"state":  "ONLINE",
		},
		map[string]interface{}{
			"read":  float64(0),
			"write": float64(0),
			"cksum": float64(0),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool.status.errors",
		map[string]string{
			"device": "mirror-0",
			"state":  "ONLINE",
		},
		map[string]interface{}{
			"read":  float64(0),
			"write": float64(0),
			"cksum": float64(0),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool.status.errors",
		map[string]string{
			"device": "c2t2d0s1",
			"state":  "ONLINE",
		},
		map[string]interface{}{
			"read":  float64(0),
			"write": float64(0),
			"cksum": float64(0),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool.status.errors",
		map[string]string{
			"device": "c2t3d0s1",
			"state":  "ONLINE",
		},
		map[string]interface{}{
			"read":  float64(0),
			"write": float64(0),
			"cksum": float64(0),
		},
		time.Now(),
	),
}

// Function tests

func TestHealthtoi(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, healthtoi("ONLINE"))
	require.Equal(t, 1, healthtoi("DEGRADED"))
	require.Equal(t, 2, healthtoi("SUSPENDED"))
	require.Equal(t, 3, healthtoi("UNAVAIL"))
	require.Equal(t, 99, healthtoi("what the heck is this nonsense"))
}

func TestParseZpool(t *testing.T) {
	t.Parallel()

	line := "big    3.62T  2.69T   959G        -         -     2%    74%  1.00x  ONLINE  -"

	require.Equal(
		t,
		Zpool{
			name: "big",
			props: map[string]interface{}{
				"size":   3.98023209254912e+12,
				"alloc":  2.95768627871744e+12,
				"free":   1.029718409216e+12,
				"frag":   2,
				"cap":    74,
				"dedup":  1.0,
				"health": 0,
			},
		},
		parseZpool(line, header),
	)
}

func TestParseHeader(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		[]string{
			"name", "size", "alloc", "free", "ckpoint", "expandsz", "frag", "cap", "dedup",
			"health", "altroot",
		},
		parseHeader(header))
}

func TestResilverTime(t *testing.T) {
	t.Parallel()

	require.Greater(
		t,
		resilverTime(sampleStatusResilverOutput),
		float64(6000),
	)

	require.Equal(
		t,
		float64(0),
		resilverTime(sampleStatusNormalOutput),
	)
}

func TestScrubTime(t *testing.T) {
	t.Parallel()

	require.Greater(
		t,
		scrubTime(sampleStatusScrubbingOutput),
		float64(100),
	)

	require.Equal(
		t,
		float64(0),
		scrubTime(sampleStatusUnscrubbedOutput),
	)
}

func TestTimeSinceScrub(t *testing.T) {
	t.Parallel()

	require.Greater(
		t,
		timeSinceScrub(sampleStatusNormalOutput),
		float64(6000),
	)

	require.Equal(
		t,
		float64(0),
		timeSinceScrub(sampleStatusUnscrubbedOutput),
	)
}

var header = "NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT"

var sampleOutput = `NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT
big    3.62T  2.69T   959G        -         -     2%    74%  1.00x  ONLINE  -
fast    262G   104G   158G        -         -    25%    39%  1.00x  ONLINE  -
rpool   199G  57.1G   142G        -         -    63%    28%  1.00x  ONLINE  -`

var sampleStatusResilverOutput = `  pool: big
 state: ONLINE
status: One or more devices is currently being resilvered.  The pool will
        continue to function, possibly in a degraded state.
action: Wait for the resilver to complete.
  scan: resilver in progress since Sun Sep 12 15:11:35 2021
        243M scanned at 20.2M/s, 344K issued at 28.7K/s, 2.56T total
        0 resilvered, 0.00% done, no estimated completion time
config:

        NAME        STATE     READ WRITE CKSUM
        big         ONLINE       0     0     0
          mirror-0  ONLINE       0     0     0
            c2t0d0  ONLINE       0     0     0
            c2t1d0  ONLINE       0     0     0

errors: No known data errors
`

var sampleStatusNormalOutput = `    pool: rpool
 state: ONLINE
  scan: scrub repaired 0 in 0 days 00:03:10 with 0 errors on Fri Feb 19 17:09:54 2021
config:

        NAME          STATE     READ WRITE CKSUM
        rpool         ONLINE       0     0     0
          mirror-0    ONLINE       0     0     0
            c2t2d0s1  ONLINE       0     0     0
            c2t3d0s1  ONLINE       0     0     0

errors: No known data errors
`

var sampleStatusUnscrubbedOutput = `  pool: fast
 state: ONLINE
  scan: scrub canceled on Tue Feb 16 15:24:24 2021
config:

        NAME          STATE     READ WRITE CKSUM
        fast          ONLINE       0     0     0
          mirror-0    ONLINE       0     0     0
            c2t2d0s2  ONLINE       0     0     0
            c2t3d0s2  ONLINE       0     0     0

errors: No known data errors
`

var sampleStatusScrubbingOutput = `  pool: rpool
 state: ONLINE
  scan: scrub in progress since Sun Sep 12 22:33:06 2021
        11.4G scanned at 3.80G/s, 1.29M issued at 440K/s, 65.8G total
        0 repaired, 0.00% done, no estimated completion time
config:

        NAME          STATE     READ WRITE CKSUM
        rpool         ONLINE       0     0     0
          mirror-0    ONLINE       0     0     0
            c2t2d0s1  ONLINE       0     0     0
            c2t3d0s1  ONLINE       0     0     0

errors: No known data errors
`

var sampleSinglePoolOutput = `NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT
rpool   199G  57.1G   142G        -         -    63%    28%  1.00x  ONLINE  -`

func TestExtractErrorCounts(t *testing.T) {
	require.Equal(
		t,
		[]statusErrorCount{
			statusErrorCount{"tank", "UNAVAIL", 0, 0, 0},
			statusErrorCount{"c1t0d0", "ONLINE", 0, 0, 0},
			statusErrorCount{"c1t1d0", "UNAVAIL", 4, 1, 0},
		},
		extractErrorCounts(sampleStatusErrorOutput),
	)
}

var sampleStatusErrorOutput = `  pool: tank
 state: UNAVAIL
status: One or more devices are faulted in response to IO failures.
action: Make sure the affected devices are connected, then run 'zpool clear'.
   see: http://www.sun.com/msg/ZFS-8000-HC
 scrub: scrub completed after 0h0m with 0 errors on Tue Feb  2 13:08:42 2010
config:

        NAME        STATE     READ WRITE CKSUM
        tank        UNAVAIL      0     0     0  insufficient replicas
          c1t0d0    ONLINE       0     0     0
          c1t1d0    UNAVAIL      4     1     0  cannot open

errors: Permanent errors have been detected in the following files:

/tank/data/aaa
/tank/data/bbb
/tank/data/ccc
`
