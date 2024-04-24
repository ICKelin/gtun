setname=GTUN_ALL
noproxy_set=NOPROXY
redirect_port=8524

clear_noproxy() {
    iptables -t mangle -D PREROUTING -m set --match-set $noproxy_set dst -j ACCEPT
    iptables -t mangle -D OUTPUT -m set --match-set $noproxy_set dst -j ACCEPT
    ipset destroy $noproxy_set >/dev/null
}

add_noproxy() {
  ipset create $noproxy_set hash:net
  cat noproxy.txt | while read line
  do
      echo "no proxy for " $line
      ipset add $noproxy_set $line
  done

  iptables -t mangle -A PREROUTING -m set --match-set $noproxy_set dst -j ACCEPT
  iptables -t mangle -A OUTPUT -m set --match-set $noproxy_set dst -j ACCEPT
}

clear_proxy() {
    ip ro del local default dev lo table 100 >/dev/null
    ip rule del fwmark 1 lookup 100 >/dev/null
    iptables -t mangle -D PREROUTING -p tcp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port $redirect_port
    iptables -t mangle -D PREROUTING -p udp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port $redirect_port
    iptables -t mangle -D OUTPUT -m set --match-set $setname dst -j MARK --set-mark 1 >/dev/null
    ipset destroy $setname >/dev/null
}

add_proxy() {
  ipset create $setname hash:net
  echo "proxy 0.0.0.0/1"
  echo "proxy 128.0.0.0/1"
  ipset add $setname 0.0.0.0/1
  ipset add $setname 128.0.0.0/1

  iptables -t mangle -A PREROUTING -p tcp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port $redirect_port
  iptables -t mangle -A PREROUTING -p udp -m set --match-set $setname dst -j TPROXY --tproxy-mark 1/1 --on-port $redirect_port
  iptables -t mangle -A OUTPUT -m set --match-set $setname dst -j MARK --set-mark 1

  # redirect dns query
#  iptables -t mangle -A PREROUTING -p udp --dport 53 -j TPROXY --tproxy-mark 1/1 --on-port $redirect_port
#  iptables -t mangle -A OUTPUT -p udp --dport 53 -j MARK --set-mark 1

  ip rule add fwmark 1 lookup 100
  ip ro add local default dev lo table 100
}

sep="============================================"

echo "gtun accelerator proxy all traffic"
clear_noproxy
clear_proxy

echo "Adding noproxy cidrs"
add_noproxy
echo "Done."
echo $sep

echo "Adding proxy cidrs"
add_proxy
echo "Done."
echo $sep