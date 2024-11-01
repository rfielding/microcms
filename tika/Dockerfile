FROM ubuntu:24.10
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York
RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y default-jre
WORKDIR /root
RUN wget https://dlcdn.apache.org/tika/2.9.2/tika-server-standard-2.9.2.jar
RUN mv tika-server-standard-2.9.2.jar tika.jar
RUN chmod 755 /root/tika.jar
USER 1000:1000
CMD java -jar /root/tika.jar --host 0.0.0.0 --port 9998
