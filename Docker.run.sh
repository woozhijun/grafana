#!/bin/bash

docker rm -f grafana
docker volume create grafana-storage

docker run -d \
  -p 3000:3000 \
  --name grafana \
  --user 472
  -v grafana-storage:/var/lib/grafana \
  -e "GF_SERVER_ROOT_URL=http://grafana.monitor.io" \
  -e "GF_EXEC_PROD=develop" \
  -e "GF_INSTALL_PLUGINS=grafana-piechart-panel,mtanda-histogram-panel,alexanderzobnin-zabbix-app,grafana-kairosdb-datasource,abhisant-druid-datasource,michaeldmoore-annunciator-panel,digiapulssi-breadcrumb-panel,btplc-trend-box-panel,natel-discrete-panel,vonage-status-panel,btplc-status-dot-panel,grafana-clock-panel,grafana-simple-json-datasource" \
  docker.mobike.io/apm/grafana:`git rev-parse --short=8 HEAD` \
