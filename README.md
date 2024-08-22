# illumos Telegraf Plugins

This is a collection of illumos-specific
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
  ([Wavefront](https://wavefront.com)) lets me wrap the series in a `rate()`
  function.
* The testing sample is very small. You may have hardware which produces
  different KStats to mine, so you may be missing tags in places. I'm thinking
  specifically of disks, but who knows what else.
* Some of the plugins (e.g. memory) will work on an x86 Solaris system, but
  some (e.g. SMF) won't. Suck it and see. I'd be delighted to receive  PRs if
  anyone modifies the code to work right across SunOS.
* I have no interest in getting any of these plugins merged with the official
  Telegraf distribution. illumos is a serious minority interest these days,
  and I can't imagine the Telegraf people have any desire to be encumbered
  with support for it. There are also difficulties in testing and
  cross-compilation, because the KStats module uses CGo. If someone wants to
  chase this, make a fork, or in any way improve the end-user experience, help
  yourself.
* You can only run the tests on an illumos box. Properly mocking all the KStat
  calls wasn't something I wanted to get involved in.

All of that said, I've found the plugins reliable and useful.

## Building

This isn't a self-contained software package. It's effectively a big patch to
Telegraf, and you'll have to do a little work to build it.

First, of course, you need a build environment. I use an OmniOS `lipkg`
zone with the following packages.

```
developer/build/gnu-make
developer/versioning/git
ooce/developer/go-116
```

I also have `golangci-lint` installed, because I am an insane masochist who
apparently doesn't think Go has enough petty and arbitrary rules built into
it.

Get the Telegraf source and pick a release. I use 1.16.3. 1.17 requires
substantially more hacking around to build, and 1.18 allocates a huge amount
of swap, which I don't like.

```
$ git clone https://github.com/influxdata/telegraf.git
$ cd telegraf
$ git checkout v1.16.3
$ vi plugins/inputs/all/all.go
```

and add the following lines to the `inputs()`. Feel free to omit any you don't
need

```go
_ "github.com/snltd/illumos-telegraf-plugins/inputs/cpu"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/disk_health"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/fma"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/io"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/memory"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/network"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/nfs_client"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/nfs_server"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/packages"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/smf"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/zfs_arc"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/zones"
_ "github.com/snltd/illumos-telegraf-plugins/inputs/zpool"
```

Certain linux-centric plugins will break a Telegraf build on illumos.  To
build 1.16.3 on my system I had to remove the `modbus`, `ecs`, and `docker`
inputs, but normally I take out way more than that. I pretty much only leave
in the things I actually need.

Now add the latest tag in this repo to `go.mod`. Get it with `git tag`, or
look at [the releases
page](https://github.com/snltd/illumos-telegraf-plugins/releases). Don't
forget the `v`! For example:

```
github.com/snltd/illumos-telegraf-plugins v0.5.0
```

And

```
$ go mod tidy
```

Now you can build Telegraf.

```
$ gmake
```

This may well fail, and you might have to start removing stuff from the
various `all.go` files. For 1.16.3, I had to take the `starlark` line out of
`plugins/processors/all/all.go`. After that, `gmake` succeeded, and I got a
`telegraf` binary.

Once you have a binary, the [`smf` directory](smf) contains just enough SMF to
get you going.

## The Plugins

### cpu
CPU usage, presented in nanoseconds, as per the kstats. It's up to you and
your graphing software to make rates, percentages, or whatever you find
useful. Can report per-zone CPU usage if running in the global.

### disk_health
Uses the `device_error` kstats to keep track of disk errors. Tries its best to
tag the metrics with information about the disks like vendor, serial number
etc. All points relate to an error, so if there are no errors, you get no points.

### fma
A very experimental plugin which parses the output of `fmadm(1m)` and
`fmstat(1m)` to produce information on system failures. Requires elevated 
privileges for `fmadm`.

### io
Gets data about IO throughput, by device or by ZFS pool.

### memory
Aggregates virtual memory information from a number of kstats and, if you want
it, the output of `swap(1m)`. Swapping/paging info defaults to per-cpu, but
can be aggregated to save point rate. Can report NGZ memory usage if running
in the global zone.

### network
Collects network KStats. If Telegraf is running in the global zone, the plugin
can present per-zone statistics.

### nfs_client
Basic measurement of NFS client stats, for all NFS protocol versions. Each
zone has its own set of KStats, so if you want per-zone NFS stats, you'll have
to run Telegraf in the zones.

### nfs_server
NFS server KStats. The same zone limitations apply as for the client.

### os
Gives you a few statistics about the kernel and OS revision.

### packages
Tells you how many packages you have installed, and how many are ready for 
upgrade.  Can refresh the package cache, if you wish it too. Works with `pkg(5)`
and pkgsrc zones. Needs `sudo` or `pfexec` configuring to work properly, as
well as the `file_dac_search` privilege.

### process
Works a bit like `prstat(8)`, showing what processes are spending time on CPU
or in memory. Works across all zones.

### smf
Parses the output of `svcs(1m)` to count the number of SMF services in
particular states. Also reports errant services with sufficient tagging to
easily track them down and fix them. Can report non-global zones from the global
if you set up `pfexec` or `sudo` to allow it.

### zfs_arc
Reports ZFS ARC statistics.

### zones
Turns `zoneadm list` into numbers, and tells you how old your zones are and
how long they've been up.

### zpool
High-level ZFS pool statistics from the output of `zpool list`.

## Contributing

Fork it, fix it, push it, PR it. I expect tests!
