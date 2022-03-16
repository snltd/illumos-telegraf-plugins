package zpool

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

const timestampFormat = "Mon Jan 2 15:04:05 2006"

var sampleConfig = `
	## The metrics you wish to report. They can be any of the headers in the output of 'zpool list',
	## and also a numeric interpretation of 'health'.
	# fields = ["size", "alloc", "free", "cap", "dedup", "health"]
	## Status metrics are things like ongoing resilver time, ongoing scrub time, error counts
	## and whatnot
	# status = true
`

type IllumosZpool struct {
	Fields []string
	Status bool
}

type statusErrorCount struct {
	device      string
	state       string
	readErrors  float64
	writeErrors float64
	cksumErrors float64
}

func (s *IllumosZpool) Description() string {
	return "Reports the health and status of ZFS pools."
}

func (s *IllumosZpool) SampleConfig() string {
	return sampleConfig
}

var zpoolOutput = func() string {
	stdout, stderr, err := helpers.RunCmd("/usr/sbin/zpool list")

	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

var zpoolStatusOutput = func(pool string) string {
	stdout, stderr, err := helpers.RunCmd(fmt.Sprintf("/usr/sbin/zpool status -pv %s", pool))

	if err != nil {
		log.Print(stderr)
		log.Print(err)
	}

	return stdout
}

func (s *IllumosZpool) Gather(acc telegraf.Accumulator) error {
	raw := zpoolOutput()
	lines := strings.Split(raw, "\n")
	fields := make(map[string]interface{})

	for _, pool := range lines[1:] {
		poolStats := parseZpool(pool, lines[0])
		tags := map[string]string{"name": poolStats.name}

		for stat, val := range poolStats.props {
			if helpers.WeWant(stat, s.Fields) {
				fields[stat] = val
			}
		}

		acc.AddFields("zpool", fields, tags)

		if s.Status {
			statusOutput := zpoolStatusOutput(poolStats.name)

			statusFields := map[string]interface{}{
				"resilverTime":   resilverTime(statusOutput),
				"scrubTime":      scrubTime(statusOutput),
				"timeSinceScrub": timeSinceScrub(statusOutput),
			}

			acc.AddFields("zpool.status", statusFields, tags)

			errorCounts := extractErrorCounts(statusOutput)

			for _, errorCount := range errorCounts {
				errorTags := map[string]string{
					"pool":   poolStats.name,
					"device": errorCount.device,
					"state":  errorCount.state,
				}

				errorFields := map[string]interface{}{
					"read":  errorCount.readErrors,
					"write": errorCount.writeErrors,
					"cksum": errorCount.cksumErrors,
				}

				acc.AddFields("zpool.status.errors", errorFields, errorTags)
			}
		}

	}

	return nil
}

// parseHeader turns the first line of `zpool list`'s output into an array of lower-case strings.
func parseHeader(raw string) []string {
	return strings.Fields(strings.ToLower(raw))
}

// healthtoi converts the health of a zpool to an integer, so you can alert off it.
// 0 : ONLINE
// 1 : DEGRADED
// 2 : SUSPENDED
// 3 : UNAVAIL
// 4 : FAULTED
// 99: <cannot parse>
func healthtoi(health string) int {
	states := []string{"ONLINE", "DEGRADED", "SUSPENDED", "UNAVAIL", "FAULTED"}

	for i, state := range states {
		if state == health {
			return i
		}
	}

	return 99
}

// Zpool stores all the Zpool properties in the `props` map, which is dynamically generated. This
// means it will work on Solaris as well as Illumos, and won't break if the output format of
// `zpool(1m)` changes.
type Zpool struct {
	name  string
	props map[string]interface{}
}

// parseZpool semi-intelligently parses a line of `zpool list` output, using that command's output
// header to pick out the fields we are interested in.
func parseZpool(raw, rawHeader string) Zpool {
	fields := parseHeader(rawHeader)
	chunks := strings.Fields(raw)
	pool := Zpool{
		name:  chunks[0],
		props: make(map[string]interface{}),
	}

	for i, field := range chunks {
		property := fields[i]

		switch property {
		case "size":
			fallthrough
		case "alloc":
			fallthrough
		case "free":
			pool.props[property], _ = helpers.Bytify(field)
		case "frag":
			fallthrough
		case "cap":
			pool.props[property], _ = strconv.Atoi(strings.TrimSuffix(field, "%"))
		case "dedup":
			strval := strings.TrimSuffix(field, "x")
			pool.props["dedup"], _ = strconv.ParseFloat(strval, 64)
		case "health":
			pool.props["health"] = healthtoi(field)
		}
	}

	return pool
}

// resilverTime pulls the 'Sun Sep 12 15:11:35 2021' format timestamp out of `zpool status`, if
// it's there, and turns it into the number of seconds from then to now. If there's no resilver in
// progress, returns 0
func resilverTime(zpoolStatusOutput string) float64 {
	return extractTime(zpoolStatusOutput, "resilver in progress since")
}

func timeSinceScrub(zpoolStatusOutput string) float64 {
	return extractTime(zpoolStatusOutput, "scrub repaired.*errors on")
}

func scrubTime(zpoolStatusOutput string) float64 {
	return extractTime(zpoolStatusOutput, "scrub in progress since")
}

// Timestamps crop up in `zpool status` output. If you supply a string preceding a timestamp,
// you'll get back the number of seconds since that timestamp. If there is no match, you get 0.
func extractTime(zpoolStatusOutput, keyPhrase string) float64 {
	rx := regexp.MustCompile(fmt.Sprintf("(?m)%s ([^\n]+)", keyPhrase))

	startTimeMatches := rx.FindStringSubmatch(zpoolStatusOutput)

	if len(startTimeMatches) == 0 {
		return 0
	}

	startTime, err := time.Parse(timestampFormat, startTimeMatches[1])

	if err != nil {
		return 0
	}

	return time.Since(startTime).Seconds()
}

func extractErrorCounts(statusOutput string) []statusErrorCount {
	ret := []statusErrorCount{}

	rx := regexp.MustCompile("(?s)config:(.*)errors:")

	blockMatches := rx.FindStringSubmatch(statusOutput)

	if len(blockMatches) == 0 {
		return ret
	}

	block := strings.TrimSpace(blockMatches[1])
	lines := strings.Split(block, "\n")

	for i, line := range lines {
		if i == 0 {
			continue
		}

		fields := strings.Fields(line)

		readErrors, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Print("cannot parse zpool status read error counts")
			continue
		}

		writeErrors, err := strconv.Atoi(fields[3])
		if err != nil {
			log.Print("cannot parse zpool status write error counts")
			continue
		}

		cksumErrors, err := strconv.Atoi(fields[4])
		if err != nil {
			log.Print("cannot parse zpool status cksum error counts")
			continue
		}

		ret = append(ret, statusErrorCount{
			device:      fields[0],
			state:       fields[1],
			readErrors:  float64(readErrors),
			writeErrors: float64(writeErrors),
			cksumErrors: float64(cksumErrors),
		})
	}

	return ret
}

func init() {
	inputs.Add("illumos_zpool", func() telegraf.Input { return &IllumosZpool{} })
}
