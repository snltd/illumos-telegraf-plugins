package cpu

/*
Collects information about illumos CPU usage. The values it outputs are the raw kstat values,
which means they are counters, and they only go up. I wrap them in a rate() function in
Wavefront, which is plenty good enough for me.

Features to add, possibly:
	- option to aggregate CPU metrics across all cores and/or CPUs.
	- deal with multiple cores AND multiple physical processors. I don't have a machine with the
	  latter.
	- emit rates rather than raw values. This would let us do proper percentages.
*/

import (
	"fmt"
	"log"
	"strings"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
  ## Report stuff from the cpu_info kstat. As of now it's just the current clock speed and some
  ## potentially useful tags
  # cpu_info_stats = true
  ## Produce metrics for sys and user CPU consumption in every zone
  # zone_cpu_stats = true
  ## Which cpu:sys kstat metrics you wish to emit. They probably won't all work, because they
  ## some will have a value type which is not an unsigned int
  # sys_fields = ["cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"]
  ## "cpu_ticks_idle", cpu_ticks_kernel", cpu_ticks_user", cpu_ticks_wait", }
`

func (s *IllumosCPU) Description() string {
	return "Reports on illumos CPU usage"
}

func (s *IllumosCPU) SampleConfig() string {
	return sampleConfig
}

type IllumosCPU struct {
	CPUInfoStats bool
	ZoneCPUStats bool
	SysFields    []string
}

func parseCPUinfoKStats(stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		switch stat.Name {
		case "current_clock_Hz":
			fields["speed"] = float64(stat.UintVal)
		case "clock_MHz":
			tags["clockMHz"] = fmt.Sprintf("%d", stat.IntVal)
		case "state":
			tags["state"] = stat.StringVal
		case "chip_id":
			tags["chipID"] = fmt.Sprintf("%d", stat.IntVal)
		case "core_id":
			tags["coreID"] = fmt.Sprintf("%d", stat.IntVal)
		}
	}

	return fields, tags
}

func gatherCPUinfoStats(acc telegraf.Accumulator, token *kstat.Token) error {
	stats := helpers.KStatsInModule(token, "cpu_info")

	for _, stat := range stats {
		namedStats, err := stat.AllNamed()
		if err != nil {
			log.Print("cannot get kstat token")

			return err
		}

		fields, tags := parseCPUinfoKStats(namedStats)
		acc.AddFields("cpu.info", fields, tags)
	}

	return nil
}

func parseZoneCPUKStats(stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		switch stat.Name {
		case "nsec_sys":
			fields["sys"] = float64(stat.UintVal)
		case "nsec_user":
			fields["user"] = float64(stat.UintVal)
		case "zonename":
			tags["name"] = stat.StringVal
		}
	}

	return fields, tags
}

// metrics reporting on CPU consumption for each zone. sys and user, each as a gauge, tagged with
// the zone name.
func gatherZoneCPUStats(acc telegraf.Accumulator, token *kstat.Token) error {
	zoneStats := helpers.KStatsInModule(token, "zones")

	for _, zone := range zoneStats {
		namedStats, err := zone.AllNamed()
		if err != nil {
			log.Print("cannot get zone CPU named stats")

			return err
		}

		fields, tags := parseZoneCPUKStats(namedStats)

		acc.AddFields("cpu.zone", fields, tags)
	}

	return nil
}

func parseSysCPUKStats(s *IllumosCPU, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if helpers.WeWant(stat.Name, s.SysFields) {
			fields[fieldToMetricPath(stat.Name)] = float64(stat.UintVal)
		}
	}

	return fields
}

func gatherSysCPUStats(s *IllumosCPU, acc telegraf.Accumulator, token *kstat.Token) error {
	cpuStats := helpers.KStatsInModule(token, "cpu")

	for _, cpu := range cpuStats {
		if cpu.Name == "sys" {
			namedStats, err := cpu.AllNamed()
			if err != nil {
				log.Print("cannot get CPU named stats")

				return err
			}

			acc.AddFields(
				"cpu",
				parseSysCPUKStats(s, namedStats),
				map[string]string{"coreID": fmt.Sprintf("%d", cpu.Instance)},
			)
		}
	}

	return nil
}

func fieldToMetricPath(field string) string {
	field = strings.Replace(field, "cpu_", "", 1)
	field = strings.Replace(field, "_", ".", 1)

	return field
}

func (s *IllumosCPU) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()

	if err != nil {
		log.Print("cannot get kstat token")

		return err
	}

	defer token.Close()

	if s.CPUInfoStats {
		err := gatherCPUinfoStats(acc, token)
		if err != nil {
			return err
		}
	}

	if s.ZoneCPUStats {
		err := gatherZoneCPUStats(acc, token)
		if err != nil {
			return err
		}
	}

	err = gatherSysCPUStats(s, acc, token)

	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("illumos_cpu", func() telegraf.Input { return &IllumosCPU{} })
}
