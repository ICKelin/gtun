rm -r bin

mkdir -p bin/gtun
mkdir -p bin/gtun/log

mkdir -p bin/gtund
mkdir -p bin/gtund/log

GOOS=linux go build -o bin/gtund/gtund cmd/gtund/*.go
cp etc/gtund.conf bin/gtund/

cd cmd/gtun
echo "building gtun_cli_darwin...."
GOOS=darwin go build -o ../../bin/gtun/gtun-darwin_amd64 
echo "builded gtun_cli_darwin...."

echo "building gtun_cli_linux...."
GOOS=linux go build -o ../../bin/gtun/gtun-linux_amd64 
echo "builded gtun_cli_linux...."

echo "building gtun_cli_win...."

GOOS=windows go build -o ../../bin/gtun/gtun-windows_amd64.exe
echo "builded gtun_cli_win...."