rm -r release
mkdir -p release
mkdir -p release/gtun/etc
mkdir -p release/gtund/etc

DIR=`pwd`

cd src/gtun
GOOS=linux go build -o gtun
mv gtun $DIR/release/gtun/
cd $DIR
cp scripts/install_gtun.sh release/gtun/install.sh
cp -r etc/gtun/* release/gtun/etc

cd src/gtund
GOOS=linux go build -o gtund
mv gtund $DIR/release/gtund/
cd $DIR
cp scripts/install_gtund.sh release/gtund/install.sh
cp -r etc/gtund/* release/gtund/etc