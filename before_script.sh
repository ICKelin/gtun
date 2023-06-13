ipset create GTUN-US hash:net
iptables -t mangle -I PREROUTING -p tcp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -I PREROUTING -p udp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -I OUTPUT -m set --match-set GTUN-US dst -j MARK --set-mark 1
ip rule add fwmark 1 lookup 100
ip ro add local default dev lo table 100