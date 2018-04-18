mkdir -p bin/gtun_cli
mkdir -p bin/gtun_cli/log

mkdir -p bin/gtun_srv
mkdir -p bin/gtun_srv/log

go build -o bin/gtun_cli/gtun_cli gtun/*.go
go build -o bin/gtun_srv/gtun_srv gtun_srv/*.go
