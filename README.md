# Illumos Telegraf Plugins

This is a collection of Illumos-specific
[Telegraf](https://github.com/influxdata/telegraf) input plugins which I wrote
because I needed them.

They work fine on my OmniOS boxes, collecting the information which I wanted
to see, and presenting it in a way I think is useful. I'm not sure exactly how
well they will work on SmartOS, but my guess would be "fine".

Things to note.

* Most of the plugins use KStats, and the KStat values are sent "as is". That
  is, I do not calculate rates inside Telegraf. Things like CPU usage, which
  the kernel measures as "total time spent on CPU" will just go up and up. I
  don't mind this because my graphing software
  ([Wavefront](https://wavefront.com) lets me wrap the series in a `rate()`
  function.
* The testing sample is very small. You may have hardware which produces
  different KStats to mine, so you may be missing tags in places. I'm thinking
  specifically of disks, but who knows what else.
* Some of the plugins (e.g. memory) will work on an x86 Solaris system, but
  some (e.g. SMF) won't. Suck it and see. I'd be delighted to receive  PRs if
  anyone modifies the code to work right across SunOS.
* I have no interest in getting any of these plugins merged with the official
  Telegraf distribution. Illumos is a serious minority interest these days,
  and I can't imagine the Telegraf people have any desire to be encumbered
  with support for it. There are also difficulties in testing and
  cross-compilation, because the KStats module uses CGo. If someone wants to
  chase this, make a fork, or in any way improve the end-user experience, help
  yourself.
* You can only run the tests on an Illumos box. Properly mocking all the KStat
  calls wasn't something I wanted to get involved in.
* Telegraf allocates a huge amount (~5Gb) of virtual memory. This is nothing
  to do with my plugins, but I've seen the code fail with `ENOMEM` because of
  a lack of swap.

All of that said, I've found the plugins reliable and useful.

## Building

I build on OmniOS, in a `lipkg` zone with the following packages.

```
developer/build/gnu-make
developer/versioning/git
ooce/developer/go-116
```

Get the Telegraf source and pick a release. I use 1.16.3. 1.17 requires
substantially more hacking around to build, and 1.18 allocates a huge amount
of swap.

```
git clone https://github.com/influxdata/telegraf.git
cd telegraf
git checkout v1.16.3
vi plugins/inputs/all/all.go
```

and add these lines to the `inputs()`. Feel free to omit any you don't need:
the fewer plugins you try to build, the lower your chance of failure!

```go
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_cpu"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_disk_health"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_fma"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_io"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_memory"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_network"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_nfs_client"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_nfs_server"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_patches"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_proc"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_smf"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_zfs_arc"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_zones"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/illumos_zpool"
```

I remove the default equivalents too, since they're useless to us.

```
gmake
```

## The Plugins

### illumos_cpu
CPU usage, presented in nanoseconds, as per the kstats. It's up to you and
your graphing software to make rates, percentages, or whatever you find
useful. Can combine cores into overall stats, to keep down the point rate. Can
also report per-zone CPU usage if running in the global.

### illumos_disk_health
Uses the `device_error` kstats to keep track of disk errors. Tries its best to
tag the metrics with information about the disks like vendor, serial number
etc.

### illumos_fma
A very experimental plugin which parses the output of `fmadm(1m)` and
`fmstat(1m)` to produce information on system failures.

### illumos_io
Gets data about IO throughput.

### illumos_memory
Aggregates virtual memory information from a number of kstats and, if you want
it, the output of `swap(1m)`. Swapping/paging info defaults to per-cpu, but
can be aggregated to save point rate.

### illumos_network
Collects network KStats. If Telegraf is running in the global zone, the plugin
can present per-zone statistics.

### illumos_nfs_client
Basic measurement of NFS client stats, for all NFS protocol versions. Each
zone has its own set of KStats, so if you want per-zone NFS stats, you'll have
to run Telegraf in the zones.

### illumos_nfs_server
NFS server KStats. Not much more to say.

## illumos_patches
Tells you how many of your installed packages are ready for upgrade.

### illumos_smf
Parses the output of `svcs(1m)` to count the number of SMF services in
particular states. Also reports errant services with sufficient tagging to
easily track them down and fix them.

### illumos_zfs_arc
Reports ZFS ARC statistics.

### illumos_zones
Turns `zoneadm list` into numbers, and tells you how old your zones are.

### illumos_zpool
High-level ZFS pool statistics from the output of `zpool list`.

## Contributing

Fork it, fix it, push it, PR it. I expect tests!
