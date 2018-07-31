FROM golang:1.10.3-alpine3.8 AS go-onbuild
RUN apk add build-base git make

COPY . /go/src/github.com/grafana/grafana
WORKDIR /go/src/github.com/grafana/grafana

RUN make all-go

#
FROM node:8.11-alpine AS node-onbuild
RUN apk add --update build-base git make python
RUN npm install -g yarn
RUN PHANTOMJS_CDNURL=https://npm.taobao.org/mirrors/phantomjs/ npm install phantomjs-prebuilt

COPY . /source
WORKDIR /source
RUN make all-js

#
FROM alpine:3.8
MAINTAINER wuzhijun
COPY . /source

ENV GF_PATHS_CONFIG="/etc/grafana" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning" 

COPY --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-server /usr/bin/grafana-server
COPY --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-cli /usr/bin/grafana-cli
COPY --from=node-onbuild /source/public /usr/share/grafana/public

RUN mkdir -p "$GF_PATHS_CONFIG" "$GF_PATHS_DATA" \
	     "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" \
	     "$GF_PATHS_PROVISIONING"
RUN cp /source/conf/grafana.ini "$GF_PATHS_CONFIG" && \
    cp /source/conf/defaults.ini "$GF_PATHS_CONFIG" && \
    cp /source/conf/ldap.toml "$GF_PATHS_CONFIG" && \
    cp -r /source/conf/provisioning "$GF_PATHS_PROVISIONING" 

RUN chmod a+x /usr/bin/grafana-server && chmod a+x /usr/bin/grafana-cli

#
ARG GF_INSTALL_PLUGINS="grafana-piechart-panel,mtanda-histogram-panel,alexanderzobnin-zabbix-app,grafana-kairosdb-datasource,abhisant-druid-datasource,michaeldmoore-annunciator-panel,digiapulssi-breadcrumb-panel,btplc-trend-box-panel,natel-discrete-panel,vonage-status-panel,btplc-status-dot-panel,grafana-clock-panel,grafana-simple-json-datasource"

ENV GF_EXEC_PROD=""
RUN if [ ! -z "${GF_INSTALL_PLUGINS}" ]; then \
    OLDIFS=$IFS; \
        IFS=','; \
    for plugin in ${GF_INSTALL_PLUGINS}; do \
        IFS=$OLDIFS; \
        /usr/bin/grafana-cli --pluginsDir "${GF_PATHS_PLUGINS}" plugins install ${plugin}; \
    done; \
    fi

ENTRYPOINT ["/start.sh", "$GF_EXEC_PROD"]
