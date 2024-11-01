FROM ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York
ARG UID=1000
ARG GID=1000
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y gcc
RUN apt-get install -y time
RUN apt-get install -y npm
RUN apt-get install -y golang

COPY cmd/rproxy/main.go /root/cmd/rproxy/main.go
WORKDIR /root/cmd/rproxy

RUN cd /root/cmd/rproxy && go mod tidy
RUN cd /root/cmd/rproxy && go build -o ./rproxy *.go
WORKDIR /root
RUN chown -R ${UID}:${GID} .
USER ${UID}:${GID}
CMD ( ./cmd/rproxy/rproxy & ); sleep 2;( cd ui ; npm install --force; npm run start )
