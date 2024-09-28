package network

import (
	"fmt"
	"log"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## The kstat fields you wish to emit. 'kstat -c net' will show what is collected. Defining
	## no fields sends everything, which is probably not what you want.
	# fields = ["obytes64", "rbytes64"]
	## The VNICs you wish to observe. Again, specifying none collects all.
	# vnics  = ["net0"]
	## The zones you wish to monitor. Specifying none collects all.
	# zones = ["zone1", "zone2"]`

func (s *IllumosNetwork) Description() string {
	return "Reports on illumos NIC Usage. Zone-aware."
}

func (s *IllumosNetwork) SampleConfig() string {
	return sampleConfig
}

type IllumosNetwork struct {
	Zones  []helpers.ZoneName
	Fields []string
	Vnics  []string
}

type zoneTagMap struct {
	zone  helpers.ZoneName
	link  string
	speed string
	name  string
}

var (
	makeZoneVnicMap = helpers.NewZoneVnicMap
	zoneName        = helpers.ZoneName("")
)

func (s *IllumosNetwork) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return err
	}

	defer token.Close()

	links := helpers.KStatsInModule(token, "link")

	for _, link := range links {
		// links are of the form link:0:dns_net0 for non-global zones, and link:0:rge0 (net) for the
		// global. (On Solaris the module number corresponds to the zone ID, but not on Illumos.)
		stats, _ := link.AllNamed()

		if err != nil {
			log.Printf("cannot get named link kstats for %s\n", link.Name)
		}

		vnicMap := makeZoneVnicMap()
		vnic := vnicMap[link.Name]
		zone := vnic.Zone

		// If our vnicMap can't tell us which zone this belongs to, let's assume that it belongs to
		// the current zone. This might need to be smarter, but it's a reasonable first step. It
		// might be nice to pull some info about the physical NIC out into tags.
		if zone == "" {
			zone = zoneName
		}

		if !helpers.WeWant(zone, s.Zones) {
			continue
		}

		zoneTags := zoneTags(zone, link.Name, vnic)
		zoneTagsMap := make(map[string]string)
		zoneTagsMap["zone"] = string(zoneTags.zone)
		zoneTagsMap["link"] = zoneTags.link
		zoneTagsMap["speed"] = zoneTags.speed
		zoneTagsMap["name"] = zoneTags.name

		acc.AddFields(
			"net",
			parseNamedStats(s, stats),
			zoneTagsMap,
		)
	}

	return nil
}

func zoneTags(zone helpers.ZoneName, link string, vnic helpers.Vnic) zoneTagMap {
	if zone == zoneName {
		return zoneTagMap{
			zone:  zoneName,
			link:  "none",
			speed: "unknown",
			name:  link,
		}
	}

	return zoneTagMap{
		zone:  vnic.Zone,
		link:  vnic.Link,
		speed: fmt.Sprintf("%dmbit", vnic.Speed),
		name:  vnic.Name,
	}
}

func parseNamedStats(s *IllumosNetwork, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if !helpers.WeWant(stat.Name, s.Fields) || !helpers.WeWant(stat.KStat.Name, s.Vnics) {
			continue
		}

		fields[stat.Name] = helpers.NamedValue(stat).(float64)
	}

	return fields
}

func init() {
	zoneName = helpers.CurrentZone()

	inputs.Add("illumos_network", func() telegraf.Input { return &IllumosNetwork{} })
}
