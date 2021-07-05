package patches

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

const (
	pkg5  = "/bin/pkg"
	pkgin = "/opt/local/bin/pkgin"
)

var sampleConfig = ""

type IllumosPatches struct {
	Installed   bool
	Upgradeable bool
}

func (s *IllumosPatches) Description() string {
	return "Reports the number of packages which can be upgraded."
}

func (s *IllumosPatches) SampleConfig() string {
	return sampleConfig
}

func (s *IllumosPatches) Gather(acc telegraf.Accumulator) error {
	if havePkg5() {
		acc.AddFields(
			"packages",
			map[string]interface{}{"upgradeable": toUpdatePkg()},
			map[string]string{"format": "pkg"},
		)
	}

	if havePkgin() {
		acc.AddFields(
			"packages",
			map[string]interface{}{"upgradeable": toUpdatePkgin()},
			map[string]string{"format": "pkgin"},
		)
	}

	return nil
}

func toUpdatePkg() int {
	raw := runPkgListCmd()

	if raw == "" {
		return 0
	}

	return len(strings.Split(raw, "\n"))
}

func toUpdatePkgin() int {
	for _, line := range strings.Split(runPkginUpgradeCmd(), "\n") {
		if !strings.Contains(line, "to upgrade:") {
			continue
		}

		fields := strings.Fields(line)
		ret, err := strconv.Atoi(fields[0])

		if err == nil {
			return ret
		}
	}

	return -1
}

var runPkgListCmd = func() string {
	// For reasons I can't explain, this command exits 1 if there are no packages to upgrade.
	stdout, stderr, err := helpers.RunCmd("/bin/pkg list -uH")

	if err != nil && stderr != "no packages have newer versions available" {
		log.Print(err)
	}

	return stdout
}

var runPkginUpgradeCmd = func() string {
	stdout, stderr, err := helpers.RunCmd("echo n | /opt/local/bin/pkgin upgrade")

	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func havePkg5() bool {
	return haveFile(pkg5)
}

func havePkgin() bool {
	return haveFile(pkgin)
}

func haveFile(file string) bool {
	_, err := os.Stat(file)

	return err == nil
}

func init() {
	inputs.Add("illumos_patches", func() telegraf.Input { return &IllumosPatches{} })
}
