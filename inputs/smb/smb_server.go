package smbserver

import (
	"log"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## The kstat fields you wish to emit.
	# Fields = ["open_files"]
`

func (s *IllumosSmbServer) Description() string {
	return "Reports illumos in-kernel SMB server statistics"
}

func (s *IllumosSmbServer) SampleConfig() string {
	return sampleConfig
}

type IllumosSmbServer struct {
	Fields []string
}

var getKStats = func() ([]*kstat.KStat, error) {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return []*kstat.KStat{}, err
	}

	token.Close()

	return helpers.KStatsInModule(token, "smbsrv"), nil
}

func (s *IllumosSmbServer) Gather(acc telegraf.Accumulator) error {
	kstats, err := getKStats()

	if err != nil {
		log.Println("failed to get smbsrv kstats")
		return err
	}

	for _, stat := range kstats {
		stats, err := stat.AllNamed()

		if err == nil {
			acc.AddFields(
				"smb.server",
				parseNamedStats(s, stats),
				map[string]string{},
			)
		}
	}

	return nil
}

func parseNamedStats(s *IllumosSmbServer, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if helpers.WeWant(stat.Name, s.Fields) {
			fields[stat.Name] = helpers.NamedValue(stat).(float64)
		}
	}

	return fields
}

func init() {
	inputs.Add("illumos_smb_server", func() telegraf.Input { return &IllumosSmbServer{} })
}
