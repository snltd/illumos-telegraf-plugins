package patches

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

const (
	pkg5Binary       = "/bin/pkg"
	pkginBinary      = "/opt/local/bin/pkgin"
	noUpdatesMessage = "no packages have newer versions available"
)

var sampleConfig = `
	## Whether you wish this plugin to try to refresh the package database. Personally, I wouldn't.
	# refresh = false
`

var runningZones = func() []string {
	zoneMap := helpers.NewZoneMap()
	return zoneMap.InState("running")
}

type IllumosPatches struct {
	Installed   bool
	Upgradeable bool
	Refresh     bool
}

func (s *IllumosPatches) Description() string {
	return "Reports the number of packages which can be upgraded."
}

func (s *IllumosPatches) SampleConfig() string {
	return sampleConfig
}

func gatherZone(zone string, refresh bool) (map[string]interface{}, map[string]string) {
	retValues := make(map[string]interface{})
	retTags := make(map[string]string)

	if havePkg5(zone) {
		if refresh {
			refreshPkg(zone)
		}

		packagesToUpdate, err := toUpdatePkg(zone)

		if err != nil {
			log.Printf("failed to count upgradeable packages (pkg5): %s", err)
		} else {
			retValues["upgradeable"] = packagesToUpdate
			retTags["format"] = "pkg"
			retTags["zone"] = zone
		}
	}

	if havePkgin(zone) {
		packagesToUpdate, err := toUpdatePkgin(zone)

		if err != nil {
			log.Printf("failed to count upgradeable packages (pkgin): %s", err)
		} else {
			retValues["upgradeable"] = packagesToUpdate
			retTags["format"] = "pkgin"
			retTags["zone"] = zone
		}
	}

	return retValues, retTags
}

func toUpdatePkg(zone string) (int, error) {
	raw, err := runPkgListCmd(zone)
	if err != nil {
		return 0, err
	}

	return len(strings.Split(raw, "\n")), nil
}

func toUpdatePkgin(zone string) (int, error) {
	pkginOutput, err := runPkginUpgradeCmd(zone)
	if err != nil {
		log.Print("failed to run pkgin")
		return 0, err
	}

	for _, line := range strings.Split(pkginOutput, "\n") {
		if !strings.Contains(line, "to upgrade:") {
			continue
		}

		fields := strings.Fields(line)
		toUpgrade, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Print("failed to parse upgrade count")
		} else {
			return toUpgrade, nil
		}
	}

	return 0, errors.New("did not find pkgin 'to upgrade' value")
}

func refreshPkg(zone string) error {
	// This needs elevated privileges
	stdout, stderr, err := helpers.RunCmdInZone("/bin/pkg refresh", zone)
	if err != nil {
		log.Print(stdout)
		log.Print(stderr)
		log.Print(err)
		return err
	}

	return nil
}

var runPkgListCmd = func(zone string) (string, error) {
	stdout, stderr, err := helpers.RunCmdInZone("/bin/pkg list -uH", zone)
	// `pkg list -u` exits 1 if there are no packages to upgrade, so an error
	// might not be an error.
	if err != nil {
		if stderr != noUpdatesMessage {
			log.Print(err)
		}
		return "", nil
	}

	return stdout, nil
}

var runPkginUpgradeCmd = func(zone string) (string, error) {
	stdout, stderr, err := helpers.RunCmdInZone("echo n | /opt/local/bin/pkgin upgrade", zone)
	if err != nil {
		log.Print(stderr)
		log.Print(err)
		return "", err
	}

	return stdout, nil
}

func havePkg5(zone string) bool {
	return helpers.HaveFileInZone(pkg5Binary, zone)
}

func havePkgin(zone string) bool {
	return helpers.HaveFileInZone(pkginBinary, zone)
}

func (s *IllumosPatches) Gather(acc telegraf.Accumulator) error {
	for _, zone := range runningZones() {
		values, tags := gatherZone(zone, s.Refresh)
		acc.AddFields("packages", values, tags)
	}

	return nil
}

func init() {
	inputs.Add("illumos_patches", func() telegraf.Input { return &IllumosPatches{} })
}
