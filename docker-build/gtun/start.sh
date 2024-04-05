#!/usr/bin/env bash
if [ "$TIME_ZONE" != "" ]; then
    ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone
fi

/opt/apps/gtun/gtun -c /opt/apps/gtun/etc/gtun.yaml