package fma

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## Whether to report fmstat(1m) metrics
	# fmstat = true
	## Which fmstat modules to report
	# fmstat_modules = []
	## Which fmstat fields to report
	# fmstat_fields = []
	## Whether to report fmadm(1m) metrics
	# fmadm = true
	## Use this command to get elevated privileges required to run fmadm. 
	## Should be a path, like "/bin/sudo" "/bin/pfexec", but can also be "none", which will 
	## omit the fmadm collection.
	# elevate_privs_with = "/bin/sudo"
`

type IllumosFma struct {
	Fmstat           bool
	FmstatModules    []string
	FmstatFields     []string
	Fmadm            bool
	ElevatePrivsWith string
}

type Fmstat struct {
	module string
	props  map[string]float64
}

func (s *IllumosFma) Description() string {
	return `A vague, experimental collector for the Illumos fault management architecture. I'm not
	sure yet what it is worth recording, and how, so this is almost certainly subject to change`
}

func (s *IllumosFma) SampleConfig() string {
	return sampleConfig
}

var runFmstatCmd = func() string {
	stdout, stderr, err := helpers.RunCmd("/usr/sbin/fmstat")
	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

var runFmadmFaultyCmd = func(cmdPrefix string) string {
	stdout, stderr, err := helpers.RunCmd(
		fmt.Sprintf("%s /usr/sbin/fmadm faulty -arf", cmdPrefix))
	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func gatherFmstat(s *IllumosFma, acc telegraf.Accumulator) {
	raw := strings.Split(strings.TrimSpace(runFmstatCmd()), "\n")
	header := parseFmstatHeader(raw[0])

	for _, statLine := range raw[1:] {
		fields := make(map[string]interface{})
		fmstats := parseFmstatLine(statLine, header)

		if !helpers.WeWant(fmstats.module, s.FmstatModules) {
			continue
		}

		for stat, val := range fmstats.props {
			if helpers.WeWant(stat, s.FmstatFields) {
				fields[stat] = val
			}
		}

		acc.AddFields("fma.fmstat", fields, map[string]string{"module": fmstats.module})
	}
}

// I originally wrote this module for Solaris, which provides significantly more information on
// fmadm faults than Illumos does. Lacking that, I've fallen back to the short form of `fmadm
// faulty -arf`. This outputs one fault per line, of the form
//
//	zfs://pool=big/vdev=3706b5d93e20f727                                  faulted
//
// Problem is, I've only seen that one fault since I wrote this plugin, plus another example I
// found in the Illumos source! So this function, and
// probably the rest of this plugin, could be subject to severe revision.
// So far as I can tell, impacts are unique, so we send a value of "1" for every impact.
func gatherFmadm(acc telegraf.Accumulator, cmdPrefix string) {
	raw := runFmadmFaultyCmd(cmdPrefix)

	for _, impact := range strings.Split(raw, "\n") {
		if strings.Contains(impact, "://") {
			acc.AddFields("fma.fmadm",
				map[string]interface{}{"faults": 1},
				parseFmadmImpact(impact))
		}
	}
}

func parseFmstatHeader(headerLine string) []string {
	return strings.Fields(strings.ReplaceAll(headerLine, "%", "pc_"))
}

func parseFmstatLine(fmstatLine string, header []string) Fmstat {
	fields := strings.Fields(fmstatLine)
	props := make(map[string]float64)

	for i, field := range fields {
		property := header[i]

		switch property {
		case "module":
		case "memsz":
			fallthrough
		case "bufsz":
			props[property], _ = helpers.Bytify(field)
		default:
			props[property], _ = strconv.ParseFloat(field, 64)
		}
	}

	return Fmstat{
		module: fields[0],
		props:  props,
	}
}

func parseFmadmImpact(impact string) map[string]string {
	ret := make(map[string]string)

	fields := strings.Fields(impact)
	ret["status"] = fields[1]

	parts := strings.Split(fields[0], "://")
	ret["module"] = parts[0]

	for _, chunk := range strings.Split(parts[1], "/") {
		if strings.Contains(chunk, "=") {
			bits := strings.Split(chunk, "=")
			ret[bits[0]] = bits[1]
		}
	}

	return ret
}

func (s *IllumosFma) Gather(acc telegraf.Accumulator) error {
	// There's no error handling in here. I'm not really sure what errors we might need to handle,
	// so if this ever gets used, it will need improvement.
	if s.Fmstat {
		gatherFmstat(s, acc)
	}

	if s.Fmadm && s.ElevatePrivsWith != "none" {
		gatherFmadm(acc, s.ElevatePrivsWith)
	}

	return nil
}

func init() {
	inputs.Add("illumos_fma", func() telegraf.Input { return &IllumosFma{} })
}
