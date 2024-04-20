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
	# fields = ["open_files"] 
`

func (s *IllumosSmbServer) Description() string {
	return "Reports Illumos in-kernel SMB server statistics"
}

func (s *IllumosSmbServer) SampleConfig() string {
	return sampleConfig
}

type IllumosSmbServer struct {
	Fields []string
}

func (s *IllumosSmbServer) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return err
	}

	stats := helpers.KStatsInModule(token, "smbsrv")

	for _, stat := range stats {
		stats, err := stat.AllNamed()

		if err == nil {
			acc.AddFields(
				"smb.server",
				parseNamedStats(s, stats),
				map[string]string{},
			)
		}
	}

	token.Close()

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
