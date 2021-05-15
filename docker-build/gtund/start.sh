#!/usr/bin/env bash
if [ "$TIME_ZONE" != "" ]; then
    ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone
fi

#项目的配置文件
if [ "$settings" != "" ]; then
    echo "$settings" > /gtund.yaml
fi

/gtund -c /gtund.yaml
