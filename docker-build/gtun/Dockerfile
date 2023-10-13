FROM ubuntu:18.04
COPY gtun-linux_amd64 /gtun
COPY start.sh /
RUN chmod +x start.sh && chmod +x gtun
RUN mkdir /opt/logs
CMD /start.sh