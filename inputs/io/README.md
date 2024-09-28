# illumos IO Input Plugin

Reports IO statistics on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports on illumos IO
[[inputs.illumos_io]]
  ## The kstat fields you wish to emit. 'kstat -c disk' will show what is collected. Not defining
  ## any fields sends everything, which is probably not what you want.
  # fields = ["reads", "nread", "writes", "nwritten"]
  ## Report on the following kstat modules. You likely have 'sd' and 'zfs'. Specifying none
  ## reports on all.
  # modules = ["sd", "zfs"]
  ## Report on the following devices, inside the above modules. Specifying none reports on all.
  # devices = ["sd0"]
```

### Metrics
- io
  - fields:
    - defined by user, from `kstat -c disk`
  - tags:
    - device (string, kstat device name)
    - module (string, kstat module)
    - product (string, text info about drive, if reported)
    - serialNo (string, text info about drive, if reported)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Show the rate of reads broken down by ZFS pool.

```
deriv(ts("io.nread", module="zfs"))
```

### Example Output

```
> io,device=blkdev0,host=serv,module=blkdev nread=18791200768,nwritten=6912024576,wcnt=0 1727449796000000000
> io,device=rpool,host=serv,module=zfs nread=1100181504,nwritten=5221171200,wcnt=0 1727449796000000000
> io,device=big,host=serv,module=zfs nread=17998178304,nwritten=3862097920,wcnt=0 1727449796000000000
> io,device=fast,host=serv,module=zfs nread=17070772224,nwritten=1690836992,wcnt=0 1727449796000000000
> io,device=sd0,host=serv,module=sd,product=Samsung\ SSD\ 870\ ,serialNo=S5STNF0TA09681M nread=17998227740,nwritten=3862097920,wcnt=0 1727449796000000000
```
