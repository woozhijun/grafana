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
RUN apk add ca-certificates
COPY . /source
WORKDIR /source

ENV GF_PATHS_CONFIG="/etc/grafana" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning" 

RUN mkdir -p "$GF_PATHS_CONFIG" "$GF_PATHS_DATA" \
             "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_PROVISIONING" "$GF_PATHS_HOME"

COPY --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-server \
     --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-cli \ 
     /usr/sbin/
COPY --from=node-onbuild /source/public \
     --from=node-onbuild /source/scripts \
     --from=node-onbuild /source/devenv \
     --from=node-onbuild /source/vendor \
     --from=node-onbuild /source/conf \
     "$GF_PATHS_HOME"/

RUN cp ./conf/grafana.ini "$GF_PATHS_CONFIG" && \
    cp ./conf/defaults.ini "$GF_PATHS_CONFIG" && \
    cp ./conf/ldap.toml "$GF_PATHS_CONFIG" && \
    cp -r ./conf/provisioning "$GF_PATHS_PROVISIONING" 

RUN chmod a+x /usr/sbin/grafana-server && chmod a+x /usr/sbin/grafana-cli && \
    chmod 777 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_HOME"

#
ARG GF_INSTALL_PLUGINS="grafana-piechart-panel,mtanda-histogram-panel,alexanderzobnin-zabbix-app,grafana-kairosdb-datasource,abhisant-druid-datasource,michaeldmoore-annunciator-panel,digiapulssi-breadcrumb-panel,btplc-trend-box-panel,natel-discrete-panel,vonage-status-panel,btplc-status-dot-panel,grafana-clock-panel,grafana-simple-json-datasource"

ENV GF_EXEC_PROD="production"
RUN if [ ! -z "${GF_INSTALL_PLUGINS}" ]; then \
    OLDIFS=$IFS; \
        IFS=','; \
    for plugin in ${GF_INSTALL_PLUGINS}; do \
        IFS=$OLDIFS; \
        /usr/sbin/grafana-cli --pluginsDir "${GF_PATHS_PLUGINS}" plugins install ${plugin}; \
    done; \
    fi

COPY ./start.sh /start.sh
RUN chmod +x /start.sh
WORKDIR /
ENTRYPOINT ["/start.sh"]
