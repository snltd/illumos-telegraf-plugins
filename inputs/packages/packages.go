package packages

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
	## Whether to report the number of installed packages
	# installed = true
	## Whether to report the number of upgradeable packages
	# upgradeable = true
	## Use this command to get the elevated privileges run commands in other zones via zlogin. 
	## and to run pkg refresh anywhere. Should be a path, like "/bin/sudo" "/bin/pfexec", but can 
	## also be "none", which will collect only the local zone.
	# elevate_privs_with = "/bin/sudo"
`

var runningZones = func() []helpers.ZoneName {
	zoneMap := helpers.NewZoneMap()
	return zoneMap.InState("running")
}

type IllumosPackages struct {
	ElevatePrivsWith string
	Installed        bool
	Upgradeable      bool
	Refresh          bool
}

func (s *IllumosPackages) Description() string {
	return "Reports the number of packages which can be upgraded."
}

func (s *IllumosPackages) SampleConfig() string {
	return sampleConfig
}

func gatherZone(zone helpers.ZoneName, s *IllumosPackages) (map[string]interface{}, map[string]string) {
	retValues := make(map[string]interface{})
	retTags := make(map[string]string)
	prefix := s.ElevatePrivsWith

	if havePkg5(zone) {
		// Continue if this fails, but log it
		if s.Refresh {
			err := refreshPkg(zone, s.ElevatePrivsWith)
			if err != nil {
				log.Printf("Failed to pkg refresh: %s", err)
			}
		}

		packagesToUpdate, err := toUpdatePkg(zone, prefix)

		if err != nil {
			log.Printf("failed to count upgradeable packages (pkg5): %s", err)
		} else {
			retValues["upgradeable"] = packagesToUpdate
			retTags["format"] = "pkg"
			retTags["zone"] = string(zone)
		}
	}

	if havePkgin(zone) {
		packagesToUpdate, err := toUpdatePkgin(zone, prefix)

		if err != nil {
			log.Printf("failed to count upgradeable packages (pkgin): %s", err)
		} else {
			retValues["upgradeable"] = packagesToUpdate
			retTags["format"] = "pkgin"
			retTags["zone"] = string(zone)
		}
	}

	return retValues, retTags
}

func toUpdatePkg(zone helpers.ZoneName, cmdPrefix string) (int, error) {
	raw, err := runPkgListCmd(cmdPrefix, zone)

	if err != nil || raw == "" {
		return 0, err
	}

	return len(strings.Split(raw, "\n")), nil
}

func toUpdatePkgin(zone helpers.ZoneName, cmdPrefix string) (int, error) {
	pkginOutput, err := runPkginUpgradeCmd(cmdPrefix, zone)
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

func refreshPkg(zone helpers.ZoneName, cmdPrefix string) error {
	_, _, err := helpers.RunCmdInZone(cmdPrefix, "/bin/pkg refresh", zone)
	if err != nil {
		log.Printf("Error refreshing pkgin: %s", err)
		return err
	}

	return nil
}

const pkgListCommand = "/bin/pkg list -uH"

var runPkgListCmd = func(cmdPrefix string, zone helpers.ZoneName) (string, error) {
	stdout, stderr, err := helpers.RunCmdInZone(cmdPrefix, pkgListCommand, zone)
	// `pkg list -u` exits 1 if there are no packages to upgrade, so an error
	// might not be an error.
	if err != nil {
		if strings.TrimSpace(stderr) == noUpdatesMessage {
			return "", nil
		} else {
			log.Printf("Error running %s in %s", pkgListCommand, zone)
			return "", err
		}
	}

	return stdout, nil
}

var runPkginUpgradeCmd = func(cmdPrefix string, zone helpers.ZoneName) (string, error) {
	stdout, _, err := helpers.RunCmdInZone(cmdPrefix, "echo n | /opt/local/bin/pkgin upgrade", zone)
	if err != nil {
		log.Print("Error running pkg upgrade", err)
		return "", err
	}

	return stdout, nil
}

func havePkg5(zone helpers.ZoneName) bool {
	return helpers.HaveFileInZone(pkg5Binary, zone)
}

func havePkgin(zone helpers.ZoneName) bool {
	return helpers.HaveFileInZone(pkginBinary, zone)
}

func (s *IllumosPackages) Gather(acc telegraf.Accumulator) error {
	var zones []helpers.ZoneName

	if s.ElevatePrivsWith == "none" {
		zones = []helpers.ZoneName{helpers.CurrentZone()}
	} else {
		zones = runningZones()
	}

	for _, zone := range zones {
		values, tags := gatherZone(zone, s)
		acc.AddFields("packages", values, tags)
	}

	return nil
}

func init() {
	inputs.Add("illumos_packages", func() telegraf.Input { return &IllumosPackages{} })
}
