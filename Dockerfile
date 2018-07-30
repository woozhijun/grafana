FROM golang:1.10.3-alpine3.8 AS go-onbuild
RUN apk add build-base git make

COPY . /go/src/github.com/grafana/grafana
WORKDIR /go/src/github.com/grafana/grafana

RUN make all-go

FROM node:8.11-alpine AS node-onbuild
RUN apk add --update build-base git make
RUN npm install -g yarn

COPY . /source
WORKDIR /source
RUN make all-js

#
FROM alpine:3.8
MAINTAINER wuzhijun
COPY . /source

#
COPY --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-server /usr/bin/grafana-server
COPY --from=go-onbuild /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-cli /usr/bin/grafana-cli
COPY --from=node-onbuild /source/public /usr/share/grafana/public

RUN mkdir -p /etc/grafana/
RUN cp /source/conf/grafana.ini /etc/grafana/
RUN cp /source/conf/ldap.toml /etc/grafana/
RUN cp -r /source/conf/provisioning /etc/grafana/provisioning

RUN chmod 777 -R /etc/grafana/
RUN chmod 777 -R /var/lib/grafana
RUN chmod 777 -R /usr/share/grafana

RUN chmod a+x /usr/bin/grafana-server
ENTRYPOINT ["/usr/bin/grafana-server"]
