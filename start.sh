#!/bin/sh -e

PERMISSIONS_OK=0

if [ ! -r "$GF_PATHS_CONFIG" ]; then
    echo "GF_PATHS_CONFIG='$GF_PATHS_CONFIG' is not readable."
    PERMISSIONS_OK=1
fi

if [ ! -w "$GF_PATHS_DATA" ]; then
    echo "GF_PATHS_DATA='$GF_PATHS_DATA' is not writable."
    PERMISSIONS_OK=1
fi

if [ ! -r "$GF_PATHS_HOME" ]; then
    echo "GF_PATHS_HOME='$GF_PATHS_HOME' is not readable."
    PERMISSIONS_OK=1
fi

if [ $PERMISSIONS_OK -eq 1 ]; then
    echo "You may have issues with file permissions, more information here: http://docs.grafana.org/installation/docker/#migration-from-a-previous-version-of-the-docker-container-to-5-1-or-later"
    exit 1
fi

configuration=""
if [ $GF_EXEC_PROD == "production" ]; then
    configuration="/etc/grafana/grafana.ini"
else
    configuration="/etc/grafana/defaults.ini"
fi

if [ -z $configuration ]; then
   echo "exec_prod params error."
   exit 1
else
   echo /usr/sbin/grafana-server --homepath=$GF_PATHS_HOME --config=$configuration cfg:default.paths.logs=$GF_PATHS_LOGS cfg:default.paths.data=$GF_PATHS_DATA cfg:default.paths.plugins=$GF_PATHS_PLUGINS cfg:default.paths.provisioning=$GF_PATHS_PROVISIONING 
   /usr/sbin/grafana-server --homepath=$GF_PATHS_HOME --config=$configuration cfg:default.paths.logs=$GF_PATHS_LOGS cfg:default.paths.data=$GF_PATHS_DATA cfg:default.paths.plugins=$GF_PATHS_PLUGINS cfg:default.paths.provisioning=$GF_PATHS_PROVISIONING 
fi
