package illumos_proc

import (
	"encoding/binary"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	//sh "github.com/snltd/solaris-telegraf-helpers"
	//"io/ioutil"
	//"github.com/fatih/structs"
	"log"
	"os"
	//"sort"
	"strconv"
	"strings"
)

var sampleConfig = `
	## Everything in 'Fields' will create a new metric path, like "proc.<execname>.<field>".
	# fields = ["size", "rssize"]
	## How many processes to send metrics for.
	# top_n = 10
	## Which tags to apply. Some, like the SMF service, are a little expensive.
	# tags = ["name", "pid", "zone", "svc"]
`

func (s *IllumosProc) Description() string {
	return "Reports on Illumos processes, like prstat(1)"
}

func (s *IllumosProc) SampleConfig() string {
	return sampleConfig
}

type IllumosProc struct {
	//Fields []string
	//Tags   []string
	//TopN   int
}

//type procItems map[string]interface{}

var procDir = "/proc" // a var so it can be overridden to use fixtures in testing

func gatherProcInfo(pidList []int) {
	for _, pid := range pidList {
		fmt.Println(pid)
		//x, _ := readProcPsinfo(pid)
		//fmt.Printf("%#v\n", x)
		//ioutil.WriteFile(fmt.Sprintf("%d.psinfo", pid), x, 0o644)
	}
}

func (s *IllumosProc) Gather(acc telegraf.Accumulator) error {
	gatherProcInfo(procPidList())

	// get a list of all proc IDs

	// make a list of all procs from that list

	// sort the procs

	// turn the top 'n' into metrics

	//all_procs := allProcs()
	//var contract_map map[int]string

	/*
		if sh.WeWant("svc", s.Tags) {
			contract_map = contractMap(runSvcsCtidCmd())
		}

		for _, field := range s.Fields {
			raw_field := "Pr_" + field
			procs := leaderboard(all_procs, raw_field, s.TopN)

			for _, proc := range procs {
				metrics := make(map[string]interface{})
				tags := make(map[string]string)

				if sh.WeWant("zone", s.Tags) {
					tags["zone"] = proc.zone
				}

				if sh.WeWant("pid", s.Tags) {
					tags["pid"] = strconv.Itoa(proc.pid)
				}

				if sh.WeWant("name", s.Tags) {
					tags["name"] = proc.name
				}

				if sh.WeWant("svc", s.Tags) {
					tags["svc"] = contract_map[int(proc.ctid)]
				}

				metrics[field] = proc.value
				acc.AddFields("solaris_proc", metrics, tags)
			}
		}
	*/

	return nil
}

/*
type procDigest struct {
	pid   int    // process ID
	name  string // exec name
	value int64  // value of counter
	zone  string // zone name
	ctid  id_t   // contract ID (for SMF service name lookup)
	ts    int64  // nanosecond time stamp, relative to global boot
}

var zoneMap sh.ZoneMap

// leaderboard() needs these procDigest structs and funcs to sort
type procDigests []procDigest

func (d procDigests) Len() int           { return len(d) }
func (d procDigests) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d procDigests) Less(i, j int) bool { return d[i].value < d[j].value }
*/

/////////////////////////////////

/*
func allProcs() map[int]procItems {
	procs := procPidList()
	ret := make(map[int]procItems)

	for _, pid := range procs {
		psinfo, info_err := procPsinfo(pid)

		if info_err == nil {
			m := structs.Map(psinfo)

			usage, usage_err := procUsage(pid)

			if usage_err == nil {
				n := structs.Map(usage)

				for k, v := range n {
					m[k] = v
				}

				ret[pid] = m
			}
		}
	}
	return ret
}
*/

/*
var runSvcsCtidCmd = func() string {
	return sh.RunCmd("/bin/svcs -vHo ctid,fmri")
}

// Returns a list the top 'n' processes, sorted on the field you
// specify
//
func leaderboard(procs map[int]procItems, field string,
	limit int) procDigests {
	var to_sort procDigests

	for pid, vals := range procs {
		// convert the exec name from a byte array
		raw_name := vals["Pr_fname"].([16]byte)
		name := strings.TrimRight(string(raw_name[:]), "\x00")

		// convert timestruc_t into a straight nanosecond value. It'll
		// be easier to work with elsewhere

		raw_ts := vals["Pr_tstamp"].(timestruc_t)
		ts := raw_ts[0]*1e9 + raw_ts[1]

		c := procDigest{
			pid:   pid,
			name:  name,
			value: int64(vals[field].(size_t)),
			zone:  zoneLookup(vals["Pr_zoneid"].(id_t)),
			ctid:  vals["Pr_contract"].(id_t),
			ts:    ts}

		to_sort = append(to_sort, c)
	}

	sort.Sort(procDigests(to_sort))
	sort.Sort(sort.Reverse(to_sort))

	return to_sort[:limit]
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

func parseProcData(psinfo psinfo_t, usage prusage_t) (map[string]interface{}, map[string]string) {
	fields := map[string]interface{}{
		"size":       float64(psinfo.Pr_size * 1024),   // convert from kb to bytes
		"rss":        float64(psinfo.Pr_rssize * 1024), // convert from kb to bytes
		"percentCPU": bpcToPc(psinfo.Pr_pctcpu),
		"percentMem": bpcToPc(psinfo.Pr_pctmem),
	}

	fmt.Printf("--------> %#v\n", usage)

	tags := map[string]string{
		"execname":   strings.Trim(fmt.Sprintf("%s", psinfo.Pr_fname), "\x00"),
		"args":       strings.Trim(fmt.Sprintf("%s", psinfo.Pr_psargs), "\x00"),
		"pid":        fmt.Sprintf("%d", psinfo.Pr_pid),
		"uid":        fmt.Sprintf("%d", psinfo.Pr_uid),
		"gid":        fmt.Sprintf("%d", psinfo.Pr_gid),
		"zoneID":     fmt.Sprintf("%d", psinfo.Pr_zoneid),
		"contractID": fmt.Sprintf("%d", psinfo.Pr_contract),
	}

	return fields, tags
}

// Convert one of psinfo's weird "binary fraction" percentage values to an actual percentage
// value. Method is copied from prstat(1).
func bpcToPc(binaryFraction ushort_t) float64 {
	return float64(binaryFraction) * 100 / 0x8000
}

// timestruc_t is an array of [seconds, nanoseconds]. Combine and convert into seconds
func tsToSeconds(ts []int64) float64 {
	return float64(ts[0]*1e9+ts[1]) / 1e9
}

func init() {
	fmt.Println()
	//zoneMap = sh.NewZoneMap()
	inputs.Add("illumos_proc", func() telegraf.Input { return &IllumosProc{} })
}
