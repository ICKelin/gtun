#!/bin/bash

echo "building gtund...."
GOOS=linux go build -o bin/gtund/gtund cmd/gtund/*.go
echo "builded gtund...."

echo "building gtun...."
GOOS=linux go build -o bin/gtun/gtun-linux_amd64 cmd/gtun/*.go
GOARCH=arm GOOS=linux go build -o bin/gtun/gtun-linux_arm  cmd/gtun/*.go
echo "builded gtun...."

cp -r etc/gtun.yaml bin/gtun/
cp -r etc/gtund.yaml bin/gtund/
cp install.sh bin/gtun/