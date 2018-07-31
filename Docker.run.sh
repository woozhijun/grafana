#!/bin/bash

docker rm -f grafana

docker volume create grafana-storage

docker run -d \
  -p 3000:3000 \
  --name grafana \
  -v grafana-storage:/var/lib/grafana \
  -e "GF_SERVER_ROOT_URL=http://grafana.monitor.io" \
  -e "GF_INSTALL_PLUGINS=grafana-piechart-panel,mtanda-histogram-panel,alexanderzobnin-zabbix-app,grafana-kairosdb-datasource,abhisant-druid-datasource,michaeldmoore-annunciator-panel,digiapulssi-breadcrumb-panel,btplc-trend-box-panel,natel-discrete-panel,vonage-status-panel,btplc-status-dot-panel,grafana-clock-panel,grafana-simple-json-datasource" \
  grafana/grafana:v5.2.1 \
  /usr/sbin/grafana-server --config=/etc/grafana/grafana.ini --pidfile=/var/run/grafana/grafana-server.pid cfg:default.paths.logs=/var/log/grafana cfg:default.paths.data=/var/lib/grafana cfg:default.paths.plugins=/var/lib/grafana/plugins cfg:default.paths.provisioning=/etc/grafana/provisioning
