package illumos_local_zone

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
	## This plugin outputs raw kstats. Most things will need to be wrapped in a rate() type
	## function in your graphing software.  Use 'kstat -pc zone_caps' to see what kstat names are
	## available, and select the ones you want. You get the 'usage' (what you are using at this
	## moment) and 'value' (the maximum available to you) values.
	# fields = ["swapresv", "lockedmem", "nprocs", "cpucaps", "physicalmem"]
	## You just get "usage" and "value" fields for almost everything in 'Names' , but cpucaps
	## has more information. Select the fields you want here.  There's no need to include 'usage' or
	## 'value', they'll be done anyway.
	# cpu_cap_on = true
	# cpu_cap_fields = ["above_base_sec", "above_sec", "baseline", "below_sec", "burst_limit_sec",
	# "bursting_sec", "effective", "maxusage", "nwait"]
	## Fields you require from the 'memory_cap' kstat module. Use 'kstat -pm memory_cap' to view
	## them.  'rss' and 'size' are gauges: they do not need converting to rates
	# memory_cap_on = true
	# memory_cap_fields = ["anon_alloc_fail", "anonpgin", "crtime", "execpgin", "fspgin",
	# "n_pf_throttle", "n_pf_throttle_usec", "nover", "pagedout", "pgpgin", "physcap", "rss",
	# "swap", "swapcap"]
`

func (s *LocalZone) Description() string {
	return "Reports metrics particular to a local  zone"
}

func (s *LocalZone) SampleConfig() string {
	return sampleConfig
}

/*
func zoneId() int {
	raw := helpers.RunCmd("/usr/sbin/zoneadm list -p")
	id, _ := strconv.Atoi(strings.Split(raw, ":")[0])
	return id
}
*/

type LocalZone struct {
	Fields          []string
	CpuCapOn        bool
	CpuCapFields    []string
	MemoryCapOn     bool
	MemoryCapFields []string
}

func (s *LocalZone) Gather(acc telegraf.Accumulator) error {
	tags := make(map[string]string)
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	zoneCapStats := helpers.KStatsInClass(token, "zone_caps")

	for _, name := range zoneCapStats {
		namedStats, err := name.AllNamed()

		if err != nil {
			log.Fatal("cannot get memorycap named stats")
		}

		//fmt.Printf("--> %v\n", parseNamedStats(s.Fields, namedStats))

		nice_name := strings.Split(name.Name, "_")[0]
		fmt.Println(nice_name)

		/*
			if !helpers.WeWant(nice_name, s.Names) {
				continue
			}

			for _, stat := range stats {
				if stat.Name == "zonename" {
					tags["zone"] = stat.StringVal
					continue
				}

				field := fmt.Sprintf("%s.%s", nice_name, stat.Name)

				if stat.Name == "value" || stat.Name == "usage" {
					fields[field] = stat.UintVal
				}

				if nice_name == "cpucaps" && helpers.WeWant(stat.Name, s.CpuCapsFields) {
					fields[field] = stat.UintVal
				}
			}
		*/
	}

	if s.MemoryCapOn {
		memoryCapStats := helpers.KStatsInModule(token, "memory_cap")

		for _, name := range memoryCapStats {
			namedStats, err := name.AllNamed()

			if err != nil {
				log.Fatal("cannot get memorycap named stats")
			}

			acc.AddFields("local", parseNamedStats(s.MemoryCapFields, namedStats), tags)
		}
	}

	token.Close()
	fmt.Println()
	return nil
}

func parseNamedStats(requiredFields []string, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if helpers.WeWant(stat.Name, requiredFields) {
			value := helpers.NamedValue(stat)

			switch value.(type) {
			case string:
				log.Printf("cannot turn '%s' field into a value", stat.Name)
			default:
				fields[stat.Name] = value
			}
		}
	}

	return fields
}

/*
	fields := make(map[string]interface{})


*/

func init() {
	inputs.Add("local_zone", func() telegraf.Input { return &LocalZone{} })
}
