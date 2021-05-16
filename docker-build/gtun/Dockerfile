FROM ubuntu:18.04
COPY gtun-linux_amd64 /gtun
COPY gtun.yaml /
COPY start.sh /
RUN chmod +x start.sh && chmod +x gtund
RUN mkdir /opt/logs
CMD /start.sh
