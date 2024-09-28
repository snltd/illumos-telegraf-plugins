## illumos Process Input Plugin

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
# Reports on illumos processes, like prstat(1)
[[inputs.illumos_process]]
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

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Show the top CPU consumers.

```
deriv(ts("process.rtime"))
```

### Example Output

```
> process,contract=916,host=serv,name=cron,uid=0,zone=serv-mariadb,zoneid=9 inblk=13i 1727453116000000000
> process,contract=919,host=serv,name=mariadbd,uid=70,zone=serv-mariadb,zoneid=9 inblk=11i 1727453116000000000
> process,contract=585,host=serv,name=nscd,uid=0,zone=serv-pkg,zoneid=11 inblk=9i 1727453116000000000
> process,contract=717,host=serv,name=named,uid=53,zone=serv-dns,zoneid=4 inblk=6i 1727453116000000000
> process,contract=703,host=serv,name=nscd,uid=0,zone=serv-build,zoneid=7 inblk=5i 1727453116000000000
> process,contract=799,host=serv,name=nscd,uid=0,zone=serv-mariadb,zoneid=9 inblk=4i 1727453116000000000
> process,contract=-1,host=serv,name=pageout,uid=0,zone=global,zoneid=0 oublk=2605i 1727453116000000000
> process,contract=1030,host=serv,name=netcfgd,uid=17,zone=serv-fs,zoneid=13 oublk=0i 1727453116000000000
> process,contract=1075,host=serv,name=automountd,uid=0,zone=serv-fs,zoneid=13 oublk=0i 1727453116000000000
> process,contract=165,host=serv,name=svc.configd,uid=0,zone=serv-ansible,zoneid=6 oublk=0i 1727453116000000000
> process,contract=59,host=serv,name=ntpd,service=svc:/network/ntp:default,uid=0,zone=global,zoneid=0 oublk=0i 1727453116000000000
> process,contract=7,host=serv,name=ipmgmtd,service=svc:/network/ip-interface-management:default,uid=16,zone=global,zoneid=0 oublk=0i 1727453116000000000
> process,contract=315,host=serv,name=in.mpathd,uid=0,zone=serv-ansible,zoneid=6 oublk=0i 1727453116000000000
> process,contract=406,host=serv,name=ipmgmtd,uid=16,zone=serv-build,zoneid=7 oublk=0i 1727453116000000000
> process,contract=522,host=serv,name=nscd,uid=0,zone=serv-media,zoneid=2 oublk=0i 1727453116000000000
> process,contract=727,host=serv,name=ttymon,uid=0,zone=serv-pkg,zoneid=11 oublk=0i 1727453116000000000
> process,contract=-1,host=serv,name=zpool-big,uid=0,zone=global,zoneid=0 pctcpu=256i 1727453116000000000
> process,contract=67,host=serv,name=telegraf,service=svc:/sysdef/telegraf:default,uid=108,zone=global,zoneid=0 pctcpu=140i 1727453116000000000
> process,contract=-1,host=serv,name=zpool-rpool,uid=0,zone=global,zoneid=0 pctcpu=55i 1727453116000000000
> process,contract=1214,host=serv,name=hx.bin,uid=264,zone=serv-ws,zoneid=8 pctcpu=53i 1727453116000000000
> process,contract=5,host=serv,name=svc.configd,uid=0,zone=global,zoneid=0 pctcpu=48i 1727453116000000000
> process,contract=904,host=serv,name=java,uid=104,zone=serv-wf,zoneid=5 pctcpu=31i 1727453116000000000
> process,contract=215,host=serv,name=svc.configd,uid=0,zone=serv-ws,zoneid=8 pctcpu=27i 1727453116000000000
> process,contract=234,host=serv,name=svc.configd,uid=0,zone=serv-www-proxy,zoneid=12 pctcpu=23i 1727453116000000000
> process,contract=-1,host=serv,name=fsflush,uid=0,zone=global,zoneid=0 pctcpu=23i 1727453116000000000
> process,contract=1026,host=serv,name=svc.configd,uid=0,zone=serv-fs,zoneid=13 pctcpu=23i 1727453116000000000
> process,contract=904,host=serv,name=java,uid=104,zone=serv-wf,zoneid=5 pctmem=493i 1727453116000000000
> process,contract=919,host=serv,name=mariadbd,uid=70,zone=serv-mariadb,zoneid=9 pctmem=214i 1727453116000000000
> process,contract=72,host=serv,name=fmd,service=svc:/system/fmd:default,uid=0,zone=global,zoneid=0 pctmem=86i 1727453116000000000
> process,contract=923,host=serv,name=rackup,uid=4567,zone=serv-www-records,zoneid=3 pctmem=57i 1727453116000000000
> process,contract=67,host=serv,name=telegraf,service=svc:/sysdef/telegraf:default,uid=108,zone=global,zoneid=0 pctmem=55i 1727453116000000000
> process,contract=717,host=serv,name=named,uid=53,zone=serv-dns,zoneid=4 pctmem=44i 1727453116000000000
> process,contract=1214,host=serv,name=hx.bin,uid=264,zone=serv-ws,zoneid=8 pctmem=42i 1727453116000000000
> process,contract=1067,host=serv,name=telegraf,uid=108,zone=serv-fs,zoneid=13 pctmem=41i 1727453116000000000
> process,contract=81,host=serv,name=dtrace,service=svc:/sysdef/cron_monitor:default,uid=107,zone=global,zoneid=0 pctmem=32i 1727453116000000000
> process,contract=101,host=serv,name=smbd,service=svc:/network/smb/server:default,uid=0,zone=global,zoneid=0 pctmem=25i 1727453116000000000
```
