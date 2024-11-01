FROM ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York

RUN apt-get update
RUN apt-get install -y coreutils
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y sqlite3
RUN apt-get install -y git
RUN apt-get install -y gcc
RUN apt-get install -y ffmpeg
RUN apt-get install -y imagemagick
RUN apt-get install -y time
RUN apt-get install -y golang

# UGH! dealing with imagemagick bug
RUN mv /etc/ImageMagick-6/policy.xml /etc/ImageMagick-6/policy.xml.bak
RUN cat /etc/ImageMagick-6/policy.xml.bak | grep -v PDF > /etc/ImageMagick-6/policy.xml
COPY . /root
RUN cd /root/cmd/microcms && CGO_ENABLED=1 go build -tags fts5 -o ./microcms *.go
WORKDIR /root
USER 1000:1000
CMD ["./bin/containerinit"]
