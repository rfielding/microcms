FROM ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York

RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y gcc
RUN apt-get install -y time

# UGH! dealing with imagemagick bug
RUN cd / && wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
RUN cd / ; tar zxf /go1.20.6.linux-amd64.tar.gz
RUN ln -s /go/bin/go /usr/local/bin/go
COPY . /root
# You are here after each code change - it is so very slow because of cgo, because of sqlite
RUN cd /root/cmd/rproxy && go build -o ./rproxy *.go
# writable volume mount... make sure we have permissions to write it and for host to delete contents
RUN chown -R 1000:1000 /root
WORKDIR /root
USER 1000:1000
CMD ["./cmd/rproxy/rproxy"]
