FROM ubuntu:20.04
RUN mkdir -p /opt/apps/gtund/logs
COPY . /opt/apps/gtund
COPY start.sh /
RUN chmod +x start.sh && chmod +x /opt/apps/gtund/gtund
CMD /start.sh