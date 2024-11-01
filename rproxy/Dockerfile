FROM ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York

RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y gcc
RUN apt-get install -y time
RUN apt-get install -y npm
RUN apt-get install -y golang

COPY . /root
WORKDIR /root/cmd/rproxy

RUN cd /root/cmd/rproxy && go build -o ./rproxy *.go
WORKDIR /root
RUN chown -R 1000:1000 .
USER 1000:1000
CMD ( ./cmd/rproxy/rproxy & ); sleep 2;( cd ui ; npm install --force; npm run start )
