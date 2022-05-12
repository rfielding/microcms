FROM ubuntu:21.10
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
# UGH! dealing with imagemagick bug
RUN mv /etc/ImageMagick-6/policy.xml /etc/ImageMagick-6/policy.xml.bak
RUN cat /etc/ImageMagick-6/policy.xml.bak | grep -v PDF > /etc/ImageMagick-6/policy.xml
RUN pwd
RUN cd / && wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz
RUN cd / ; tar zxf /go1.18.1.linux-amd64.tar.gz
RUN ln -s /go/bin/go /usr/local/bin/go
RUN go version
COPY . /root
RUN cd /root && go build -tags fts5 -o ./gosqlite main.go
RUN cd /root && rm schema.db ; sqlite3 schema.db < schema.sql
RUN cd /root ; mkdir files || true
# writable volume mount... make sure we have permissions to write it and for host to delete contents
RUN chown -R 1000:1000 /root
RUN chmod -R 755 /root/files
RUN chmod 755 /root/bin/tika-server-standard.jar
WORKDIR /root
USER 1000:1000
CMD mkdir /root/files/images && cp /root/media/*100.png /root/files/images && ./gosqlite & java -jar ./bin/tika-server-standard.jar
