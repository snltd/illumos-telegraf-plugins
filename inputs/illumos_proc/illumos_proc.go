package illumos_proc

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/snltd/illumos-telegraf-plugins/helpers"
	"log"
	"os"
	"strconv"
	"strings"
)

var sampleConfig = `
	## Each field in top_fields will produce a new metric path under 'proc.top.<field>', which will
	## chart the top 'n' values for the given field.  Think prstat(1m).
	# top_fields = ["size", "rss"]
	## This specifies the value of 'n' in the above. Setting to 0 turns the feature off.
	# top_field_limit = 10
	## detailed_procs is a list of execnames which will be reported in detail. Metrics go under
	## 'proc.detail.<execname>.<field>'.  An empty list means no processes will be reported in
	## detail.
	# detailed_procs = []
	## detail_fields is a list of the proc fields you wish to chart. Each one gets its own metric
	## under 'proc.detail.<execname>'.
	# detail_fields = ["size", "rss"]
	## Which tags to apply, to all metrics. Some, like the SMF service, are a little expensive.
	# tags = ["name", "pid", "zone", "svc"]
`

func (s *IllumosProc) Description() string {
	return "Reports on Illumos processes, like prstat(1)"
}

func (s *IllumosProc) SampleConfig() string {
	return sampleConfig
}

type IllumosProc struct {
	TopFields     []string
	TopFieldLimit int
	DetailedProcs []string
	DetailFields  []string
	Tags          []string
}

//type procItems map[string]interface{}

var procDir = "/proc" // a var so it can be overridden to use fixtures in testing

type processDetail struct {
	Fields map[string]interface{} // ready to pass to Telegraf's accumulator
	Tags   map[string]string      // ditto
}

func findTopProcs(limit int, field string, proclist []processDetail, tags []string) []processDetail {
	return topProcesses(sortProcessList(proclist, field), limit)
}

func selectMetrics(procList []processDetail, desiredFields, desiredTags []string) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, proc := range procList {
		for field, value := range proc.Fields {
			if helpers.WeWant(field, desiredFields) {
				fields[field] = value
			}
		}

		for tag, value := range proc.Tags {
			if helpers.WeWant(tag, desiredTags) {
				tags[tag] = value
			}
		}
	}

	return fields, tags
}

func findDetailedProcs(desiredProcs []string, procList []processDetail) []processDetail {
	var ret []processDetail

	for _, proc := range procList {
		if helpers.WeWant(proc.Tags["execname"], desiredProcs) {
			ret = append(ret, proc)
		}
	}

	return ret
}

func (s *IllumosProc) Gather(acc telegraf.Accumulator) error {
	procList := processDetails(procPidList())

	if s.TopFieldLimit > 0 {
		for _, field := range s.TopFields {
			topProcList := findTopProcs(s.TopFieldLimit, field, procList, s.Tags)
			fields, tags := selectMetrics(topProcList, s.TopFields, s.Tags)

			acc.AddFields(fmt.Sprintf("proc.top.%s", field), fields, tags)
		}
	}

	if len(s.DetailedProcs) > 0 {
		for _, execname := range s.DetailedProcs {
			detailProcList := findDetailedProcs(s.DetailedProcs, procList)
			fields, tags := selectMetrics(detailProcList, s.DetailFields, s.Tags)

			acc.AddFields(fmt.Sprintf("proc.detail.%s", execname), fields, tags)
		}
	}

	return nil
}

func topProcesses(procList []processDetail, n int) []processDetail {
	return procList[0:n]
}

func sortProcessList(procList []processDetail, key string) []processDetail {
	for _, proc := range procList {
		if _, ok := proc.Fields[key]; !ok {
			log.Printf("cannot find proc key %s", key)
			return procList
		}
	}

	sort.Slice(procList, func(i, j int) bool {
		return procList[i].Fields[key].(float64) > procList[j].Fields[key].(float64)
	})

	return procList
}

/*
var zoneMap sh.ZoneMap

*/

/*
var runSvcsCtidCmd = func() string {
	return sh.RunCmd("/bin/svcs -vHo ctid,fmri")
}

func zoneLookup(zid id_t) string {
	zone, _ := zoneMap.ZoneByID(int(zid))
	return zone.Name
}

// Return the service name associated with a contract Id. If there
// isn't one, it'll return the empty string, which is fine.
//
// func ctidToSvc(ctmap map[id_t]string, ctid id_t) string {
	// svc := ctmap[ctid]
	// return svc
// }

///////////////////////////////////////////////////////////////////////

func contractMap(svcsOutput string) map[int]string {
	ret := make(map[int]string)

	for _, svcLine := range strings.Split(svcsOutput, "\n") {
		fields := strings.Fields(svcLine)
		ctidStr := fields[0]
		svc := fields[1]

		if ctidStr == "-" {
			continue
		}

		ctid, err := strconv.Atoi(fields[0])

		if err == nil {
			ret[ctid] = svc
		}
	}

	return ret
}

*/

func procPidList() []int {
	procs, err := os.ReadDir(procDir)

	if err != nil {
		log.Fatal(fmt.Sprintf("cannot read %v", procDir))
	}

	var ret []int

	for _, proc := range procs {
		pid, _ := strconv.Atoi(proc.Name())
		ret = append(ret, pid)
	}

	return ret
}

// process wink in and out of existence all the time. I think that a process not being found isn't
// even worth mentioning, but we will return an error
func readProcPsinfo(pid int) (psinfo_t, error) {
	var psinfo psinfo_t
	file := fmt.Sprintf("%s/%d/psinfo", procDir, pid)
	fh, err := os.Open(file)

	if err != nil {
		return psinfo, err
	}

	err = binary.Read(fh, binary.LittleEndian, &psinfo)
	return psinfo, err
}

func readProcUsage(pid int) (prusage_t, error) {
	var prusage prusage_t
	file := fmt.Sprintf("%s/%d/usage", procDir, pid)
	fh, err := os.Open(file)

	if err != nil {
		return prusage, err
	}

	err = binary.Read(fh, binary.LittleEndian, &prusage)
	return prusage, err
}

// processDetails takes a list of PIDs and returns a list of processDetails, each containing
// everything we know about a process, ready to be turned into metrics
func processDetails(pidList []int) []processDetail {
	var procList []processDetail

	for _, pid := range pidList {
		psinfo, err := readProcPsinfo(pid)

		if err != nil {
			continue
		}

		usage, err := readProcUsage(pid)

		if err != nil {
			continue
		}

		procList = append(procList, parseProcData(psinfo, usage))
	}

	return procList
}

// Tage the structs we made from the psinfo and usage fields, and make two maps of fields and
// tags, ready for passing to Telegraf's accumulator. We'll filter on certain fields, but no more
// processing will be necessary.
//
// We only deal with a selection of the psinfo and usage structs. The ones chosen are the ones I
// think right now might be useful to me. Extend this if you need more info.
func parseProcData(psinfo psinfo_t, usage prusage_t) processDetail {
	fields := map[string]interface{}{
		"size":       float64(psinfo.Pr_size * 1024),   // convert from kb to bytes
		"rss":        float64(psinfo.Pr_rssize * 1024), // convert from kb to bytes
		"percentCPU": bpcToPc(psinfo.Pr_pctcpu),        // convert to percentage
		"percentMem": bpcToPc(psinfo.Pr_pctmem),        // convert to a percentage
		"userTime":   tsToSeconds(usage.Pr_utime),      // convert to seconds
		"systemTime": tsToSeconds(usage.Pr_stime),      // convert to seconds
		"waitTime":   tsToSeconds(usage.Pr_wtime),      // convert to seconds
	}

	tags := map[string]string{
		"execname":   strings.Trim(fmt.Sprintf("%s", psinfo.Pr_fname), "\x00"),
		"args":       strings.Trim(fmt.Sprintf("%s", psinfo.Pr_psargs), "\x00"),
		"pid":        fmt.Sprintf("%d", psinfo.Pr_pid),
		"uid":        fmt.Sprintf("%d", psinfo.Pr_uid),
		"gid":        fmt.Sprintf("%d", psinfo.Pr_gid),
		"zoneID":     fmt.Sprintf("%d", psinfo.Pr_zoneid),
		"contractID": fmt.Sprintf("%d", psinfo.Pr_contract),
	}

	return processDetail{fields, tags}
}

// Convert one of psinfo's weird "binary fraction" percentage values to an actual percentage
// value. Method is copied from prstat(1).
func bpcToPc(binaryFraction ushort_t) float64 {
	return float64(binaryFraction) * 100 / 0x8000
}

// timestruc_t is an array of [seconds, nanoseconds]. Combine and convert into seconds
func tsToSeconds(ts timestruc_t) float64 {
	return float64(ts[0]*1e9+ts[1]) / 1e9
}

func init() {
	fmt.Println()
	//zoneMap = sh.NewZoneMap()
	inputs.Add("illumos_proc", func() telegraf.Input { return &IllumosProc{} })
}
