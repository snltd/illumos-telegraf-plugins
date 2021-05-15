package illumos_io

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

var sampleConfig = `
	## The kstat fields you wish to emit. 'kstat -c disk' will show what is collected. Not defining
	## any fields sends everything, which is probably not what you want.
	# fields = ["reads", "nread", "writes", "nwritten"]
	## Report on the following kstat modules. You likely have 'sd' and 'zfs'. Specifying none
	## reports on all.
	# modules = ["sd", "zfs"]
	## Report on the following devices, inside the above modules. Specifying none reports on all.
	# devices = ["sd0"]
`

func (s *IllumosIO) Description() string {
	return "Reports on Illumos IO"
}

func (s *IllumosIO) SampleConfig() string {
	return sampleConfig
}

type IllumosIO struct {
	Devices []string
	Fields  []string
	Modules []string
}

func extractFields(s *IllumosIO, stat *kstat.IO) map[string]interface{} {
	fields := make(map[string]interface{})

	if sth.WeWant("nread", s.Fields) {
		fields["nread"] = float64(stat.Nread)
	}

	if sth.WeWant("nwritten", s.Fields) {
		fields["nwritten"] = float64(stat.Nwritten)
	}

	if sth.WeWant("reads", s.Fields) {
		fields["reads"] = float64(stat.Writes)
	}

	if sth.WeWant("writes", s.Fields) {
		fields["writes"] = float64(stat.Writes)
	}

	if sth.WeWant("wtime", s.Fields) {
		fields["wtime"] = float64(stat.Wtime)
	}

	if sth.WeWant("wlentime", s.Fields) {
		fields["wlentime"] = float64(stat.Wlentime)
	}

	if sth.WeWant("wlastupdate", s.Fields) {
		fields["wlastupdate"] = float64(stat.Wlastupdate)
	}

	if sth.WeWant("rtime", s.Fields) {
		fields["rtime"] = float64(stat.Rtime)
	}

	if sth.WeWant("rlentime", s.Fields) {
		fields["rlentime"] = float64(stat.Rlentime)
	}

	if sth.WeWant("rlastupdate", s.Fields) {
		fields["rlastupdate"] = float64(stat.Wlastupdate)
	}

	if sth.WeWant("wcnt", s.Fields) {
		fields["wcnt"] = float64(stat.Wcnt)
	}

	if sth.WeWant("rcnt", s.Fields) {
		fields["rcnt"] = float64(stat.Rcnt)
	}

	return fields
}

func createTags(token *kstat.Token, mod, device string) map[string]string {
	tags := map[string]string{
		"module": mod,
		"device": device,
	}

	deviceRegex := regexp.MustCompile("[0-9]+$")
	instance, err := strconv.Atoi(deviceRegex.FindString(device))

	if err != nil {
		return tags
	}

	name := fmt.Sprintf("%s,err", device)

	serialNo, err := token.GetNamed("sderr", instance, name, "Serial No")

	if err == nil {
		tags["serialNo"] = sth.NamedValue(serialNo).(string)
	}

	product, err := token.GetNamed("sderr", instance, name, "Product")

	if err == nil {
		tags["product"] = sth.NamedValue(product).(string)
	}

	return tags
}

func (s *IllumosIO) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	ioStats := sth.KStatIoClass(token, "disk")

	for modName, stat := range ioStats {
		chunks := strings.Split(modName, ":")
		mod := chunks[0]
		name := chunks[1]

		if !sth.WeWant(mod, s.Modules) || !sth.WeWant(name, s.Devices) {
			continue
		}

		acc.AddFields(
			"io",
			extractFields(s, stat),
			createTags(token, mod, name),
		)
	}

	token.Close()
	return nil
}

func init() {
	inputs.Add("illumos_io", func() telegraf.Input { return &IllumosIO{} })
}
