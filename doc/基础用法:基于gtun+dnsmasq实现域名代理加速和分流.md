# 基于gtun+dnsmasq实现域名代理加速和分流.md

之前的文章分享了[使用gtun实现ip代理加速和分流](最佳实践:基于gtun+ipset实现ip代理加速和分流.md)，实现了一个最简单的基于ip代理加速的场景，但是在实际应用当中会有两个不太优雅的地方：

- 基于ip的方式，如果ip发生变动，可能会出现分流策略不准的问题
- 有时候并不需要加速这么多ip，只需要加速部分网站或者应用即可（非常典型的比如SaaS应用加速）

基于此我们有了基于dnsmasq的域名解析策略来实现基于域名的加速和分流，最终拓扑如下：

![img.png](assets/img.png)

# 前置准备
您可以参考这篇文章来安装gtund和gtun。安装完gtun和gtund之后，您需要再安装dnsmasq并且成功启动。

# 配置dnsmasq解析策略

首先还是创建好基本的运行环境，参考scripts/redirect_all.sh

```shell

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
      echo "no proxy for" $line
      ipset add $noproxy_set $line
  done

  iptables -t mangle -A PREROUTING -m set --match-set $noproxy_set dst -j ACCEPT
  iptables -t mangle -A OUTPUT -m set --match-set $noproxy_set dst -j ACCEPT
}

```

然后配置dnsmasq的规则，dnsmasq的规则主要有两个：
- 域名解析的上游地址是多少
- 解析接入写入到哪个ipset里面
通过这两个控制我们就能够实现dnsmasq和gtun的无缝结合。

在本例当中，还是配置的代理所有ip，其中大陆地区的域名使用`114.114.114.114`这个上游地址进行解析，并将解析结果加入到NOPROXY这个ipset当中。
从而实现大陆地区的域名不加速。

```shell
config_dnsmasq() {
    echo "configuring dnsmasq service"
    cp dnsmasq/dnsmasq.conf /etc/dnsmasq.conf
    cp dnsmasq/dnsmasq.resolv /etc/dnsmasq.resolv
    echo "configuring dnsmasq cn domain list"
    cp dnsmasq/cn.conf /etc/dnsmasq.d/
    cp dnsmasq/cn_set.conf /etc/dnsmasq.d/
    systemctl restart dnsmasq
}

```

完整命令可以参考`gtun/scripts/redirect_domains.sh`。修改完之后本机需要设置`/etc/resolv.conf`文件的`nameserver 127.0.0.1`
才会真正用本机的dnsmasq去解析。

# 测试
接下来进行一轮测试，我们使用我们自己的一个域名`demo.xxxx.tech`进行测试。

第一步将demo.xxxx.tech配置进dnsmasq里面

```shell
root@OpenWrt:~/gtun# head /etc/dnsmasq.d/cn.conf
server=/demo.xxxx.tech/114.114.114.114

root@OpenWrt:~/gtun# head /etc/dnsmasq.d/cn_set.conf
ipset=/demo.xxxx.tech/NOPROXY
```

第二步nslookup解析测试

```shell
root@OpenWrt:~/gtun# nslookup demo.xxxx.tech 127.0.0.1
Server:		127.0.0.1
Address:	127.0.0.1:53

Non-authoritative answer:
Name:	demo.xxxx.tech
Address: 47.115.xx.xx

Non-authoritative answer:

root@OpenWrt:~/gtun# ipset -T NOPROXY 47.115.xx.xx
Warning: 47.115.xx.xx is in set NOPROXY.
```

demo.xxxx.tech这个域名已经被加入到NOPROXY里面了，根据之前的文章，加入到NOPROXY之后不会再走加速出口出，这里不再赘述了。

# 结束语
本文粗略的讲解了如何结合dnsmasq实现域名加速，截止目前位置我们已经了解到了ip加速和域名加速，但是所有的加速都是加速本机的流量，接下来我会结合
软路由的方式，详细说明如何实现连接Wi-Fi就能实现gtun的加速。
