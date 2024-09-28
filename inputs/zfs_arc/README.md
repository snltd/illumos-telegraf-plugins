# illumos SMF Input Plugin

Gathers ZFS ARC metrics on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports illumos ZFS ARC statistics
[[inputs.illumos_zfs_arc]]
  # fields = ["hits", "misses", "l2_hits", "l2_misses", "prefetch_data_hits",
  # "prefetch_data_misses", "prefetch_metadata_hits", "prefetch_metadata_misses",
  # "demand_data_hits", "demand_data_misses", "demand_metadata_hits", "demand_metadata_misses",
  # "l2_size", "l2_read_bytes", "l2_write_bytes", "l2_cksum_bad", "c", "size"]
```
### Metrics
- smf
  - fields:
  - hits
  - misses
  - l2_hits
  - l2_misses
  - prefetch_data_hits
  - prefetch_data_misses
  - prefetch_metadata_hits
  - prefetch_metadata_misses
  - demand_data_hits
  - demand_data_misses
  - demand_metadata_hits
  - demand_metadata_misses
  - l2_size
  - l2_read_bytes
  - l2_write_bytes
  - l2_cksum_bad
  - c

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Show the L2 cache

```
ts("zfs.arcstats.l2_size")
```

### Example Output

```
> zfs.arcstats,host=serv c=15439577088,demand_data_hits=819547,demand_data_misses=73168,demand_metadata_hits=89643795,demand_metadata_misses=9960140,hits=90463342,misses=10054348,prefetch_data_hits=0,prefetch_data_misses=1888,prefetch_metadata_hits=0,prefetch_metadata_misses=19152,size=1727049728 1727515854000000000
```
