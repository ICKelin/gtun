echo "add no proxy address"
ipset create GTUN-NOPROXY hash:net
iptables -t mangle -I PREROUTING -m set --match-set GTUN-NOPROXY dst -j ACCEPT
iptables -t mangle -I OUTPUT -m set --match-set GTUN-NOPROXY dst -j ACCEPT

echo "start gtun"
nohup ./gtun-linux-amd64 -c gtun.yaml &

echo "start success."