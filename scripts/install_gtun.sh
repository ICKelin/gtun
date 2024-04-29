systemctl stop gtun
GTUN_DIR="/opt/apps/gtun"
mkdir -p $GTUN_DIR/logs
cp -r . $GTUN_DIR
cp etc/gtun.service /lib/systemd/system/
systemctl daemon-reload
systemctl start gtun