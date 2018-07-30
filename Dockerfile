FROM golang:1.10.3-alpine go
FROM node:8.11.3-alpine node
#FROM kkarczmarczyk/node-yarn
FROM alpine:3.8
MAINTAINER wuzhijun

RUN apk add make git npm
#RUN npm install -g yarn

COPY . /source
WORKDIR /source
RUN make all

#RUN cp /source/bin/darwin-amd64/grafana-server /usr/bin/grafana-server
#RUN cp /source/bin/darwin-amd64/grafana-cli /usr/bin/grafana-cli

#COPY /source/conf/grafana.ini /etc/grafana/
#COPY /source/conf/ldap.toml /etc/grafana/
#COPY /source/conf/provisioning/ /etc/grafana/provisioning/

#RUN chmod 777 -R /etc/grafana/
#RUN chmod 777 -R /var/lib/grafana
#RUN chmod 777 -R /usr/share/grafana

#RUN chmod a+x /usr/bin/grafana-server
#ENTRYPOINT ["/usr/bin/grafana-server"]
