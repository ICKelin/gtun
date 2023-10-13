echo "add no proxy address"
iptables -t mangle -D PREROUTING -m set --match-set GTUN-NOPROXY dst -j ACCEPT
iptables -t mangle -D OUTPUT -m set --match-set GTUN-NOPROXY dst -j ACCEPT
ipset destroy GTUN-NOPROXY

echo "stop gtun"
killall gtun-linux_amd64