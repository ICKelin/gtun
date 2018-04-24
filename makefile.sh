rm -r bin
mkdir -p bin/gtun_cli
mkdir -p bin/gtun_cli/log

mkdir -p bin/gtun_srv
mkdir -p bin/gtun_srv/log


GOOS=linux go build -o bin/gtun_srv/gtun_srv gtun_srv/*.go

cd gtun

echo "building gtun_cli_darwin...."
GOOS=darwin go build -o ../bin/gtun_cli/gtun_cli_darwin 
echo "builded gtun_cli_darwin...."

echo "building gtun_cli_linux...."
GOOS=linux go build -o ../bin/gtun_cli/gtun_cli_linux 
echo "builded gtun_cli_linux...."

echo "building gtun_cli_win...."
GOOS=windows go build -o ../bin/gtun_cli/gtun_cli_win.exe 
echo "builded gtun_cli_win...."