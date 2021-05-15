package illumos_zfs_arc

import (
	"log"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

var sampleConfig = `
	# fields = ["hits", "misses", "l2_hits", "l2_misses", "prefetch_data_hits",
	# "prefetch_data_misses", "prefetch_metadata_hits", "prefetch_metadata_misses",
	# "demand_data_hits", "demand_data_misses", "demand_metadata_hits", "demand_metadata_misses",
	# "l2_size", "l2_read_bytes", "l2_write_bytes", "l2_cksum_bad", "c", "size"]
`

func (s *IllumosZfsArc) Description() string {
	return "Reports Illumos ZFS ARC statistics"
}

func (s *IllumosZfsArc) SampleConfig() string {
	return sampleConfig
}

type IllumosZfsArc struct {
	Fields []string
}

func (s *IllumosZfsArc) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	stats := sth.KStatsInModule(token, "zfs")

	for _, statGroup := range stats {
		if statGroup.Name == "arcstats" {
			namedStats, err := statGroup.AllNamed()

			if err != nil {
				log.Fatal("failed to get named ZFS arcstats")
			}

			acc.AddFields(
				"zfs.arcstats",
				parseNamedStats(s, namedStats),
				map[string]string{},
			)
		}
	}

	token.Close()

	return nil
}

func parseNamedStats(s *IllumosZfsArc, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if sth.WeWant(stat.Name, s.Fields) {
			fields[stat.Name] = sth.NamedValue(stat).(float64)
		}
	}

	return fields
}

func init() {
	inputs.Add("illumos_zfs_arc", func() telegraf.Input { return &IllumosZfsArc{} })
}
