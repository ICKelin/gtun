#!/bin/bash

echo "building gtund...."
GOOS=linux go build -o bin/gtund/gtund cmd/gtund/*.go
echo "builded gtund...."

cd cmd/gtun
echo "building gtun...."
GOOS=linux go build -o ../../bin/gtun/gtun-linux_amd64 
GOARCH=arm GOOS=linux go build -o ../../bin/gtun/gtun-linux_arm 
echo "builded gtun...."
