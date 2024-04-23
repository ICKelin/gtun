setname=GTUN_ALL

ipset create noproxy hash:net
# add no proxy file
iptables -t mangle -A PREROUTING -m set --match-set noproxy dst -j ACCEPT
iptables -t mangle -A OUTPUT -m set --match-set noproxy dst -j ACCEPT

ipset create $setname hash:net
ipset add 0.0.0.0/0 $setname

iptables -t mangle -A PREROUTING -p tcp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -A PREROUTING -p udp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port 8524
iptables -t mangle -A OUTPUT -m set --match-set $setname dst -j MARK --set-mark 1

ip rule add fwmark 1 lookup 100
ip ro add local default dev lo table 100