# illumos Memory Input Plugin

Reports memory statistics on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports on illumos virtual and physical memory usage.
[[inputs.illumos_memory]]
  ## Whether to produce metrics from the output of 'swap -s'
  # swap_on = true
  ## And which fields to use. Specifying none implies all.
  # swap_fields = ["allocated", "reserved", "used", "available"]
  ## Whether to report "extra" fields, and which ones (kernel, arcsize, freelist)
  # extra_on = true
  # extra_fields = ["kernel", "arcsize", "freelist"]
  ## Whether to collect vminfo kstats, and which ones.
  # vminfo_on = true
  # vminfo_fields = ["freemem", "swap_alloc", "swap_avail", "swap_free", "swap_resv"]
  ## Whether to collect cpu::vm kstats
  # cpuvm_on =true
  # cpuvm_fields = ["pgin", "anonpgin", "pgpgin", "pgout", "anonpgout", "pgpgout"]
  ## Whether to aggregate cpuvm fields. (False sents a set of metrics for each vcpu)
  # cpuvm_aggregate = false
  ## Whether to collect zone memory_cap fields, and which ones
  # zone_memcap_on = true
  # zone_memcap_zones = []
  # zone_memcap_fields = ["physcap", "rss", "swap"]
```

### Metrics
- memory
  - fields:
    - arcsize (int, bytes)
    - freelist (int, bytes)
    - kernel (int, bytes)
- memory.swap
  - fields:
    - allocated (int, bytes)
    - available (int, bytes)
    - reserved (int, bytes)
    - used (int, bytes)
- memory.vminfo
  - fields:
    - selected by user
- memory.cpuVm
  - fields:
    - selected by user
- memory.zone
  - fields:
      - rss (int, bytes)
      - swap (int, bytes)
  - tags:
    - zone (string, name of zone)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Break down RSS usage by zone:

```
ts("memory.zone.rss")
```

Report paging events.

```
deriv(ts("memory.cpuVm.vm.aggregate.*pgin") 
deriv(ts("memory.cpuVm.vm.aggregate.*pgout")
```

### Example Output

```
> memory.swap,host=serv allocated=1083809792,available=3033550848,reserved=1210863616,used=2294673408 1727450040000000000
> memory,host=serv arcsize=11765448480,freelist=518955008,kernel=15435886592 1727450040000000000
> memory.vminfo,host=serv freemem=21610793312256,swapAlloc=25554691784704,swapAvail=89680351727616,swapFree=120151798505472,swapResv=56026138562560 1727450040000000000
> memory.cpuVm,host=serv vm.aggregate.anonpgin=1354,vm.aggregate.anonpgout=27925,vm.aggregate.pgin=1358,vm.aggregate.pgout=2605,vm.aggregate.pgpgin=1358,vm.aggregate.pgpgout=28402,vm.aggregate.pgswapin=0,vm.aggregate.pgswapout=0,vm.aggregate.swapin=0,vm.aggregate.swapout=0 1727450040000000000
> memory.zone,host=serv,zone=global rss=181059584,swap=309559296 1727450040000000000
> memory.zone,host=serv,zone=serv-backup rss=26521600,swap=32051200 1727450040000000000
> memory.zone,host=serv,zone=serv-media rss=25391104,swap=65277952 1727450040000000000
```
