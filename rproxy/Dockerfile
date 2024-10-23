FROM --platform=linux/amd64 ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York

RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y gcc
RUN apt-get install -y time
RUN apt-get install -y npm

# UGH! dealing with imagemagick bug
RUN cd / && wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
RUN cd / ; tar zxf /go1.22.5.linux-amd64.tar.gz
RUN ln -s /go/bin/go /usr/local/bin/go
COPY . /root
WORKDIR /root/cmd/rproxy
RUN chown -R 1000:1000 .
RUN cd /root/cmd/rproxy && go build -o ./rproxy *.go
WORKDIR /root
USER 1000:1000
CMD ( ./cmd/rproxy/rproxy & ); sleep 2;( cd ui ; npm install --force; npm run start )
