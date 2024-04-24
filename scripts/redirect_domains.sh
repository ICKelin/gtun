
config_dnsmasq() {
    echo "configuring dnsmasq service"
    cp dnsmasq/dnsmasq.conf /etc/dnsmasq.conf
    cp dnsmasq/dnsmasq.resolv /etc/dnsmasq.resolv
    echo "configuring dnsmasq cn domain list"
    cp dnsmasq/cn.conf /etc/dnsmasq.d/
    systemctl restart dnsmasq
}

./redirect_all.sh
echo "Configuring dnsmasq"
config_dnsmasq
echo "Done."
echo $sep

