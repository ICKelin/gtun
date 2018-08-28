rm -r bin
mkdir -p bin/gtun
mkdir -p bin/gtun/log

mkdir -p bin/gtund
mkdir -p bin/gtund/log


GOOS=linux go build -o bin/gtund/gtund main/gtund/*.go

cd main/gtun

echo "building gtun_cli_darwin...."
GOOS=darwin go build -o ../bin/gtun_cli/gtun_cli_darwin 
echo "builded gtun_cli_darwin...."

echo "building gtun_cli_linux...."
GOOS=linux go build -o ../bin/gtun_cli/gtun_cli_linux 
echo "builded gtun_cli_linux...."

echo "building gtun_cli_win...."
GOOS=windows go build -o ../bin/gtun_cli/gtun_cli_win.exe 
echo "builded gtun_cli_win...."