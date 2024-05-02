package helpers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Zone struct {
	ID      int
	Name    string
	Status  string
	Path    string
	UUID    string
	Brand   string
	IPType  string
	DebugID int
}

type Vnic struct {
	Name  string
	Zone  string
	Link  string
	Speed int
}

// ZoneMap maps the name of a zone to a zone struct containing all its zoneadm properties.
type ZoneMap map[string]Zone

// ZoneVnicMap maps a VNIC name to a vnic struct which explains it.
type ZoneVnicMap map[string]Vnic

// NewZoneMap creates a ZoneMap describing the current state of the system.
func NewZoneMap() ZoneMap {
	stdout, _, err := RunCmd("/usr/sbin/zoneadm list -cp")
	if err != nil {
		log.Fatal(err)
	}

	return ParseZones(stdout)
}

// Used to generate test fixtures
func NewZoneMapFromText(raw string) ZoneMap {
	return ParseZones(raw)
}

func NewZoneVnicMap() ZoneVnicMap {
	stdout, _, err := RunCmd("/usr/sbin/dladm show-vnic -po link,zone,over,speed")
	if err != nil {
		log.Fatal(err)
	}

	return ParseZoneVnics(stdout)
}

// Names returns a list of zones in the map.
func (z ZoneMap) Names() []string {
	zones := []string{}

	for zone := range z {
		zones = append(zones, zone)
	}

	return zones
}

// ZoneByID returns the zone with the given ID.
func (z ZoneMap) ZoneByID(id int) (Zone, error) {
	for _, zone := range z {
		if zone.ID == id {
			return zone, nil
		}
	}

	return Zone{}, fmt.Errorf("no zone with ID %d", id)
}

// Names returns a list of zones in the map.
func (z ZoneMap) InState(state string) []string {
	zones := []string{}

	for zone, data := range z {
		if data.Status == state {
			zones = append(zones, zone)
		}
	}

	return zones
}

// ZoneName returns the name of the current zone.
func ZoneName() string {
	stdout, _, err := RunCmd("/bin/zonename")
	if err != nil {
		log.Fatal("could not get zonename")
	}

	return stdout
}

// ParseZones turns a chunk of raw `zoneadm list -p` output into a ZoneMap. It is public so
// Telegraf tests can use it.
func ParseZones(raw string) ZoneMap {
	rawZones := strings.Split(raw, "\n")
	ret := ZoneMap{}

	for _, rawZone := range rawZones {
		zone, err := parseZone(rawZone)

		if err == nil {
			ret[zone.Name] = zone
		}
	}

	return ret
}

// parseZone turns a line of raw `zoneadm list -p` output into a zone struct. The format of such a
// line is zoneid:zonename:state:zonepath:uuid:brand:ip-type:debugid.
func parseZone(raw string) (Zone, error) {
	chunks := strings.Split(raw, ":")

	if len(chunks) != 8 {
		return Zone{}, fmt.Errorf("found %d fields", len(chunks))
	}

	if chunks[0] == "-" {
		return Zone{}, errors.New("zone not running")
	}

	zoneID, _ := strconv.Atoi(chunks[0])
	debugID, _ := strconv.Atoi(chunks[7])

	return Zone{
		zoneID,
		chunks[1],
		chunks[2],
		chunks[3],
		chunks[4],
		chunks[5],
		chunks[6],
		debugID,
	}, nil
}

func ParseZoneVnics(raw string) ZoneVnicMap {
	rawVnics := strings.Split(raw, "\n")
	ret := ZoneVnicMap{}

	if raw != "" {
		for _, rawVnic := range rawVnics {
			vnic := parseZoneVnic(rawVnic)
			ret[vnic.Name] = vnic
		}
	}

	return ret
}

func parseZoneVnic(raw string) Vnic {
	chunks := strings.Split(raw, ":")
	speed, _ := strconv.Atoi(chunks[3])

	return Vnic{
		chunks[0],
		chunks[1],
		chunks[2],
		speed,
	}
}
