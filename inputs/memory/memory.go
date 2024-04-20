package memory

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## Whether to produce metrics from the output of 'swap -s'
	# swap_on = true
	## And which fields to use. Specifying none implies all.
	# swap_fields = ["allocated", "reserved", "used", "available"]
	## Whether to report "extra" fields, and which ones (kernel, arcsize, freelist)
	# extra_on = true
	# extra_fields = ["kernel", "arcsize", "freelist"]
  ## Whether to collect vminfo kstats, and which ones.
	# vminfo_on = true
	# vminfo_fields = ["freemem", "swap_alloc", "swap_avail", "swap_free", "swap_resv"]
	## Whether to collect cpu::vm kstats
	# cpuvm_on =true
	# cpuvm_fields = ["pgin", "anonpgin", "pgpgin", "pgout", "anonpgout", "pgpgout"]
	## Whether to aggregate cpuvm fields. (False sents a set of metrics for each vcpu)
	# cpuvm_aggregate = false
  ## Whether to collect zone memory_cap fields, and which ones
	# zone_memcap_on = true
	# zone_memcap_zones = []
	# zone_memcap_fields = ["physcap", "rss", "swap"]
`

var pageSize float64

func (s *IllumosMemory) Description() string {
	return "Reports on Illumos virtual and physical memory usage."
}

func (s *IllumosMemory) SampleConfig() string {
	return sampleConfig
}

type IllumosMemory struct {
	SwapOn           bool
	SwapFields       []string
	ExtraOn          bool
	ExtraFields      []string
	VminfoOn         bool
	VminfoFields     []string
	CpuvmOn          bool
	CpuvmFields      []string
	CpuvmAggregate   bool
	ZoneMemcapOn     bool
	ZoneMemcapZones  []string
	ZoneMemcapFields []string
}

func (s *IllumosMemory) Gather(acc telegraf.Accumulator) error {
	tags := make(map[string]string)

	if s.SwapOn {
		acc.AddFields("memory.swap", parseSwap(s), tags)
	}

	token, err := kstat.Open()

	if err != nil {
		return err
	}

	defer token.Close()

	if s.ExtraOn {
		acc.AddFields("memory", extraKStats(s, token), tags)
	}

	if s.VminfoOn {
		acc.AddFields("memory.vminfo", vminfoKStats(s, token), tags)
	}

	if s.CpuvmOn {
		acc.AddFields("memory.cpuVm", cpuvmKStats(s, token), tags)
	}

	if s.ZoneMemcapOn {
		gatherZoneMemcapStats(s, acc, token)
	}

	return nil
}

func parseZoneMemcapStats(s *IllumosMemory, stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		if !helpers.WeWant(stat.Name, s.ZoneMemcapFields) || !helpers.WeWant(stat.KStat.Name, s.ZoneMemcapZones) {
			continue
		}

		tags["zone"] = stat.KStat.Name
		fields[stat.Name] = helpers.NamedValue(stat).(float64)
	}

	return fields, tags
}

func gatherZoneMemcapStats(s *IllumosMemory, acc telegraf.Accumulator, token *kstat.Token) error {
	memcapStats := helpers.KStatsInModule(token, "memory_cap")

	for _, stat := range memcapStats {
		namedStats, err := stat.AllNamed()

		if err != nil {
			log.Printf("cannot get named zone_memcap stats for %s\n", stat.Name)
			return err
		}

		fields, tags := parseZoneMemcapStats(s, namedStats)
		acc.AddFields("memory.zone", fields, tags)
	}

	return nil
}

func extraKStats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	fields := make(map[string]interface{})

	// My error handling here is permissive. I don't see why failing to get one kstat should stop us
	// trying to get the rest, so errors are noted and otherwise ignored. We'll see how it goes.

	if helpers.WeWant("kernel", s.ExtraFields) {
		stat, err := token.GetNamed("unix", 0, "system_pages", "pp_kernel")

		if err == nil {
			fields["kernel"] = helpers.NamedValue(stat).(float64) * pageSize
		} else {
			log.Print("could not get kernel kstats")
		}
	}

	if helpers.WeWant("freelist", s.ExtraFields) {
		stat, err := token.GetNamed("unix", 0, "system_pages", "pagesfree")

		if err == nil {
			fields["freelist"] = helpers.NamedValue(stat).(float64) * pageSize
		} else {
			log.Print("could not get freelist kstats")
		}
	}

	if helpers.WeWant("arcsize", s.ExtraFields) {
		stat, err := token.GetNamed("zfs", 0, "arcstats", "size")

		if err == nil {
			fields["arcsize"] = helpers.NamedValue(stat).(float64)
		} else {
			log.Print("could not get freelist kstats")
		}
	}

	return fields
}

// The raw kstats in here are gauges, measured in pages. So we need to convert them to bytes here,
// and you need to apply some kind of rate() function in your graphing software.
func vminfoKStats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	fields := make(map[string]interface{})

	_, vi, err := token.Vminfo()

	if err != nil {
		log.Print("cannot get vminfo kstats")

		return fields
	}

	if helpers.WeWant("freemem", s.VminfoFields) {
		fields["freemem"] = float64(vi.Freemem) * pageSize
	}

	if helpers.WeWant("swap_alloc", s.VminfoFields) {
		fields["swapAlloc"] = float64(vi.Alloc) * pageSize
	}

	if helpers.WeWant("swap_avail", s.VminfoFields) {
		fields["swapAvail"] = float64(vi.Avail) * pageSize
	}

	if helpers.WeWant("swap_free", s.VminfoFields) {
		fields["swapFree"] = float64(vi.Free) * pageSize
	}

	if helpers.WeWant("swap_resv", s.VminfoFields) {
		fields["swapResv"] = float64(vi.Resv) * pageSize
	}

	return fields
}

// The only named stats we need to parse in this collector are the ones from cpuvmKStats().
func parseNamedStats(s *IllumosMemory, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if helpers.WeWant(stat.Name, s.CpuvmFields) {
			fields[stat.Name] = helpers.NamedValue(stat).(float64)
		}
	}

	return fields
}

type cpuvmStatHolder map[int]map[string]interface{}

func perCpuvmKStats(s *IllumosMemory, token *kstat.Token) cpuvmStatHolder {
	perCPUStats := make(cpuvmStatHolder)
	modStats := helpers.KStatsInModule(token, "cpu")

	for _, statGroup := range modStats {
		if statGroup.Name != "vm" {
			continue
		}

		stats, err := statGroup.AllNamed()

		if err == nil {
			perCPUStats[statGroup.Instance] = parseNamedStats(s, stats)
		} else {
			log.Print("cannot get named cpu.vm kstats")
		}
	}

	return perCPUStats
}

func individualCpuvmKStats(stats cpuvmStatHolder) map[string]interface{} {
	fields := make(map[string]interface{})

	for cpuID, vmStats := range stats {
		for name, value := range vmStats {
			fieldName := fmt.Sprintf("vm.cpu%d.%s", cpuID, name)
			fields[fieldName] = value
		}
	}

	return fields
}

func aggregateCpuvmKStats(stats cpuvmStatHolder) map[string]interface{} {
	counters := make(map[string]float64)

	for _, vmStats := range stats {
		for name, value := range vmStats {
			fieldName := fmt.Sprintf("vm.aggregate.%s", name)
			counters[fieldName] += value.(float64)
		}
	}

	fields := make(map[string]interface{})

	for key, val := range counters {
		fields[key] = val
	}

	return fields
}

func cpuvmKStats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	allStats := perCpuvmKStats(s, token)

	if s.CpuvmAggregate {
		return aggregateCpuvmKStats(allStats)
	}

	return individualCpuvmKStats(allStats)
}

var runSwapCmd = func() string {
	stdout, stderr, err := helpers.RunCmd("/usr/sbin/swap -s")

	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func parseSwap(s *IllumosMemory) map[string]interface{} {
	fields := make(map[string]interface{})
	swapline := runSwapCmd()
	re := regexp.MustCompile(`total: (\d+k) [\w ]* \+ (\d+k).*= (\d+k) used, (\d+k).*$`)
	m := re.FindAllStringSubmatch(swapline, -1)[0]

	if helpers.WeWant("allocated", s.SwapFields) {
		bytes, err := helpers.Bytify(m[1])

		if err == nil {
			fields["allocated"] = bytes
		}
	}

	if helpers.WeWant("reserved", s.SwapFields) {
		bytes, err := helpers.Bytify(m[2])

		if err == nil {
			fields["reserved"] = bytes
		}
	}

	if helpers.WeWant("used", s.SwapFields) {
		bytes, err := helpers.Bytify(m[3])

		if err == nil {
			fields["used"] = bytes
		}
	}

	if helpers.WeWant("available", s.SwapFields) {
		bytes, err := helpers.Bytify(m[4])

		if err == nil {
			fields["available"] = bytes
		}
	}

	return fields
}

func init() {
	pageSize = float64(os.Getpagesize())

	inputs.Add("illumos_memory", func() telegraf.Input { return &IllumosMemory{} })
}
