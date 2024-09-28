package os

/*
Inspired by, that is "ripped off from" the node_exporter 'os' output.
*/

import (
	"log"
	"os"
	"slices"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	Outputs 1, but with tags that can be combined with other metrics.
`

var osRelease = "/etc/os-release"

func (s *IllumosOS) Description() string {
	return "Reports illumos operating system information"
}

func (s *IllumosOS) SampleConfig() string {
	return sampleConfig
}

type IllumosOS struct{}

var runUnameCmd = func() string {
	stdout, stderr, err := helpers.RunCmd("/bin/uname -v")
	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func parseOsRelease(osRelease string) map[string]string {
	ret := map[string]string{}
	lines := strings.Split(strings.TrimSpace(osRelease), "\n")
	required := []string{"name", "version", "build_id"}

	for _, line := range lines {
		chunks := strings.SplitN(line, "=", 2)

		if len(chunks) == 2 {
			key := strings.ToLower(chunks[0])
			val := strings.ReplaceAll(chunks[1], "\"", "")

			if slices.Contains(required, key) {
				ret[key] = val
			}
		}
	}

	return ret
}

func (s *IllumosOS) Gather(acc telegraf.Accumulator) error {
	osReleaseContents, err := os.ReadFile(osRelease)
	if err != nil {
		return err
	}

	tags := parseOsRelease(string(osReleaseContents))
	tags["kernel"] = runUnameCmd()

	acc.AddFields("os", map[string]interface{}{"release": 1}, tags)

	return nil
}

func init() {
	inputs.Add("illumos_os", func() telegraf.Input { return &IllumosOS{} })
}
