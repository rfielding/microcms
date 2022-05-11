FROM ubuntu:latest
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y sqlite3
RUN apt-get install -y git
RUN apt-get install -y gcc
RUN apt-get install -y default-jre
RUN apt-get install -y imagemagick
# UGH!
RUN mv /etc/ImageMagick-6/policy.xml /etc/ImageMagick-6/policy.xml.bak
RUN cat /etc/ImageMagick-6/policy.xml.bak | grep -v PDF > /etc/ImageMagick-6/policy.xml
RUN pwd
RUN cd / && wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz
RUN cd / ; tar zxf /go1.18.1.linux-amd64.tar.gz
RUN ln -s /go/bin/go /usr/local/bin/go
RUN go version
COPY . /root
RUN cd /root && ./build
WORKDIR /root
CMD ls -al & ./gosqlite & java -jar ./bin/tika-server-standard.jar
