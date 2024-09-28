# illumos CPU Input Plugin

Gathers metrics about CPU usage on an illumos system.

If it is running in the global zone, this plugin is able to collect
information about global CPU usage, and also about user and system usage in
each non-global zone.

I don't think this input will work on Solaris.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports on illumos CPU usage
[[inputs.illumos_cpu]]
  ## report stuff from the cpu_info kstat. As of now it's just the current clock speed and some
  ## potentially useful tags
  # cpu_info_stats = true
  ## Produce metrics for sys and user CPU consumption in every zone
  # zone_cpu_stats = true
  ## Which cpu:sys kstat metrics you wish to emit. They probably won't all work, because they
  ## some will have a value type which is not an unsigned int
  # sys_fields = ["cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"]
  ## "cpu_ticks_idle", cpu_ticks_kernel", cpu_ticks_user", cpu_ticks_wait", }
```

### Metrics
- cpu.info
  - fields:
    - speed (int, current clock speed of core)
  - tags:
    - chipID (string, numeric ID of processor)
    - clockMHz (string, specified speed of processor)
    - state (string, state of CPU, e.g. `on-line`)
    - coreID (string, numeric ID of core)

- cpu.zone
  - fields:
    - sys (int, counter nsec on CPU)
    - user (int, counter nsec on CPU)
  - tags:
    - name (string, zone)

- cpu.nsec
  - fields:
    - sys (int, counter nsec on CPU)
    - user (int, counter nsec on CPU)
    - dtrace (int, counter nsec on CPU)
    - intr (int, counter nsec on CPU)
  - tags:
    - coreID (string, numeric ID of core)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Get system CPU time, aggregated across all CPUS, as a percentage.

```
ts("deriv(sum(ts("cpu.nsec.system"), coreID)) / (count(ts("cpu.nsec.system")) * 1e7)")
```

Simple CPU usage broken down by zone.

```
sum(deriv(ts("cpu.zone.*")), name)
```

### Example Output

```
> cpu.info,chipID=0,clockMHz=2208,coreID=0,host=serv,state=on-line speed=2200000000 1727447689000000000
> cpu.info,chipID=0,clockMHz=2208,coreID=1,host=serv,state=on-line speed=2200000000 1727447689000000000
> cpu.info,chipID=0,clockMHz=2208,coreID=0,host=serv,state=on-line speed=2200000000 1727447689000000000
> cpu.info,chipID=0,clockMHz=2208,coreID=1,host=serv,state=on-line speed=2200000000 1727447689000000000
> cpu.zone,host=serv,name=global sys=1614858873401,user=1051505614579 1727447689000000000
> cpu.zone,host=serv,name=serv-backup sys=33745707826,user=96416832898 1727447689000000000
> cpu.zone,host=serv,name=serv-media sys=109226122003,user=157877788356 1727447689000000000
> cpu.zone,host=serv,name=serv-www-records sys=28332243504,user=66530074064 1727447689000000000
> cpu.zone,host=serv,name=serv-dns sys=26313196079,user=48062555919 1727447689000000000
> cpu.zone,host=serv,name=serv-wf sys=30445814311,user=164062486965 1727447689000000000
> cpu.zone,host=serv,name=serv-ansible sys=60902025642,user=198089374005 1727447689000000000
> cpu.zone,host=serv,name=serv-build sys=32249197570,user=120713998853 1727447689000000000
> cpu.zone,host=serv,name=serv-ws sys=30104296860,user=108273477966 1727447689000000000
> cpu.zone,host=serv,name=serv-mariadb sys=26028267103,user=54691671158 1727447689000000000
> cpu.zone,host=serv,name=serv-pkg sys=32724093112,user=106816149481 1727447689000000000
> cpu.zone,host=serv,name=serv-cron sys=24997534139,user=52723261418 1727447689000000000
> cpu.zone,host=serv,name=serv-www-proxy sys=25583006627,user=55402470176 1727447689000000000
> cpu.zone,host=serv,name=serv-fs sys=94118767307,user=147827015458 1727447689000000000
> cpu,coreID=0,host=serv nsec.dtrace=80003423,nsec.intr=141697867897,nsec.kernel=1366703643495,nsec.user=585585246815 1727447689000000000
> cpu,coreID=1,host=serv nsec.dtrace=76245864,nsec.intr=34915273891,nsec.kernel=1001005495978,nsec.user=626944227314 1727447689000000000
> cpu,coreID=2,host=serv nsec.dtrace=82728479,nsec.intr=14349457873,nsec.kernel=934539272078,nsec.user=560842246759 1727447689000000000
> cpu,coreID=3,host=serv nsec.dtrace=83912696,nsec.intr=14339323290,nsec.kernel=942298725104,nsec.user=645552073673 1727447689000000000
```
