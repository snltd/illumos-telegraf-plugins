package helpers

import (
	"fmt"
	kstat "github.com/illumos/go-kstat"
	"log"
)

// pulled out into a function to facilitate testing.
var allKStats = func(token *kstat.Token) []*kstat.KStat {
	return token.All()
}

// NamedValue returns the useable value of the given named kstat. If said value is numeric, it is
// sent as a float64, which is what Telegraf expects as a value.
func NamedValue(stat *kstat.Named) interface{} {
	switch stat.Type.String() {
	case "string", "char":
		return stat.StringVal
	case "int32", "int64":
		return float64(stat.IntVal)
	case "uint32", "uint64":
		return float64(stat.UintVal)
	default:
		log.Fatalf("%s is of type %s", stat.Name, stat.Type)
	}

	return nil
}

// KStatIoClass returns a map of module:name => kstat for IO kstats.
func KStatIoClass(token *kstat.Token, class string) map[string]*kstat.IO {
	ret := make(map[string]*kstat.IO)

	for _, n := range token.All() {
		if n.Class != class {
			continue
		}

		stat, err := n.GetIO()
		if err != nil {
			log.Fatal("cannot get kstat")
		}

		ret[fmt.Sprintf("%s:%s", n.Module, n.Name)] = stat
	}

	return ret
}

// KStatsInClass returns a list of kstats in the given class.
func KStatsInClass(token *kstat.Token, class string) []*kstat.KStat {
	var ret []*kstat.KStat

	for _, stat := range allKStats(token) {
		if stat.Class == class {
			ret = append(ret, stat)
		}
	}

	return ret
}

// KStatsInModule returns a list of kstats in the given module. Asking for the 'cpu' module would
// give you something like:
// cpu:0:intrstat
// cpu:0:sys
// cpu:0:vm
// cpu:1:intrstat
// ...
func KStatsInModule(token *kstat.Token, module string) []*kstat.KStat {
	var ret []*kstat.KStat

	for _, stat := range allKStats(token) {
		if stat.Module == module {
			ret = append(ret, stat)
		}
	}

	return ret
}
