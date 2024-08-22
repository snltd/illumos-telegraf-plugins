package fma

import (
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestParseFmstatLine(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		Fmstat{
			module: "fmd-self-diagnosis",
			props: map[string]float64{
				"ev_recv": float64(367),
				"ev_acpt": float64(0),
				"wait":    float64(0),
				"svc_t":   float64(25.7),
				"pc_w":    float64(0),
				"pc_b":    float64(0),
				"open":    float64(0),
				"solve":   float64(0),
				"memsz":   float64(0),
				"bufsz":   float64(0),
			},
		},
		parseFmstatLine(
			"fmd-self-diagnosis     367       0  0.0   25.7   0   0     0     0      0      0",
			parseFmstatHeader("module   ev_recv ev_acpt wait  svc_t  %w  %b  open solve  memsz  bufsz"),
		),
	)
}

func TestParseFmstatHeader(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		[]string{
			"module",
			"ev_recv",
			"ev_acpt",
			"wait",
			"svc_t",
			"pc_w",
			"pc_b",
			"open",
			"solve",
			"memsz",
			"bufsz",
		},
		parseFmstatHeader("module      ev_recv ev_acpt wait  svc_t  %w  %b  open solve  memsz  bufsz"),
	)
}

func TestParseFmadmImpact(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		map[string]string{
			"module":            "mem",
			"motherboard":       "0",
			"chip":              "0",
			"memory-controller": "0",
			"dimm":              "0",
			"rank":              "0",
			"status":            "faulty",
		},
		parseFmadmImpact("mem:///motherboard=0/chip=0/memory-controller=0/dimm=0/rank=0       faulty"),
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosFma{
		Fmadm:         true,
		Fmstat:        true,
		FmstatFields:  []string{"svc_t", "open", "memsz", "bufsz"},
		FmstatModules: []string{"software-response", "zfs-retire"},
	}

	runFmadmFaultyCmd = func(cmdPrefix string) string {
		ret, _ := os.ReadFile("testdata/fmadm_output.txt")

		return string(ret)
	}

	runFmstatCmd = func() string {
		ret, _ := os.ReadFile("testdata/fmstat_output.txt")

		return string(ret)
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

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"fma.fmadm",
		map[string]string{
			"module": "zfs",
			"pool":   "big",
			"vdev":   "3706b5d93e20f727",
			"status": "faulted",
		},
		map[string]interface{}{
			"faults": 1,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"fma.fmstat",
		map[string]string{
			"module": "software-response",
		},
		map[string]interface{}{
			"svc_t": float64(0.9),
			"open":  float64(0),
			"memsz": float64(2355.2),
			"bufsz": float64(2048),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"fma.fmstat",
		map[string]string{
			"module": "zfs-retire",
		},
		map[string]interface{}{
			"svc_t": float64(377.8),
			"open":  float64(0),
			"memsz": float64(4),
			"bufsz": float64(0),
		},
		time.Now(),
	),
}
