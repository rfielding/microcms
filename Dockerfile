FROM --platform=linux/amd64 ubuntu:22.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y sqlite3
RUN apt-get install -y git
RUN apt-get install -y gcc
RUN apt-get install -y default-jre
RUN apt-get install -y ffmpeg
RUN apt-get install -y imagemagick
RUN apt-get install -y time
# UGH! dealing with imagemagick bug
RUN mv /etc/ImageMagick-6/policy.xml /etc/ImageMagick-6/policy.xml.bak
RUN cat /etc/ImageMagick-6/policy.xml.bak | grep -v PDF > /etc/ImageMagick-6/policy.xml
RUN pwd
RUN cd / && wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
RUN cd / ; tar zxf /go1.20.6.linux-amd64.tar.gz
RUN ln -s /go/bin/go /usr/local/bin/go
COPY . /root
RUN cd /root ; mkdir -p persistent/files || true
RUN cd /root && sqlite3 persistent/schema.db < schema.sql
# You are here after each code change - it is so very slow because of cgo, because of sqlite
RUN cd ;root/cmd/gosqlite ; go mod tidy
RUN cd /root/cmd/gosqlite && GOOS=linux GOARCH=amd64 go build -tags fts5 -o ./gosqlite *.go
# writable volume mount... make sure we have permissions to write it and for host to delete contents
RUN chown -R 1000:1000 /root
RUN chmod -R 755 /root/persistent
RUN chmod 755 /root/bin/tika-server-standard.jar
WORKDIR /root
USER 1000:1000
CMD ./cmd/gosqlite/gosqlite & java -jar ./bin/tika-server-standard.jar
