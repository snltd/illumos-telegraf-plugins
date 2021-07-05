#!/bin/ksh

# Start Telegraf, as a daemon. It's stopped by an SMF :kill in the manifest.

. /lib/svc/share/smf_include.sh

TELEGRAF=/bin/telegraf
CONFIG=/etc/telegraf/telegraf.conf
PIDFILE=/var/run/telegraf.pid

$TELEGRAF --config=$CONFIG --pidfile=$PIDFILE

sleep 1

if test -f $PIDFILE && /bin/ps -p $(cat $PIDFILE) >/dev/null
then
	EXIT_CODE=$SMF_EXIT_OK
else
	EXIT_CODE=$SMF_EXIT_ERR_FATAL
fi

exit $EXIT_CODE
