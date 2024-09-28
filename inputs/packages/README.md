# illumos Packages Input Plugin

Presents information about upgradeable and installed packages on a zoned
illumos system.

Requires elevated privileges to `zlogin` to non-global zones, which means
some `sudo` or `pfexec` configuration.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

This plugin requires no configuration.

```toml
# Reports the number of packages which can be upgraded.
[[inputs.illumos_packages]]
  ## Whether you wish this plugin to try to refresh the package database. Personally, I wouldn't.
  # refresh = false
  ## Whether to report the number of installed packages
  # installed = true
  ## Whether to report the number of upgradeable packages
  # upgradeable = true
  ## Use this command to get the elevated privileges run commands in other zones via zlogin.
  ## and to run pkg refresh anywhere. Should be a path, like "/bin/sudo" "/bin/pfexec", but can
  ## also be "none", which will collect only the local zone.
  # elevate_privs_with = "/bin/sudo"
```

### Metrics
- package
  - fields:
    - installed (int, number of installed packages)
    - upgradeable (int, number of installed packages)
  - tags:
    - zone (string, the zone name)
    - format (string, the package format: `pkgsrc` or `pkg`)

### Example Output

```
> packages,format=pkg,host=serv,zone=serv-fs upgradeable=10i 1727452933000000000
> packages,format=pkg,host=serv,zone=global upgradeable=0i 1727452934000000000
> packages,format=pkg,host=serv,zone=serv-backup upgradeable=6i 1727452935000000000
> packages,format=pkg,host=serv,zone=serv-build upgradeable=0i 1727452936000000000
> packages,format=pkg,host=serv,zone=serv-ws upgradeable=0i 1727452936000000000
> packages,format=pkg,host=serv,zone=serv-media upgradeable=7i 1727452937000000000
```
