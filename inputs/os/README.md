# Illumos Zones Input Plugin

Extracts basic information about the running OS and kernel.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

This plugin requires no configuration.

```toml
[[inputs.example]]
```

### Metrics

- os
    - name (string, the distribution, e.g. `OmniOS`)
    - build_id (string)
    - version (string)
    - kernel (string, running kernel, from `uname -v`)
  - fields:
    - release (int, always `1`)

