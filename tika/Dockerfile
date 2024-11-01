FROM ubuntu:24.10
ARG UID=1000
ARG GID=1000
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York
RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y default-jre
WORKDIR /root
RUN wget https://dlcdn.apache.org/tika/2.9.2/tika-server-standard-2.9.2.jar
RUN mv tika-server-standard-2.9.2.jar tika.jar
RUN chmod 755 /root/tika.jar
RUN chown -R ${UID}:${GID} .
USER ${UID}:${GID}
CMD java -jar /root/tika.jar --host 0.0.0.0 --port 9998
