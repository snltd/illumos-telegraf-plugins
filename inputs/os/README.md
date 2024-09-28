# illumos Zones Input Plugin

Extracts basic information about the running OS and kernel.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

This plugin requires no configuration.

```toml
# Reports illumos operating system information
[[inputs.illumos_os]]
  Outputs 1, but with tags that can be combined with other metrics.
```

### Metrics
- os
  - fields:
    - release (int, always `1`)
    - name (string, the distribution, e.g. `OmniOS`)
    - build_id (string)
    - version (string)
    - kernel (string, running kernel, from `uname -v`)

### Example Output

```
> os,build_id=151050.19.2024.09.15,host=serv,kernel=omnios-r151050-49db1c0a0fe,name=OmniOS,version=r151050s release=1i 1727452520000000000
```
