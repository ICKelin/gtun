FROM ubuntu:18.04
COPY gtund /
COPY gtund.yaml /
COPY start.sh /
RUN chmod +x start.sh && chmod +x gtund
RUN mkdir /opt/logs
CMD /start.sh