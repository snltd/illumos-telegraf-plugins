# illumos NFS Client Input Plugin

Gathers kstat metrics relating to an illumos system's NFS client traffic. It
works with any supported NFS protocol version.

Each zone keeps its own NFS kstats. There's no (elegant) way for the global
zone to see the NFS kstats of an NGZ, so if you care about those, you'll have
to run a dedicated Telegraf in the zone(s).

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports illumos NFS client statistics
[[inputs.illumos_nfs_client]]
  ## The NFS versions you wish to monitor
  # nfs_versions = ["v3", "v4"]
  ## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfsreqcnt' lists the
  ## possibilities
  # fields = ["read", "write", "remove", "create", "getattr", "setattr"]
```

### Metrics
- nfs.client
  - fields:
    - selected by user 
  - tags:
    - nfsVersion (NFS protocol major version, e.g. "v4")

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Write ops for NFSv4

```
deriv(ts("dev.telegraf.nfs.client.write", nfsVersion="v4"))
```

All reads

```
deriv(ts("dev.telegraf.nfs.client.read")) 
```

### Example Output

```
> nfs.client,host=serv,nfsVersion=v3 create=0,getattr=0,read=0,remove=0,setattr=0,write=0 1727452326000000000
> nfs.client,host=serv,nfsVersion=v4 create=0,getattr=0,read=0,remove=0,setattr=0,write=0 1727452326000000000
```
