package illumos_memory

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"github.com/stretchr/testify/require"
)

func TestParseSwap(t *testing.T) {
	t.Parallel()

	s := &IllumosMemory{}

	runSwapCmd = func() string {
		return "total: 2852796k bytes allocated + 1950828k reserved = 4803624k used, 2638448k available"
	}

	require.Equal(
		t,
		map[string]interface{}{
			"allocated": float64(2921263104),
			"reserved":  float64(1997647872),
			"used":      float64(4918910976),
			"available": float64(2701770752),
		},
		parseSwap(s),
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosMemory{
		SwapOn:         true,
		SwapFields:     []string{"allocated", "reserved", "used", "available"},
		ExtraOn:        true,
		ExtraFields:    []string{"kernel", "arcsize", "freelist"},
		VminfoOn:       true,
		VminfoFields:   []string{"swap_alloc", "swap_avail", "swap_free", "swap_resv"},
		CpuvmOn:        true,
		CpuvmFields:    []string{"pgin", "anonpgin", "pgout", "anonpgout"},
		CpuvmAggregate: false,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	// 'swap -s' metrics
	swapMetric := acc.GetTelegrafMetrics()[0]
	require.Equal(t, "memory.swap", swapMetric.Name())

	for _, field := range s.SwapFields {
		_, present := swapMetric.GetField(field)
		require.True(t, present)
	}

	// "extra" metrics
	extraMetric := acc.GetTelegrafMetrics()[1]
	require.Equal(t, "memory", extraMetric.Name())

	for _, field := range s.ExtraFields {
		_, present := extraMetric.GetField(field)
		require.True(t, present)
	}

	// vminfo metrics
	vminfoMetric := acc.GetTelegrafMetrics()[2]
	require.Equal(t, "memory.vminfo", vminfoMetric.Name())

	vminfoMetricFields := []string{"swapAlloc", "swapAvail", "swapFree", "swapResv"}

	for _, field := range vminfoMetricFields {
		_, present := vminfoMetric.GetField(field)
		require.True(t, present)
	}

	// cpu_vm metrics. I think we'll always have CPU0
	cpuvmMetric := acc.GetTelegrafMetrics()[3]
	require.Equal(t, "memory.cpuVm", cpuvmMetric.Name())

	for _, field := range s.CpuvmFields {
		fieldName := fmt.Sprintf("vm.cpu0.%s", field)
		_, present := cpuvmMetric.GetField(fieldName)
		require.True(t, present)
	}

	_, present := cpuvmMetric.GetField("vm.aggregate.pgin")
	require.False(t, present)
}

func TestPluginAggregates(t *testing.T) {
	t.Parallel()

	s := &IllumosMemory{
		CpuvmOn:        true,
		CpuvmFields:    []string{"pgin", "anonpgin", "pgout", "anonpgout"},
		CpuvmAggregate: true,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	cpuvmMetric := acc.GetTelegrafMetrics()[0]
	require.Equal(t, "memory.cpuVm", cpuvmMetric.Name())

	for _, field := range s.CpuvmFields {
		fieldName := fmt.Sprintf("vm.cpu0.%s", field)
		_, present := cpuvmMetric.GetField(fieldName)
		require.False(t, present)
	}

	for _, field := range s.CpuvmFields {
		fieldName := fmt.Sprintf("vm.aggregate.%s", field)
		_, present := cpuvmMetric.GetField(fieldName)
		require.True(t, present)
	}
}

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosMemory{
		CpuvmOn:        true,
		CpuvmFields:    []string{"pgin", "anonpgin", "pgout", "anonpgout"},
		CpuvmAggregate: false,
	}

	testData := helpers.FromFixture("cpu:0:vm.kstat")
	fields := parseNamedStats(s, testData)

	require.Equal(
		t,
		fields,
		map[string]interface{}{
			"pgin":      float64(4245),
			"anonpgin":  float64(397),
			"pgout":     float64(836),
			"anonpgout": float64(9935),
		},
	)
}

func TestAggregateCpuVmKStats(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		map[string]interface{}{
			"vm.aggregate.anonpgin":  float64(864),
			"vm.aggregate.anonpgout": float64(19083),
			"vm.aggregate.pgin":      float64(9600),
			"vm.aggregate.pgout":     float64(1609),
		},
		aggregateCpuvmKStats(sampleStatHolder),
	)
}

func TestIndividualCpuvmKStats(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		map[string]interface{}{
			"vm.cpu0.anonpgin":  float64(397),
			"vm.cpu0.anonpgout": float64(9935),
			"vm.cpu0.pgin":      float64(4245),
			"vm.cpu0.pgout":     float64(836),
			"vm.cpu1.anonpgin":  float64(467),
			"vm.cpu1.anonpgout": float64(9148),
			"vm.cpu1.pgin":      float64(5355),
			"vm.cpu1.pgout":     float64(773),
		},
		individualCpuvmKStats(sampleStatHolder),
	)
}

var sampleStatHolder = cpuvmStatHolder{
	0: map[string]interface{}{
		"anonpgin":  float64(397),
		"anonpgout": float64(9935),
		"pgin":      float64(4245),
		"pgout":     float64(836),
	},
	1: map[string]interface{}{
		"anonpgin":  float64(467),
		"anonpgout": float64(9148),
		"pgin":      float64(5355),
		"pgout":     float64(773),
	},
}
