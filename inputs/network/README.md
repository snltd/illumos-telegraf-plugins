# illumos Net Input Plugin

Gathers high-level metrics about network traffic through illumos NICs and
VNICs.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports on illumos NIC Usage. Zone-aware.
[[inputs.illumos_network]]
  ## The kstat fields you wish to emit. 'kstat -c net' will show what is collected. Defining
  ## no fields sends everything, which is probably not what you want.
  # fields = ["obytes64", "rbytes64"]
  ## The VNICs you wish to observe. Again, specifying none collects all.
  # vnics  = ["net0"]
  ## The zones you wish to monitor. Specifying none collects all.
  # zones = []
```

### Metrics
- network
  - fields:
    - selected by user from `kstat -c net`
  - tags:
    - name (string, name of VNIC, "none" in case physical NIC)
    - link (string, physical NIC to which VNIC belongs, "none" in case of physical NIC)
    - speed (string, text info about NIC/VNIC, if reported)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

Bytes read by non-global zones

```
rate(ts("net.rbytes64", zone != "global"))
```

### Example Output

```
> net,host=serv,link=none,name=e1000g0,speed=unknown,zone=global obytes64=1262389396,rbytes64=1258587594 1727450591000000000
> net,host=serv,link=e1000g0,name=fs_net0,speed=1000mbit,zone=serv-fs obytes64=120331996,rbytes64=82952634 1727450591000000000
> net,host=serv,link=e1000g0,name=build_net0,speed=1000mbit,zone=serv-build obytes64=824723,rbytes64=20393222 1727450591000000000
> net,host=serv,link=e1000g0,name=backup_net0,speed=1000mbit,zone=serv-backup obytes64=1013776,rbytes64=24607078 1727450591000000000
> net,host=serv,link=e1000g0,name=media_net0,speed=1000mbit,zone=serv-media obytes64=168086,rbytes64=11760982 1727450591000000000
> net,host=serv,link=e1000g0,name=pkg_net0,speed=1000mbit,zone=serv-pkg obytes64=956955,rbytes64=23390614 1727450591000000000
> net,host=serv,link=e1000g0,name=ansible_net0,speed=1000mbit,zone=serv-ansible obytes64=96594087,rbytes64=14349228 1727450591000000000
> net,host=serv,link=e1000g0,name=mariadb_net0,speed=1000mbit,zone=serv-mariadb obytes64=207485,rbytes64=3069233 1727450591000000000
```
