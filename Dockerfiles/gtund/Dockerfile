FROM ubuntu:18.04
RUN apt update -y && apt install -y iproute2 && apt install -y net-tools
COPY gtund /
COPY gtund.conf /
COPY start.sh /
RUN chmod +x start.sh && chmod +x gtund
RUN mkdir /logs
CMD /start.sh