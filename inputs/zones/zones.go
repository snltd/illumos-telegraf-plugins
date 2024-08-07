package zones

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

func (z *IllumosZones) Description() string {
	return "Report on zone states, brands, and other properties."
}

var (
	sampleConfig = ""
	makeZoneMap  = helpers.NewZoneMap
	zoneDir      = "/etc/zones"
)

type IllumosZones struct{}

func (z *IllumosZones) Gather(acc telegraf.Accumulator) error {
	gatherProperties(acc, makeZoneMap())

	return nil
}

var zoneBootTime = func(zoneName helpers.ZoneName, zoneID int) (interface{}, error) {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return nil, err
	}

	defer token.Close()

	bootTime, err := token.GetNamed("zones", zoneID, string(zoneName), "boot_time")
	// Not being able to get a zone boot time probably isn't really an error. It just means the zone
	// isn't running.
	if err != nil {
		return nil, err
	}

	return helpers.NamedValue(bootTime), nil
}

var zoneUptime = func(zoneName helpers.ZoneName, zoneID int) float64 {
	bootTime, err := zoneBootTime(zoneName, zoneID)
	if err != nil {
		return -1
	}

	return float64(time.Now().Unix()) - bootTime.(float64)
}

// zoneAge tries to give you the age of a zone by inspecting the mtime of the XML file which
// zonecfg(1m) creates when it makes the zone. There may be a better way. Let me know.
func zoneAge(zoneDir string, zoneName helpers.ZoneName) (float64, error) {
	zoneFile := path.Join(zoneDir, fmt.Sprintf("%s.xml", zoneName))

	fh, err := os.Stat(zoneFile)
	if err != nil {
		return 0, err
	}

	return float64(time.Now().Unix() - fh.ModTime().Unix()), nil
}

func gatherProperties(acc telegraf.Accumulator, zonemap helpers.ZoneMap) {
	for zone, zoneData := range zonemap {
		if zone == "global" {
			continue
		}

		tags := map[string]string{
			"name":   string(zone),
			"status": zoneData.Status,
			"ipType": zoneData.IPType,
			"brand":  zoneData.Brand,
		}

		acc.AddFields(
			"zones",
			map[string]interface{}{"uptime": zoneUptime(zoneData.Name, zoneData.ID)},
			tags,
		)

		age, err := zoneAge(zoneDir, zoneData.Name)

		if err == nil {
			acc.AddFields(
				"zones",
				map[string]interface{}{"age": age},
				tags,
			)
		}
	}
}

func (z *IllumosZones) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("illumos_zones", func() telegraf.Input { return &IllumosZones{} })
}
