# illumos NFS Server Input Plugin

Gathers kstat metrics relating to an illumos system's NFS server. It works with
any NFS server version.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports illumos NFS server statistics
[[inputs.illumos_nfs_server]]
  ## The NFS versions you wish to monitor.
  # nfs_versions = ["v3", "v4"]
  ## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfsproccnt' lists the
  ## possibilities
  # fields = ["read", "write", "remove", "create", "getattr", "setattr"]
```

### Metrics

- zpool
  - fields:
    - selected by user
  - tags:
    - nfsVersion (string, NFS protocol major version, e.g. "v4")

The final field of any `nfs:0:rfs*` kstat is a valid field.

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Write ops for NFSv4

```
deriv(ts("nfs.server.write", nfsVersion="v4"))
```

All reads

```
deriv(ts("nfs.server.read"))
```


### Example Output

```
> nfs.server,host=cube,nfsVersion=v3 create=0i,getattr=122i,read=194816i,remove=0i,setattr=0i,write=0i 1618958834000000000
> nfs.server,host=cube,nfsVersion=v4 create=291i,getattr=34952i,read=10793i,remove=1930i,setattr=854i,write=987i 1618958834000000000
```
