# Illumos CPU Input Plugin

Gathers metrics about CPU usage on an Illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
[[inputs.illumos_cpu]]
  ## report stuff from the cpu_info kstat. As of now it's just the current clock speed and some
  ## potentially useful tags
  # cpu_info_stats = true
  ## Produce metrics for sys and user CPU consumption in every zone
  # zone_cpu_stats = true
  ## Which cpu:sys kstat metrics you wish to emit. They probably won't all work, because they
  ## some will have a value type which is not an unsigned int
  # sys_fields = ["cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"]
  ## "cpu_ticks_idle", cpu_ticks_kernel", cpu_ticks_user", cpu_ticks_wait", }`

  ## The service states you wish to count.
  # svc_states = ["online", "uninitialized", "degraded", "maintenance"]
  ## The Zones you wish to examine. If this is unset or empty, all visible zones are counted.
  # zones = ["zone1", "zone2"]
  ## Whether or not you wish to generate individual, detailed points for services which are in
  ## SvcStates but are not "online"
  # generate_details = true
```

If it is running in the global zone, this plugin is able to collect
information about global CPU usage, and also about user and system usage in
each non-global zone.

I don't think this plugin will work on Solaris.

### Metrics
- cpu
  - fields:
    - info.speed (int, current clock speed of core)
  - tags:
    - chipID (string, numeric ID of processor)
    - clockMHz (string, specified speed of processor)
    - state (string, state of CPU, e.g. `on-line`)

  - fields:
    - zone. (int, always 1)
  - tags:
    - fmri (string, service FMRI)
    - state (string, state the service is in)
    - zone (zone to which service belongs)


### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

To track all services which are not online:

```
ts("dev.telegraf.smf.states", NOT state="online")
```

To get detailed information on errant services. (Assuming `generate_details`
is true.)

```
ts("dev.telegraf.smf.errors")
```


### Example Output

```
> cpu.info,chipID=0,clockMHz=2712,coreID=0,host=cube,state=on-line speed=2701000000 1622407171000000000
> cpu.info,chipID=0,clockMHz=2712,coreID=1,host=cube,state=on-line speed=2701000000 1622407171000000000
> cpu.info,chipID=0,clockMHz=2712,coreID=2,host=cube,state=on-line speed=2701000000 1622407171000000000
> cpu.info,chipID=0,clockMHz=2712,coreID=3,host=cube,state=on-line speed=2701000000 1622407171000000000
> cpu.zone,host=cube,name=global sys=211641867263174,user=220931083284958 1622407171000000000
> cpu.zone,host=cube,name=cube-pkgsrc sys=1418352950997,user=2698387744765 1622407171000000000
> cpu.zone,host=cube,name=cube-pkg sys=7680908127510,user=31217720288340 1622407171000000000
> cpu.zone,host=cube,name=cube-mariadb sys=5613924461156,user=24639166435874 1622407171000000000
> cpu.zone,host=cube,name=cube-www-proxy sys=2647925802944,user=12516837781509 1622407171000000000
> cpu.zone,host=cube,name=cube-ws sys=11530909984334,user=1970920582327167 1622407171000000000
> cpu.zone,host=cube,name=cube-www-records sys=4087866645625,user=24126104355189 1622407171000000000
> cpu.zone,host=cube,name=cube-dns sys=3582897522812,user=19978291236548 1622407171000000000
> cpu.zone,host=cube,name=cube-www-sysdef sys=2807990593186,user=12451011608777 1622407171000000000
> cpu.zone,host=cube,name=cube-audit sys=3731532429600,user=19539708038266 1622407171000000000
> cpu.zone,host=cube,name=cube-media sys=7013045378137,user=24818772877428 1622407171000000000
> cpu.zone,host=cube,name=cube-build sys=2826584182636,user=9603921076345 1622407171000000000
> cpu.zone,host=cube,name=cube-puppet sys=2474776621787,user=20157567611521 1622407171000000000
> cpu.zone,host=cube,name=cube-cron sys=3397604324668,user=23144422908180 1622407171000000000
> cpu.zone,host=cube,name=cube-backup sys=3927911202255,user=21783633827955 1622407171000000000
> cpu.zone,host=cube,name=cube-www-cassingle sys=3685940961354,user=20557968612203 1622407171000000000
> cpu.zone,host=cube,name=cube-www-meetup sys=5685185977205,user=21792599815809 1622407171000000000
> cpu.zone,host=cube,name=cube-wavefront sys=3455866947648,user=22742087043531 1622407171000000000
> cpu,coreID=0,host=cube nsec.dtrace=24606352612,nsec.intr=13646574163308,nsec.kernel=112496661250515,nsec.user=620946813134102 1622407171000000000
> cpu,coreID=1,host=cube nsec.dtrace=23127230208,nsec.intr=3478359474075,nsec.kernel=104851612208465,nsec.user=633782274549740 1622407171000000000
> cpu,coreID=2,host=cube nsec.dtrace=24295697248,nsec.intr=3355762794973,nsec.kernel=106133944726631,nsec.user=625925284403678 1622407171000000000
> cpu,coreID=3,host=cube nsec.dtrace=24175871578,nsec.intr=9713476770533,nsec.kernel=116416964095711,nsec.user=644245589713409 1622407171000000000
```
