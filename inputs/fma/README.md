# illumos FMA Input Plugin

Reports FMA information on an illumos system. Work in progress as I'm unsure
of its value.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
# A vague, experimental collector for the illumos fault management architecture. I'm not
# sure yet what it is worth recording, and how, so this is almost certainly subject to change
[[inputs.illumos_fma]]
  ## Whether to report fmstat(1m) metrics
  # fmstat = true
  ## Which fmstat modules to report
  # fmstat_modules = []
  ## Which fmstat fields to report
  # fmstat_fields = []
  ## Whether to report fmadm(1m) metrics
  # fmadm = true
  ## Use this command to get elevated privileges required to run fmadm.
  ## Should be a path, like "/bin/sudo" "/bin/pfexec", but can also be "none", which will
  ## omit the fmadm collection.
  # elevate_privs_with = "/bin/sudo"
```

### Metrics
- fma.fmadm
- fma.fmstat

### Sample Queries

### Example Output
