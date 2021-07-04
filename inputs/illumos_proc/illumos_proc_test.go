package illumos_proc

import (
	"encoding/gob"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPlugin(t *testing.T) {
	s := &IllumosProc{
		TopFields:     []string{"size", "rss"},
		TopFieldLimit: 3,
		DetailedProcs: []string{"vim", "cron"},
		DetailFields:  []string{"userTime", "systemTime"},
		Tags:          []string{"pid", "execname", "uid"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	for _, metric := range acc.GetTelegrafMetrics() {
		fmt.Println(metric.Name())
		switch metric.Name() {
		case "proc.top":
			testProcTopMetric(s, t, metric)
		//case "proc.top.rss":
		//testProcTopMetric(s, t, metric)
		case "proc.detail":
			testProcDetailMetric(s, t, metric)
		//case "proc.detail.cron":
		////testProcDetailMetric(s, t, metric)
		default:
			t.Logf("no match for %s", metric.Name())
			require.Equal(t, 0, 1) // because this should never happen
		}

	}
}

func testProcDetailMetric(s *IllumosProc, t *testing.T, metric telegraf.Metric) {
	t.Helper()

	for _, field := range s.DetailFields {
		require.True(t, metric.HasField(field))
	}
}

func testProcTopMetric(s *IllumosProc, t *testing.T, metric telegraf.Metric) {
	t.Helper()

	for _, field := range s.TopFields {
		require.True(t, metric.HasField(field))
	}

	require.Equal(t, len(s.TopFields), len(metric.FieldList()))
	require.Equal(t, len(s.Tags), len(metric.TagList()))

	for _, tag := range s.Tags {
		require.True(t, metric.HasTag(tag))
	}
}

func TestTopProcesses(t *testing.T) {
	t.Parallel()

	sortedInput := sortProcessList(unsortedInput, "size")

	require.Equal(
		t,
		[]processDetail{
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(4),
					"rss":  float64(40),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(3),
					"rss":  float64(10),
				},
				Tags: map[string]string{},
			},
		},
		topProcesses(sortedInput, 2),
	)
}

func TestSortProcessList(t *testing.T) {
	require.Equal(
		t,
		unsortedInput,
		sortProcessList(unsortedInput, "no-such-field"),
	)

	require.Equal(
		t,
		[]processDetail{
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(4),
					"rss":  float64(40),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(3),
					"rss":  float64(10),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(2),
					"rss":  float64(20),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(1),
					"rss":  float64(30),
				},
				Tags: map[string]string{},
			},
		},
		sortProcessList(unsortedInput, "size"),
	)

	require.Equal(
		t,
		[]processDetail{
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(4),
					"rss":  float64(40),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(1),
					"rss":  float64(30),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(2),
					"rss":  float64(20),
				},
				Tags: map[string]string{},
			},
			processDetail{
				Fields: map[string]interface{}{
					"size": float64(3),
					"rss":  float64(10),
				},
				Tags: map[string]string{},
			},
		},
		sortProcessList(unsortedInput, "rss"),
	)
}

func TestParseProcInfo(t *testing.T) {
	t.Parallel()

	var psinfoData psinfo_t
	var usageData prusage_t

	raw, _ := os.Open("testdata/19230.psinfo")
	dec := gob.NewDecoder(raw)
	dec.Decode(&psinfoData)

	raw, _ = os.Open("testdata/19230.usage")
	dec = gob.NewDecoder(raw)
	dec.Decode(&usageData)

	expected := parseProcData(psinfoData, usageData)

	require.Equal(
		t,
		map[string]interface{}{
			"percentCPU": float64(0),
			"percentMem": float64(0.0823974609375),
			"rss":        float64(14196736),
			"size":       float64(17248256),
			"userTime":   float64(94.762231982),
			"systemTime": float64(6.176471327),
			"waitTime":   float64(1.271980574),
		},
		expected.Fields,
	)

	require.Equal(
		t,
		map[string]string{
			"args":       "vi -p illumos_proc_test.go illumos_proc.go types.go",
			"contractID": "40532",
			"execname":   "vim",
			"pid":        "19230",
			"gid":        "14",
			"uid":        "264",
			"zoneID":     "13",
		},
		expected.Tags,
	)
}

func TestReadProcPsinfo(t *testing.T) {
	// making this parallel fails....
	actual, err := readProcPsinfo(os.Getpid())

	require.Nil(t, err)
	require.Equal(t, os.Getpid(), int(actual.Pr_pid))
	require.Equal(t, os.Getppid(), int(actual.Pr_ppid))

	_, err = readProcPsinfo(0)
	require.Error(t, err)
}

func TestReadProcUsage(t *testing.T) {
	// making this parallel fails....
	actual, err := readProcUsage(os.Getpid())

	require.Nil(t, err)
	require.IsType(t, timestruc_t{}, actual.Pr_utime)

	_, err = readProcUsage(0)
	require.Error(t, err)
}

func TestProcPidList(t *testing.T) {
	t.Parallel()

	procDir = "testdata/proc"
	require.Equal(t, []int{10887, 11022, 8530}, procPidList())
	fmt.Println("DELETE ME AT THE END")
}

/*
func TestZoneLookup(t *testing.T) {
	require.Equal(
		t,
		"cube-media",
		zoneLookup(42),
	)

}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179`

var sampleSvcsOutput = `83 lrc:/etc/rc2_d/S89PRESERVE
   629 svc:/sysdef/cron_monitor:default
     - svc:/sysdef/puppet:default
     - svc:/network/netmask:default
    72 svc:/network/ssh:default
    67 svc:/network/inetd:default
    78 svc:/network/smb/client:default
     - svc:/system/boot-archive:default`
*/

var unsortedInput = []processDetail{
	processDetail{
		Fields: map[string]interface{}{
			"size": float64(3),
			"rss":  float64(10),
		},
		Tags: map[string]string{},
	},
	processDetail{
		Fields: map[string]interface{}{
			"size": float64(1),
			"rss":  float64(30),
		},
		Tags: map[string]string{},
	},
	processDetail{
		Fields: map[string]interface{}{
			"size": float64(4),
			"rss":  float64(40),
		},
		Tags: map[string]string{},
	},
	processDetail{
		Fields: map[string]interface{}{
			"size": float64(2),
			"rss":  float64(20),
		},
		Tags: map[string]string{},
	},
}
