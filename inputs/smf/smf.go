package smf

import (
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## The service states you wish to count.
	# svc_states = ["online", "uninitialized", "degraded", "maintenance"]
	## The Zones you wish to examine. If this is unset or empty, all visible zones are counted.
	# zones = ["zone1", "zone2"]
	## Whether or not you wish to generate individual, detailed points for services which are in
	## SvcStates but are not "online"
	# generate_details = true
`

type IllumosSmf struct {
	SvcStates       []string
	Zones           []string
	GenerateDetails bool
}

type svcSummary struct {
	counts  svcCounts
	svcErrs svcErrs
}

type svcCounts map[string]zoneSvcSummary

type zoneSvcSummary map[string]int

type svcErrs []svcErr

type svcErr struct {
	zone  string
	state string
	fmri  string
}

const svcsCmd = "/bin/svcs -aHZ -ozone,state,fmri"

func (s *IllumosSmf) Description() string {
	return "Aggregates the states of SMF services across a host."
}

func (s *IllumosSmf) SampleConfig() string {
	return sampleConfig
}

var rawSvcsOutput = func() string {
	stdout, stderr, err := helpers.RunCmd(svcsCmd)

	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func (s *IllumosSmf) Gather(acc telegraf.Accumulator) error {
	data := parseSvcs(*s, rawSvcsOutput())

	for zone, stateCounts := range data.counts {
		for state, count := range stateCounts {
			acc.AddFields(
				"smf",
				map[string]interface{}{
					"states": count,
				},
				map[string]string{
					"zone":  zone,
					"state": state,
				},
			)
		}
	}

	for _, tags := range data.svcErrs {
		acc.AddFields(
			"smf",
			map[string]interface{}{
				"errors": 1,
			},
			map[string]string{
				"zone":  tags.zone,
				"state": tags.state,
				"fmri":  tags.fmri,
			},
		)
	}

	return nil
}

func parseSvcs(s IllumosSmf, raw string) svcSummary {
	ret := svcSummary{
		counts:  svcCounts{},
		svcErrs: svcErrs{},
	}

	for _, svcLine := range strings.Split(raw, "\n") {
		chunks := strings.Fields(svcLine)

		if len(chunks) != 3 {
			log.Printf("could not parse svc '%s'", svcLine)

			continue
		}

		zone, state, fmri := chunks[0], chunks[1], chunks[2]

		if !helpers.WeWant(zone, s.Zones) || !helpers.WeWant(state, s.SvcStates) {
			continue
		}

		_, zoneExists := ret.counts[zone]

		if !zoneExists {
			ret.counts[zone] = zoneSvcSummary{}
		}

		_, stateExists := ret.counts[zone][state]

		if !stateExists {
			ret.counts[zone][state] = 0
		}

		ret.counts[zone][state]++

		if s.GenerateDetails && state != "online" {
			ret.svcErrs = append(ret.svcErrs, svcErr{zone, state, fmri})
		}
	}

	return ret
}

func init() {
	inputs.Add("illumos_smf", func() telegraf.Input { return &IllumosSmf{} })
}
