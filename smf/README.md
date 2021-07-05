Here is some "get you started" SMF content.

It should all work with no changes if the files go here:

* `/bin/telegraf` -- the `telegraf` binary
* `/lib/svc/manifest/site/telegraf.xml` -- the SMF manifest
* `/lib/svc/method/telegraf.sh` -- the SMF method referenced by `telegraf.xml`
* `/etc/telegraf/telegfaf.conf` -- your Telegraf config

Telegraf will run as the `daemon` user and `bin` group, with the extra
`file_dac_search` privilege, which lets the `smf` plugin see inside zones.
