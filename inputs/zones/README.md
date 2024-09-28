# illumos Zones Input Plugin

Gathers high-level metrics about the zones on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

This plugin requires no configuration.

```toml
# Report on zone states, brands, and other properties.
[[inputs.illumos_zones]]
  # no configuration
```

### Metrics

- zones
  - fields:
    - age (integer, count of seconds since zone was created)
    - age (integer, count of seconds since zone last booted)
  - tags:
    - name (the zone name)
    - brand (the zone brand)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Get zone uptimes.

```
ts("zones.uptime")
```

### Example Output

```
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-dns,status=running uptime=5650 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-dns,status=running age=5958438 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-media,status=running uptime=5649 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-media,status=running age=5956972 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-pkg,status=running uptime=5648 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-pkg,status=running age=5957832 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-cron,status=running uptime=5649 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-cron,status=running age=5957082 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-build,status=running uptime=5648 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-build,status=running age=5953843 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-backup,status=running uptime=5648 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-backup,status=running age=5306328 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-wf,status=running uptime=5646 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-wf,status=running age=5957497 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-mariadb,status=running uptime=5650 1727516123000000000
> zones,brand=lipkg,host=serv,ipType=excl,name=serv-mariadb,status=running age=5957187 1727516123000000000

```
