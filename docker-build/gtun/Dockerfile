FROM ubuntu:20.04
RUN mkdir -p /opt/apps/gtun/logs
COPY . /opt/apps/gtun
COPY start.sh /
RUN chmod +x start.sh && chmod +x /opt/apps/gtun/gtun
CMD /start.sh