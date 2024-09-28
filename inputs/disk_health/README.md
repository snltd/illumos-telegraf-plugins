# illumos Disk Error Input Plugin

Reports disk errors on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports on illumos disk errors
[[inputs.illumos_disk_health]]
  ## The kstat fields you wish to emit. 'kstat -c device_error' will show what is collected. Field
  ## names will be camelCased in the metric path.
  # fields = ["Hard Errors", "Soft Errors", "Transport Errors", "Illegal Request"]
  ## The tags you wish your data points to have. Not all devices are able to supply all tags, but
  ## they will fail silently. Tag names are camelCased.
  # tags = ["Vendor", "Serial No", "Product", "Revision"]
  ## Report on the following devices. Specifying none reports on all.
  # devices = ["sd6"]
```

### Metrics
- diskHealth
  - fields:
    - hardErrors (int, gauge)
    - softErrors (int, gauge)
    - transportErrors (int, gauge)
    - illegalRequest (int, gauge)
  - tags:
    - model (string, text info about drive, if reported)
    - product (string, text info about drive, if reported)
    - vendor (string, text info about drive, if reported)
    - serialNo (string, text info about drive, if reported)
    - size (string, text info about drive, if reported)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Get hard errors currently reported by kernel for all drives.

```
ts("diskHealth.hardErrors")
```

### Example Output

```
> diskHealth,host=serv,model=CT1000P3SSD8,revision=P9CR30A,serialNo=2301E699B2E7,size=931.5Gb hardErrors=0,illegalRequest=0,softErrors=0,transportErrors=0 1727448858000000000
> diskHealth,host=serv,product=Samsung\ SSD\ 870,revision=2B6Q,serialNo=S5STNF0TA09681M,size=3.6Tb,vendor=ATA hardErrors=0,illegalRequest=0,softErrors=0,transportErrors=0 1727448858000000000
```
