#!/bin/bash
WORKSPACE=`pwd`
./build_exec.sh

echo "building gtund docker image"
cd docker-build/gtund
cp $WORKSPACE/bin/gtund .
docker build . -t gtund
echo "builded gtund docker image"


# echo "building gtun docker image"
# cd $WORKSPACE/docker-build/gtun
# cp $WORKSPACE/bin/gtun .
# echo "builded gtun docker image"