package nfsserver

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
	## The NFS versions you wish to monitor.
	# nfs_versions = ["v3", "v4"]
	## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfsproccnt' lists the
	## possibilities
	# fields = ["read", "write", "remove", "create", "getattr", "setattr"]
`

func (s *IllumosNfsServer) Description() string {
	return "Reports illumos NFS server statistics"
}

func (s *IllumosNfsServer) SampleConfig() string {
	return sampleConfig
}

type IllumosNfsServer struct {
	Fields      []string
	NfsVersions []string
}

func (s *IllumosNfsServer) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return err
	}

	stats := helpers.KStatsInModule(token, "nfs")

	for _, stat := range stats {
		if !strings.HasPrefix(stat.Name, "rfsproccnt_v") {
			continue
		}

		nfsVersion := fmt.Sprintf("v%s", stat.Name[len(stat.Name)-1:])

		if !helpers.WeWant(nfsVersion, s.NfsVersions) {
			continue
		}

		stats, err := stat.AllNamed()

		if err == nil {
			acc.AddFields(
				"nfs.server",
				parseNamedStats(s, stats),
				map[string]string{"nfsVersion": nfsVersion},
			)
		} else {
			log.Printf("cannot get named NFS server kstats for %s\n", stat.Name)
		}
	}

	token.Close()

	return nil
}

func parseNamedStats(s *IllumosNfsServer, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if helpers.WeWant(stat.Name, s.Fields) {
			fields[stat.Name] = helpers.NamedValue(stat).(float64)
		}
	}

	return fields
}

func init() {
	inputs.Add("illumos_nfs_server", func() telegraf.Input { return &IllumosNfsServer{} })
}
