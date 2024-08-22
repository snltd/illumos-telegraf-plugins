package smf

import (
	"fmt"
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
	## svc_states but are not "online"
	# generate_details = true
	## Use this command to get the elevated privileges svcs requires to observe other zones. 
	## Should be a path, like "/bin/sudo" "/bin/pfexec", but can also be "none", which will 
	## collect only the local zone.
	# elevate_privs_with = "/bin/sudo"
`

type IllumosSmf struct {
	SvcStates        []string
	Zones            []string
	GenerateDetails  bool
	ElevatePrivsWith string
}

type svcSummary struct {
	counts  svcCounts
	svcErrs svcErrs
}

type svcCounts map[string]zoneSvcSummary

type zoneSvcSummary map[string]int

// /bin/svcs -aHZ -ozone,state,fmri
// serv-wf          online         svc:/network/initial:default
//
// /bin/svcs -a -ozone,state,fmri
// global           online         svc:/network/physical:default

type svcErrs []svcErr

type svcErr struct {
	zone  string
	state string
	fmri  string
}

const (
	svcsCmdAllZones = "/bin/svcs -aHZ -ozone,state,fmri"
	svcsCmdThisZone = "/bin/svcs -aH -ozone,state,fmri"
)

func (s *IllumosSmf) Description() string {
	return "Reports the states of SMF services for a single zone or across a host."
}

func (s *IllumosSmf) SampleConfig() string {
	return sampleConfig
}

var rawSvcsOutput = func(s IllumosSmf) string {
	var svcsCmd string

	if s.ElevatePrivsWith == "none" {
		svcsCmd = svcsCmdThisZone
	} else {
		svcsCmd = fmt.Sprintf("%s %s", s.ElevatePrivsWith, svcsCmdAllZones)
	}

	stdout, stderr, err := helpers.RunCmd(svcsCmd)
	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func parseSvcs(s IllumosSmf, raw string) svcSummary {
	ret := svcSummary{
		counts:  svcCounts{},
		svcErrs: svcErrs{},
	}

	for _, svcLine := range strings.Split(strings.TrimSpace(raw), "\n") {
		chunks := strings.Fields(svcLine)

		if len(chunks) != 3 {
			log.Printf("Could not parse svc '%s'", svcLine)
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

func (s *IllumosSmf) Gather(acc telegraf.Accumulator) error {
	data := parseSvcs(*s, rawSvcsOutput(*s))

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

func init() {
	inputs.Add("illumos_smf", func() telegraf.Input { return &IllumosSmf{} })
}
