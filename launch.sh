#!/bin/bash

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
   /usr/sbin/grafana-server --config=$configuration --pidfile=/var/run/grafana/grafana-server.pid cfg:default.paths.logs=/var/log/grafana cfg:default.paths.data=/var/lib/grafana cfg:default.paths.plugins=/var/lib/grafana/plugins cfg:default.paths.provisioning=/etc/grafana/provisionin
fi
