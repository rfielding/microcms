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
RUN apt-get install -y npm
RUN apt-get install -y golang

SHELL ["/bin/bash", "--login", "-i", "-c"]
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.35.2/install.sh | bash
RUN source /root/.bashrc && nvm install 14.21.2
SHELL ["/bin/bash", "--login", "-c"]

# UGH! dealing with imagemagick bug
RUN mv /etc/ImageMagick-6/policy.xml /etc/ImageMagick-6/policy.xml.bak
RUN cat /etc/ImageMagick-6/policy.xml.bak | grep -v PDF > /etc/ImageMagick-6/policy.xml
RUN pwd
COPY . /root
# You are here after each code change - it is so very slow because of cgo, because of sqlite
RUN cd /root/cmd/microcms && CGO_ENABLED=1 go build -tags fts5 -o ./microcms *.go
# writable volume mount... make sure we have permissions to write it and for host to delete contents
#RUN chown -R 1000:1000 /root/persistent # just the persistent dir is written
WORKDIR /root
USER 1000:1000
#RUN cd react/init/ui; npm install --force; npm run build; cd build 
CMD ./bin/containerinit
