package process

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

const (
	zoneMapTxt = "0:global:running:/::ipkg:shared:0\n6:serv-build:running:/:bb60d6b1-de11-4714-a748-fd67dfde44bd:native:excl:0"
	ctidMapTxt = "   433 svc:/test/service:default\n - svc:/system/rbac:default\n  845 svc:/system/cron:default"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosProcess{
		Values:            []string{"rtime", "size"},
		Tags:              []string{"name", "uid", "zoneid", "contract"},
		TopK:              2,
		ExpandZoneTag:     true,
		ExpandContractTag: true,
	}

	procRootDir = "testdata/proc"

	loadProcPsinfo = func(pid int) (psinfo_t, error) {
		return psinfoFromFixture(pid), nil
	}

	loadProcUsage = func(pid int) (prusage_t, error) {
		return usageFromFixture(pid), nil
	}

	collectContractInfo = func() (string, error) {
		return ctidMapTxt, nil
	}

	newZoneMap = func() helpers.ZoneMap {
		return helpers.NewZoneMapFromText(zoneMapTxt)
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	var testMetricsFull = []telegraf.Metric{
		testutil.MustMetric(
			"process",
			map[string]string{
				"zone":     "serv-build",
				"uid":      "0",
				"zoneid":   "6",
				"contract": "845",
				"name":     "cron",
				"service":  "svc:/system/cron:default",
			},
			map[string]interface{}{
				"rtime": 55522972213516,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"process",
			map[string]string{
				"zone":     "serv-build",
				"uid":      "264",
				"zoneid":   "6",
				"contract": "1415",
				"name":     "zsh",
			},
			map[string]interface{}{
				"rtime": 23376390518801,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"process",
			map[string]string{
				"zone":     "serv-build",
				"uid":      "264",
				"zoneid":   "6",
				"contract": "1415",
				"name":     "zsh",
			},
			map[string]interface{}{
				"size": 9805824,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"process",
			map[string]string{
				"zone":     "serv-build",
				"uid":      "0",
				"zoneid":   "6",
				"contract": "845",
				"name":     "cron",
				"service":  "svc:/system/cron:default",
			},
			map[string]interface{}{
				"size": 1978368,
			},
			time.Now(),
		),
	}

	testutil.RequireMetricsEqual(
		t,
		testMetricsFull,
		acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func TestTopKPids(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{10, 40, 20, 60}, topKPids(&testObject, "rtime", 4))
	require.Equal(t, []int{10, 40, 20}, topKPids(&testObject, "rtime", 3))
	require.Equal(t, []int{30, 10, 60}, topKPids(&testObject, "sysc", 3))
}

func TestToNs(t *testing.T) {
	t.Parallel()
	require.Equal(t, int64(1714216749000000123), toNs(timestruc_t{1714216749, 123}))
}

func TestNewProcObjectCannotLoadProc(t *testing.T) {
	loadProcPsinfo = func(pid int) (psinfo_t, error) {
		return psinfo_t{}, errors.New("fake test error")
	}

	_, err := newProcObject(&IllumosProcess{}, 999)
	require.Error(t, err)
}

func TestNewProcObject(t *testing.T) {
	t.Parallel()

	s := &IllumosProcess{
		Values: []string{"rtime", "size"},
		Tags:   []string{"name", "uid"},
	}

	loadProcPsinfo = func(pid int) (psinfo_t, error) {
		return psinfoFromFixture(pid), nil
	}

	loadProcUsage = func(pid int) (prusage_t, error) {
		return usageFromFixture(pid), nil
	}

	procRootDir = "testdata/proc"

	result, err := newProcObject(s, 26939)

	require.NoError(t, err)
	require.IsType(t, procObject{}, result)

	require.Equal(
		t,
		procObjectValues{
			"rtime": int64(23376390518801),
			"size":  int64(9805824),
		},
		result.Values,
	)

	require.Equal(
		t,
		procObjectTags{
			"name": "zsh",
			"uid":  "264",
		},
		result.Tags,
	)
}

func TestNewProcObjectNoFilters(t *testing.T) {
	t.Parallel()

	s := &IllumosProcess{}
	procRootDir = "testdata/proc"

	loadProcPsinfo = func(pid int) (psinfo_t, error) {
		return psinfoFromFixture(pid), nil
	}

	loadProcUsage = func(pid int) (prusage_t, error) {
		return usageFromFixture(pid), nil
	}

	result, err := newProcObject(s, 8055)

	require.NoError(t, err)
	require.IsType(t, procObject{}, result)

	require.Equal(
		t,
		procObjectValues{
			"rtime":  int64(55522972213516),
			"utime":  int64(722329),
			"stime":  int64(1697875),
			"wtime":  int64(8630699),
			"inblk":  int64(0),
			"oublk":  int64(0),
			"sysc":   int64(149),
			"ioch":   int64(4050),
			"size":   int64(1978368),
			"rssize": int64(1146880),
			"pctcpu": int64(0),
			"pctmem": int64(2),
			"nlwp":   int64(1),
			"count":  int64(1),
		},
		result.Values,
	)

	require.Equal(
		t,
		procObjectTags{
			"name":     "cron",
			"uid":      "0",
			"gid":      "0",
			"euid":     "0",
			"egid":     "0",
			"taskid":   "769",
			"projid":   "0",
			"zoneid":   "6",
			"contract": "845",
			"pid":      "8055",
			"ppid":     "4231",
		},
		result.Tags,
	)
}

func TestExpandZoneTag(t *testing.T) {
	t.Parallel()
	zoneMap := helpers.NewZoneMapFromText(zoneMapTxt)

	testObject := procObjectMap{
		8055: procObject{
			Values: procObjectValues{"ioch": int64(4050)},
			Tags:   procObjectTags{"zoneid": "6"},
		},
		26939: procObject{
			Values: procObjectValues{"ioch": int64(1133708)},
			Tags:   procObjectTags{"zoneid": "0"},
		},
		1234: procObject{
			Values: procObjectValues{"ioch": int64(57834)},
			Tags:   procObjectTags{"zoneid": "10"}, // zone not in map
		},
	}

	expandZoneTag(&testObject, zoneMap)

	require.Equal(
		t,
		procObjectMap{
			8055: procObject{
				Values: procObjectValues{"ioch": int64(4050)},
				Tags:   procObjectTags{"zoneid": "6", "zone": "serv-build"},
			},
			26939: procObject{
				Values: procObjectValues{"ioch": int64(1133708)},
				Tags:   procObjectTags{"zoneid": "0", "zone": "global"},
			},
			1234: procObject{
				Values: procObjectValues{"ioch": int64(57834)},
				Tags:   procObjectTags{"zoneid": "10"},
			},
		},
		testObject,
	)
}

func TestExpandContractTag(t *testing.T) {
	t.Parallel()
	ctidMap := newContractMap(ctidMapTxt)

	testObject := procObjectMap{
		8055: procObject{
			Values: procObjectValues{"ioch": int64(4050)},
			Tags:   procObjectTags{"contract": "433"},
		},
		26939: procObject{
			Values: procObjectValues{"ioch": int64(1133708)},
			Tags:   procObjectTags{"contract": "999"},
		},
	}

	expandContractTag(&testObject, ctidMap)

	require.Equal(
		t,
		procObjectMap{
			8055: procObject{
				Values: procObjectValues{"ioch": int64(4050)},
				Tags: procObjectTags{
					"contract": "433",
					"service":  "svc:/test/service:default",
				},
			},
			26939: procObject{
				Values: procObjectValues{"ioch": int64(1133708)},
				Tags:   procObjectTags{"contract": "999"},
			},
		},
		testObject,
	)
}

func TestNewProcObjectMap(t *testing.T) {
	t.Parallel()

	s := &IllumosProcess{
		Values: []string{"ioch"},
		Tags:   []string{"euid"},
	}

	procRootDir = "testdata/proc"

	result := newProcObjectMap(s, allProcs())
	require.Equal(
		t,
		procObjectMap{
			8055: procObject{
				Values: procObjectValues{"ioch": int64(4050)},
				Tags:   procObjectTags{"euid": "0"},
			},
			26939: procObject{
				Values: procObjectValues{"ioch": int64(1133708)},
				Tags:   procObjectTags{"euid": "264"},
			},
		},
		result,
	)
}

func TestAllprocs(t *testing.T) {
	t.Parallel()
	procRootDir = "testdata/proc"
	result := allProcs()
	require.IsType(t, []fs.DirEntry{}, result)
	require.Equal(t, 2, len(result))
}

func TestNewContractMap(t *testing.T) {
	t.Parallel()
	testData, _ := os.ReadFile("testdata/svcs--contract_map")

	require.Equal(
		t,
		newContractMap(string(testData)),
		contractMap{
			248: "svc:/system/svc/restarter:default",
			345: "svc:/network/ip-interface-management:default",
			359: "svc:/network/netcfg:default",
			594: "svc:/network/ipmp:default",
			622: "svc:/system/pfexec:default",
			638: "svc:/system/name-service-cache:default",
			641: "svc:/application/cups/scheduler:default",
			706: "svc:/system/utmp:default",
			727: "svc:/system/console-login:default",
			812: "svc:/network/rpc/bind:default",
			817: "svc:/system/fmd:default",
			828: "svc:/network/inetd:default",
			845: "svc:/system/cron:default",
			852: "svc:/system/filesystem/autofs:default",
			859: "svc:/system/system-log:rsyslog",
			860: "svc:/network/ssh:default",
			861: "svc:/system/sac:default",
			869: "lrc:/etc/rc2_d/S20sysetup",
			870: "lrc:/etc/rc2_d/S89PRESERVE",
		},
	)

	require.Equal(
		t,
		newContractMap("junk"),
		contractMap{},
	)
}

func psinfoFromFixture(pid int) psinfo_t {
	var ret psinfo_t

	filename := path.Join("testdata", "proc", fmt.Sprint(pid), "psinfo")
	raw, err := os.Open(filename)

	if err != nil {
		log.Fatalf("Could not load serialized data from disk: %v\n", err)
	}

	dec := gob.NewDecoder(raw)
	err = dec.Decode(&ret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load decode prstat data: %v\n", err)
		os.Exit(1)
	}

	return ret
}

func usageFromFixture(pid int) prusage_t {
	var ret prusage_t

	filename := path.Join("testdata", "proc", fmt.Sprint(pid), "usage")
	raw, err := os.Open(filename)

	if err != nil {
		log.Fatalf("Could not load serialized data from disk: %v\n", err)
	}

	dec := gob.NewDecoder(raw)
	err = dec.Decode(&ret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load decode prusage data: %v\n", err)
		os.Exit(1)
	}

	return ret
}

var testObject = procObjectMap{
	10: procObject{
		Values: procObjectValues{"rtime": int64(1000), "sysc": int64(111)},
		Tags:   procObjectTags{"zoneid": "6", "uid": "264"},
	},
	20: procObject{
		Values: procObjectValues{"rtime": int64(200), "sysc": int64(22)},
		Tags:   procObjectTags{"zoneid": "2", "uid": "2"},
	},
	30: procObject{
		Values: procObjectValues{"rtime": int64(33), "sysc": int64(333)},
		Tags:   procObjectTags{"zoneid": "3", "uid": "0"},
	},
	40: procObject{
		Values: procObjectValues{"rtime": int64(404), "sysc": int64(22)},
		Tags:   procObjectTags{"zoneid": "6", "uid": "264"},
	},
	50: procObject{
		Values: procObjectValues{"rtime": int64(55), "sysc": int64(5)},
		Tags:   procObjectTags{"zoneid": "6", "uid": "0"},
	},
	60: procObject{
		Values: procObjectValues{"rtime": int64(66), "sysc": int64(66)},
		Tags:   procObjectTags{"zoneid": "2", "uid": "14"},
	},
}
