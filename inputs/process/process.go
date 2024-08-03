/*
 * This collector sends everything as a gauge, so it's up to you and your
 * graphing software to convert them into meaningful rates.
 * It's work-in-progress, rough-and-ready, there to do a quick job.
 *
 * If you want to add more tags, like project ID or something, they
 * need to go in the procDigest struct.
 */

package process

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
)

var sampleConfig = `
	## A list of the kstat values you wish to turn into metrics. Each value
	## will create a new timeseries. Look at the plugin source for a full
	## list of values.
	# Values = ["rtime", "rsssize", "inblk", "oublk", "prtcpu", "prtmem"]
	## Tags you wish to be attached to ALL metrics. Again, see source for 
	## all your options.
	# Tags = ["name", "zoneid", "uid", "contract"]
	## How many processes to send metrics for. You get this many process for 
	## EACH of the Values you listed above. Don't set it to zero.
	# TopK = 10
	## It's slightly expensive, but we can expand zone IDs and contract IDs
	## to zone names and service names.
	# ExpandZoneTag = true
	# ExpandContractTag = true
`

func (s *IllumosProcess) Description() string {
	return "Reports on illumos processes, like prstat(1)"
}

func (s *IllumosProcess) SampleConfig() string {
	return sampleConfig
}

type IllumosProcess struct {
	Values            []string
	Tags              []string
	TopK              int
	ExpandZoneTag     bool
	ExpandContractTag bool
}

type procObject struct {
	Values procObjectValues
	Tags   procObjectTags
}

type (
	contractMap      map[id_t]string
	procObjectValues map[string]interface{}
	procObjectTags   map[string]string
	procObjectMap    map[int]procObject
)

var procRootDir = "/proc"

func toNs(ts timestruc_t) int64 {
	return ts[0]*1e9 + ts[1]
}

func newProcObject(s *IllumosProcess, pid int) (procObject, error) {
	psinfo, err := loadProcPsinfo(pid)
	if err != nil {
		return procObject{}, err
	}

	prusage, err := loadProcUsage(pid)
	if err != nil {
		return procObject{}, err
	}

	// This is my selection of prusage_t fields which I think might be useful.
	// If you want any that aren't here, add them and recompile.

	values := procObjectValues{}

	if helpers.WeWant("rtime", s.Values) {
		values["rtime"] = toNs(prusage.Pr_rtime)
	}

	if helpers.WeWant("utime", s.Values) {
		values["utime"] = toNs(prusage.Pr_utime)
	}

	if helpers.WeWant("stime", s.Values) {
		values["stime"] = toNs(prusage.Pr_stime)
	}

	if helpers.WeWant("wtime", s.Values) {
		values["wtime"] = toNs(prusage.Pr_wtime)
	}

	if helpers.WeWant("inblk", s.Values) {
		values["inblk"] = int64(prusage.Pr_inblk)
	}

	if helpers.WeWant("oublk", s.Values) {
		values["oublk"] = int64(prusage.Pr_oublk)
	}

	if helpers.WeWant("sysc", s.Values) {
		values["sysc"] = int64(prusage.Pr_sysc)
	}

	if helpers.WeWant("ioch", s.Values) {
		values["ioch"] = int64(prusage.Pr_ioch)
	}

	if helpers.WeWant("size", s.Values) { // kb to b
		values["size"] = int64(psinfo.Pr_size) * 1024
	}

	if helpers.WeWant("rssize", s.Values) { // kb to b
		values["rssize"] = int64(psinfo.Pr_rssize) * 1024
	}

	if helpers.WeWant("pctcpu", s.Values) { // div by 10,000 for actual %age
		values["pctcpu"] = int64(psinfo.Pr_pctcpu)
	}

	if helpers.WeWant("pctmem", s.Values) { // div by 10,000 for actual %age
		values["pctmem"] = int64(psinfo.Pr_pctmem)
	}

	if helpers.WeWant("nlwp", s.Values) {
		values["nlwp"] = int64(psinfo.Pr_nlwp)
	}

	if helpers.WeWant("count", s.Values) {
		values["count"] = int64(prusage.Pr_count)
	}

	// Another educated guess, this time at things which might be useful as
	// point tags.

	tags := procObjectTags{}

	if helpers.WeWant("name", s.Tags) {
		tags["name"] = string(bytes.TrimRight(psinfo.Pr_fname[:], "\x00"))
	}

	if helpers.WeWant("uid", s.Tags) {
		tags["uid"] = fmt.Sprint(psinfo.Pr_uid)
	}

	if helpers.WeWant("gid", s.Tags) {
		tags["gid"] = fmt.Sprint(psinfo.Pr_gid)
	}

	if helpers.WeWant("euid", s.Tags) {
		tags["euid"] = fmt.Sprint(psinfo.Pr_euid)
	}

	if helpers.WeWant("egid", s.Tags) {
		tags["egid"] = fmt.Sprint(psinfo.Pr_egid)
	}

	if helpers.WeWant("taskid", s.Tags) {
		tags["taskid"] = fmt.Sprint(psinfo.Pr_taskid)
	}

	if helpers.WeWant("projid", s.Tags) {
		tags["projid"] = fmt.Sprint(psinfo.Pr_projid)
	}

	if helpers.WeWant("zoneid", s.Tags) {
		tags["zoneid"] = fmt.Sprint(psinfo.Pr_zoneid)
	}

	if helpers.WeWant("contract", s.Tags) {
		tags["contract"] = fmt.Sprint(psinfo.Pr_contract)
	}

	if helpers.WeWant("pid", s.Tags) {
		tags["pid"] = fmt.Sprint(psinfo.Pr_pid)
	}

	if helpers.WeWant("ppid", s.Tags) {
		tags["ppid"] = fmt.Sprint(psinfo.Pr_ppid)
	}

	return procObject{Values: values, Tags: tags}, nil
}

func newProcObjectMap(s *IllumosProcess, procs []fs.DirEntry) procObjectMap {
	ret := procObjectMap{}

	for _, proc := range procs {
		pid, err := strconv.Atoi(proc.Name())
		if err != nil {
			log.Printf("cannot process pid %v", proc)
			continue
		}

		ret[pid], _ = newProcObject(s, pid)
	}

	return ret
}

func newContractMap(svcsOutput string) contractMap {
	svcs := strings.Split(svcsOutput, "\n")
	ret := make(contractMap)

	for _, svc := range svcs {
		fields := strings.Fields(svc)

		if len(fields) != 2 || fields[0] == "-" {
			continue
		}

		ctid, err := strconv.Atoi(fields[0])
		svcFmri := fields[1]

		if err != nil {
			log.Printf("error converting %v", fields)
			continue
		}

		ret[id_t(ctid)] = svcFmri
	}

	return ret
}

func allProcs() []fs.DirEntry {
	procs, err := os.ReadDir(procRootDir)
	if err != nil {
		log.Fatalf("cannot read %s", procRootDir)
	}

	return procs
}

func expandZoneTag(processMap *procObjectMap, zoneMap helpers.ZoneMap) {
	for _, procObj := range *processMap {
		zid, ok := procObj.Tags["zoneid"]

		if ok {
			zoneId, err := strconv.Atoi(zid)
			if err == nil {
				zone, err := zoneMap.ZoneByID(zoneId)
				if err == nil {
					procObj.Tags["zone"] = string(zone.Name)
				}
			}
		}
	}
}

func expandContractTag(processMap *procObjectMap, contractMap contractMap) {
	for _, procObj := range *processMap {
		ctid, ok := procObj.Tags["contract"]

		// If we can't find a contract with the given ID, we don't add a tag.
		if ok {
			ctid, err := strconv.Atoi(ctid)
			if err == nil {
				svc, ok := contractMap[id_t(ctid)]
				if ok {
					procObj.Tags["service"] = svc
				}
			}
		}
	}
}

type SortPair struct {
	Pid   int
	Value int64
}

func topKPids(processMap *procObjectMap, field string, topK int) []int {
	pairs := make([]SortPair, len(*processMap))
	i := 0

	for pid, procObj := range *processMap {
		value, ok := procObj.Values[field]

		if ok {
			pairs[i] = SortPair{pid, value.(int64)}
			i++
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	ret := make([]int, topK)
	i = 0

	for _, pair := range pairs[0:topK] {
		ret[i] = pair.Pid
		i++
	}

	return ret
}

func (s *IllumosProcess) Gather(acc telegraf.Accumulator) error {
	processMap := newProcObjectMap(s, allProcs())

	if s.ExpandZoneTag {
		expandZoneTag(&processMap, newZoneMap())
	}

	if s.ExpandContractTag {
		svcsOutput, err := collectContractInfo()
		if err == nil {
			expandContractTag(&processMap, newContractMap(svcsOutput))
		}
	}

	for _, field := range s.Values {
		pidList := topKPids(&processMap, field, s.TopK)

		for _, pid := range pidList {
			procObj, ok := processMap[pid]
			if ok {
				fields := make(map[string]interface{})
				val, ok := procObj.Values[field]
				if ok {
					fields[field] = val.(int64)
					acc.AddFields("process", fields, procObj.Tags)
				}
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("illumos_process", func() telegraf.Input { return &IllumosProcess{} })
}

// Functions below here are vars so they can be injected by the tests.

var collectContractInfo = func() (string, error) {
	svcsOutput, svcsErr, err := helpers.RunCmd("/bin/svcs -vHo ctid,fmri")

	if err != nil {
		log.Printf("cannot create contract map. No svc info: %s\n", svcsErr)
		return "", errors.New("failed to collect svcs info")
	} else {
		return svcsOutput, nil
	}
}

var newZoneMap = func() helpers.ZoneMap {
	return helpers.NewZoneMap()
}

var loadProcUsage = func(pid int) (prusage_t, error) {
	file := fmt.Sprintf("/proc/%d/usage", pid)
	var prusage prusage_t
	fh, err := os.Open(file)
	// Often the process is gone before we get to inspect it. Don't bother
	// logging an error.
	if err != nil {
		return prusage, err
	}

	err = binary.Read(fh, binary.LittleEndian, &prusage)
	if err != nil {
		log.Printf("cannot parse %s\n", file)
		return prusage, err
	}

	return prusage, err
}

var loadProcPsinfo = func(pid int) (psinfo_t, error) {
	file := path.Join(procRootDir, fmt.Sprint(pid), "psinfo")
	var psinfo psinfo_t
	fh, err := os.Open(file)
	if err != nil {
		return psinfo, err
	}

	err = binary.Read(fh, binary.LittleEndian, &psinfo)
	if err != nil {
		log.Printf("cannot parse %s\n", file)
		return psinfo, err
	}

	return psinfo, nil
}
