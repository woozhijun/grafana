#!/bin/bash

docker rm -f grafana

docker build -t grafana:v5.2.1 .

docker build -t grafana:latest-with-plugins \
  --build-arg "GRAFANA_VERSION=v5.2.1" \
  --build-arg "GF_INSTALL_PLUGINS=grafana-piechart-panel,mtanda-histogram-panel,alexanderzobnin-zabbix-app,grafana-kairosdb-datasource,abhisant-druid-datasource,michaeldmoore-annunciator-panel,digiapulssi-breadcrumb-panel,btplc-trend-box-panel,natel-discrete-panel,vonage-status-panel,btplc-status-dot-panel,grafana-clock-panel,grafana-simple-json-datasource"
