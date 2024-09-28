# illumos SMF Input Plugin

Gathers high-level metrics about SMF services on an illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# Reports the states of SMF services for a single zone or across a host.
[[inputs.illumos_smf]]
  ## The service states you wish to count.
  # svc_states = ["online", "uninitialized", "degraded", "maintenance"]
  ## The Zones you wish to examine. If this is unset or empty, all visible zones are counted.
  # zones = ["zone1", "zone2"]
  ## Whether or not you wish to generate individual, detailed points for services which are in
  ## svc_states but are not "online"
  # generate_details = true
  ## Use this command to get the elevated privileges svcs requires to observe other zones.
  ## Should be a path, like "/bin/sudo" "/bin/pfexec", but can also be "none", which will
  ## collect only the local zone.
  # elevate_privs_with = "/bin/sudo"
```

If it is running in the global zone, this plugin is able to collect SMF
information for all NGZs, using the `elevate_privs_with` command. You can use
`sudo`, assuming a correctly configured `sudoers`, or set up a profile with
`file_dac_search` and `proc_owner` privileges, and use `pfexec`.

This plugin will not work on Solaris.

### Metrics
- smf
  - fields:
    - states (int, count of services defined by the following tags)
  - tags:
    - zone (string, zone name)
    - state (string, SMF service state)

  - fields:
    - errors (int, always 1)
  - tags:
    - fmri (string, service FMRI)
    - state (string, state the service is in)
    - zone (zone to which service belongs)


### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

To track all services which are not online:

```
ts("smf.states", NOT state="online")
```

To get detailed information on errant services. (Assuming `generate_details`
is true.)

```
ts("smf.errors")
```


### Example Output

```
> smf,host=serv,state=online,zone=serv-fs states=68i 1727453362000000000
> smf,host=serv,state=online,zone=serv-cron states=61i 1727453362000000000
> smf,host=serv,state=online,zone=serv-build states=63i 1727453362000000000
> smf,host=serv,state=online,zone=serv-wf states=62i 1727453362000000000
> smf,host=serv,state=online,zone=serv-dns states=62i 1727453362000000000
> smf,host=serv,state=online,zone=serv-pkg states=62i 1727453362000000000
> smf,host=serv,state=online,zone=serv-ansible states=61i 1727453362000000000
> smf,host=serv,state=online,zone=serv-www-proxy states=62i 1727453362000000000
> smf,host=serv,state=online,zone=serv-mariadb states=62i 1727453362000000000
> smf,host=serv,state=online,zone=serv-www-records states=58i 1727453362000000000
> smf,host=serv,state=online,zone=serv-backup states=61i 1727453362000000000
```
