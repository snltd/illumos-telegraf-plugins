## Illumos Zones Input Plugin

Works a little bit like `prstat(1)`, collecting process resource information.

On every collection interval, the plugin examines every visible process, populating 
`prusage` and `psinfo` structs. See [the proc(5) man page](https://www.illumos.org/man/5/proc)
for more information.

Most of these fields can be exposed as metrics or tags. See below for details.

For every value you choose, the top `K` processes (set by the `TopK` parameter)
will be send as metrics, decorated with whatever tags you list.

### Caveats
* Short-lived processes can be missed
* Lots of ephemeral processes can make charts hard to read
* On boxes with a lot of processes, the collector can be quite heavy.
* All values are exposed as literal values, which are gauges, so you may need
  to wrap things in a `rate()` in your graphing/alerting software.
* Using the `pid` tag could lead to high cardinality.

### Extras

Setting `ExpandZoneTag` to true will add a `zone` tag with the name of the zone
in which the process runs. This requires a shell-out to `zoneadm` on each 
collection.

Setting `ExpandContractTag` to true will, if a process is under control of an
SMF service, add a `service` tag with the FRMI of that service. This requires a
shell-out to `svcs` on each collection.

### Requirements

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
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
```

### Metrics

- process
  - fields:
    - rtime (int64, total LWP real (elapsed)_ time)
    - utime (int64, user level cpu time)
    - stime (int64, system call cpu time)
    - wtime (int64, wait-cpu (latency) time)
    - inblk (int64, input blocks)
    - oublk (int64, output blocks)
    - sysc (int64, system calls)
    - ioch (int64, chars read and written)
    - size (int64, size of process image in bytes [kstat is kb, but we convert])
    - rssize (int64, resident set size in bytes [kstat is kb, but we convert])
    - pctcpu (int64, %age of total CPU usage. Divide by 10,000 for the actual value)
    - pctmem (int64, %age of total memory. Divide by 10,000 for the actual value)
    - nlwp (int64, number of LWPs in process)
    - count (int64)
  - tags:
    - name (string, name of execed file)
    - uid (string, real user ID)
    - gid (string, real group ID)
    - euid (string, effective user ID)
    - egid (string, effective group ID)
    - pid (string, process ID)
    - ppid (string, parent process ID)
    - taskid (string, task ID)
    - projid (string, project ID)
    - zoneid (string, zone ID)
    - contract (string, contract ID)

