# illumos Zpool Input Plugin

Gathers high-level metrics about the ZFS pools on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports the health and status of ZFS pools.
[[inputs.illumos_zpool]]
  ## The metrics you wish to report. They can be any of the headers in the output of 'zpool list',
  ## and also a numeric interpretation of 'health'.
  # fields = ["size", "alloc", "free", "cap", "dedup", "health"]
  ## Status metrics are things like ongoing resilver time, ongoing scrub time, error counts
  ## and whatnot
  # status = true
```

### Metrics
- zpool
  - tags:
    - name (the pool name)
  - fields:
    - size (float, the size of the pool in bytes)
    - alloc (float, number of allocated bytes)
    - free (float, number of free bytes)
    - cap (int, the percentage of the pool used up)
    - dedup (float, the pool's deduplication ratio)
    - frag (int, the percentage fragmentation of the pool)
    - health (int, a numeric mapping of the pool's health, where 0: ONLINE, 1:
      DEGRADED, 2: SUSPENDED, 3: UNAVAIL, 4: FAULTED, and 99 identifies and
      unknown value
- zpool.status
  - tags:
    - name (the pool name)
  - fields:
    - resilverTime (int, number of seconds since resilver began)
    - scrubTime (int, number of seconds since active scrub began, zero if no
    scrub is in progress)
    - timeSinceScrub (int, number of seconds since a scrub completed)
- zpool.status.errors
  - tags:
    - pool (the pool name)
    - device (short device name)
  - fields:
    - cksum (int, count of checksum errors)
    - read (int, count of read errors)
    - write (int, count of write errors)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Show the percentage of the root pool in use

```
ts("zpool.cap", name="rpool")
```

Find errant pools, for an alert.

```
highpass(0, ts("zpool.health"))
```

### Example Output

```
> zpool,host=cube,name=big alloc=2957686278717.44,cap=74i,dedup=1,frag=2i,free=1029718409216,health=0i,size=3980232092549.12 1618875483000000000
> zpool,host=cube,name=fast alloc=111669149696,cap=39i,dedup=1,frag=25i,free=169651208192,health=0i,size=281320357888 1618875483000000000
> zpool,host=cube,name=rpool alloc=61310658150.4,cap=28i,dedup=1,frag=63i,free=152471339008,health=0i,size=213674622976 1618875483000000000
```
