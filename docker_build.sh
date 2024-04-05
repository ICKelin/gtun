./build.sh

rm -r images

mkdir -p images/gtun
cp -r release/gtun/* images/gtun/
cp -r docker-build/gtun/* images/gtun/

mkdir -p images/gtund
cp -r release/gtund/* images/gtund
cp -r docker-build/gtund/* images/gtund/