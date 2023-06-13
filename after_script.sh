iptables -t mangle -D PREROUTING -p tcp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -D PREROUTING -p udp -m set --match-set GTUN-US dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -D OUTPUT -m set --match-set GTUN-US dst -j MARK --set-mark 1
ip rule del fwmark 1 lookup 100
ip ro del local default dev lo table 100
ipset destroy GTUN-US