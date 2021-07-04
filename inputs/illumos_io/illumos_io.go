package io

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/illumos/go-kstat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
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

func extractFields(s *IllumosIO, stat *kstat.IO) map[string]interface{} { //nolint:cyclop
	fields := make(map[string]interface{})

	if helpers.WeWant("nread", s.Fields) {
		fields["nread"] = float64(stat.Nread)
	}

	if helpers.WeWant("nwritten", s.Fields) {
		fields["nwritten"] = float64(stat.Nwritten)
	}

	if helpers.WeWant("reads", s.Fields) {
		fields["reads"] = float64(stat.Writes)
	}

	if helpers.WeWant("writes", s.Fields) {
		fields["writes"] = float64(stat.Writes)
	}

	if helpers.WeWant("wtime", s.Fields) {
		fields["wtime"] = float64(stat.Wtime)
	}

	if helpers.WeWant("wlentime", s.Fields) {
		fields["wlentime"] = float64(stat.Wlentime)
	}

	if helpers.WeWant("wlastupdate", s.Fields) {
		fields["wlastupdate"] = float64(stat.Wlastupdate)
	}

	if helpers.WeWant("rtime", s.Fields) {
		fields["rtime"] = float64(stat.Rtime)
	}

	if helpers.WeWant("rlentime", s.Fields) {
		fields["rlentime"] = float64(stat.Rlentime)
	}

	if helpers.WeWant("rlastupdate", s.Fields) {
		fields["rlastupdate"] = float64(stat.Wlastupdate)
	}

	if helpers.WeWant("wcnt", s.Fields) {
		fields["wcnt"] = float64(stat.Wcnt)
	}

	if helpers.WeWant("rcnt", s.Fields) {
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
		tags["serialNo"] = helpers.NamedValue(serialNo).(string)
	}

	product, err := token.GetNamed("sderr", instance, name, "Product")

	if err == nil {
		tags["product"] = helpers.NamedValue(product).(string)
	}

	return tags
}

func (s *IllumosIO) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Print("cannot get kstat token")

		return err
	}

	defer token.Close()

	ioStats := helpers.KStatIoClass(token, "disk")

	for modName, stat := range ioStats {
		chunks := strings.Split(modName, ":")
		mod := chunks[0]
		name := chunks[1]

		if !helpers.WeWant(mod, s.Modules) || !helpers.WeWant(name, s.Devices) {
			continue
		}

		acc.AddFields(
			"io",
			extractFields(s, stat),
			createTags(token, mod, name),
		)
	}

	return nil
}

func init() {
	inputs.Add("illumos_io", func() telegraf.Input { return &IllumosIO{} })
}
