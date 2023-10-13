WORKSPACE=`pwd`
./build_exec.sh

echo "building gtund docker image"
cd docker-build/gtund
cp $WORKSPACE/bin/gtund/gtund .
docker build . -t gtund
echo "builded gtund docker image"

echo "building gtun docker image"
cd $WORKSPACE/docker-build/gtun
cp $WORKSPACE/bin/gtun/gtun-linux_amd64 .
docker build . -t gtun
echo "builded gtun docker image"