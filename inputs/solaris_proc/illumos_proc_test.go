package illumos_proc

import (
	//"fmt"
	//"github.com/influxdata/telegraf"
	"encoding/gob"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	//"time"
)

func _TestPlugin(t *testing.T) {
	s := &IllumosProc{}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))
}

/*
var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"merp",
		map[string]string{},
		map[string]interface{}{},
		time.Now(),
	),
}
*/
/*
func TestContractMap(t *testing.T) {
	contractMap := contractMap(sampleSvcsOutput)

	assert.Equal(
		t,
		map[int]string{
			67:  "svc:/network/inetd:default",
			72:  "svc:/network/ssh:default",
			78:  "svc:/network/smb/client:default",
			629: "svc:/sysdef/cron_monitor:default",
			83:  "lrc:/etc/rc2_d/S89PRESERVE",
		},
		contractMap,
	)

	assert.Equal(t, "svc:/network/ssh:default", contractMap[72])
}

// This test uses the current running process to test a few easily accessible members of the
// psinfo_t struct. This means it will only work on Illumos, but so will virtually everything else
// in this repo, so hey-ho.
func TestProcPsinfo(t *testing.T) {
	psinfo, err := procPsinfo(os.Getpid())
	assert.Nil(t, err)
	assert.Equal(t, psinfo.Pr_pid, pid_t(os.Getpid()))
	assert.Equal(t, psinfo.Pr_ppid, pid_t(os.Getppid()))
	assert.Equal(t, psinfo.Pr_uid, uid_t(os.Getuid()))
	assert.Equal(t, psinfo.Pr_gid, gid_t(os.Getgid()))
}
*/

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

	fields, tags := parseProcData(psinfoData, usageData)

	require.Equal(
		t,
		map[string]interface{}{
			"percentCPU": float64(0),
			"percentMem": float64(0.0823974609375),
			"rss":        float64(14196736),
			"size":       float64(17248256),
		},
		fields,
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
		tags,
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
	assert.Equal(
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
