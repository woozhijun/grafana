#!/bin/bash

docker rm -f grafana
docker volume create grafana-storage

docker run -d \
  -p 3000:3000 \
  --name grafana \
  -v grafana-storage:/var/lib/grafana \
  -e "GF_SERVER_ROOT_URL=http://grafana.monitor.io" \
  -e "GF_EXEC_PROD=production" \
  docker.mobike.io/apm/grafana:`git rev-parse --short=8 HEAD` \
