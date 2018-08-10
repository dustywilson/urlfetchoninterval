#!/bin/bash -e

HTTP_HOST=${HTTP_HOST:-$SITE_PORT_80_TCP_ADDR}

if [ -z "$SITE_PORT_80_TCP_ADDR" ]; then
	echo "You must --link \$CONTAINERNAME:site or this is useless."
	exit 1
fi

if [ "$HTTP_HOST" != "$SITE_PORT_80_TCP_ADDR" ]; then
	echo "$SITE_PORT_80_TCP_ADDR $HTTP_HOST" >> /etc/hosts
fi

while true; do
	sleep 60s
	wget -O/dev/null --header="X-Forwarded-Email: emmaly.wilson@scjalliance.com" --header="X-Forwarded-User: emmaly.wilson" "http://$HTTP_HOST$GET_PATH"
done
