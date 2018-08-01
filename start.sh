#!/bin/bash

PERMISSIONS_OK=0

if [ ! -r "$GF_PATHS_CONFIG" ]; then
    echo "GF_PATHS_CONFIG='$GF_PATHS_CONFIG' is not readable."
    PERMISSIONS_OK=1
fi

if [ ! -w "$GF_PATHS_DATA" ]; then
    echo "GF_PATHS_DATA='$GF_PATHS_DATA' is not writable."
    PERMISSIONS_OK=1
fi

if [ $PERMISSIONS_OK -eq 1 ]; then
    echo "You may have issues with file permissions, more information here: http://docs.grafana.org/installation/docker/#migration-from-a-previous-version-of-the-docker-container-to-5-1-or-later"
    exit 1
fi

configuration=""
if [ $# -ge 1 ] && [ $1 == "production" ]; then
    configuration="/etc/grafana/grafana.ini"
else
    configuration="/etc/grafana/defaults.ini"
fi

if [ -z $configuration ]; then
   echo "exec_prod params error."
   exit 1
else
   /usr/sbin/grafana-server --config=$configuration --pidfile=/var/run/grafana/grafana-server.pid cfg:default.paths.logs=/var/log/grafana cfg:default.paths.data=/var/lib/grafana cfg:default.paths.plugins=/var/lib/grafana/plugins cfg:default.paths.provisioning=/etc/grafana/provisionin >> /var/log/grafana/start.log
fi
